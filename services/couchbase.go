package services

import (
	"encoding/hex"
	"fmt"

	"github.com/couchbase/gocb/v2"
	"github.com/janrockdev/darkblock/proto"
	"github.com/janrockdev/darkblock/types"
)

type CouchbaseService struct {
	cluster     *gocb.Cluster
	blockBucket *gocb.Bucket
	txBucket    *gocb.Bucket
	blockColl   *gocb.Collection
	txColl      *gocb.Collection
}

func NewCouchbaseService(connectionString, username, password, blockBucketName, txBucketName string) (*CouchbaseService, error) {
	cluster, err := gocb.Connect(connectionString, gocb.ClusterOptions{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Couchbase: %v", err)
	}

	blockBucket := cluster.Bucket(blockBucketName)
	txBucket := cluster.Bucket(txBucketName)

	blockColl := blockBucket.DefaultCollection()
	txColl := txBucket.DefaultCollection()

	return &CouchbaseService{
		cluster:     cluster,
		blockBucket: blockBucket,
		txBucket:    txBucket,
		blockColl:   blockColl,
		txColl:      txColl,
	}, nil
}

func (cs *CouchbaseService) StoreBlock(block *proto.Block) error {

	blockPrefix := fmt.Sprintf("%016d", block.Header.Height)
	blockHash := hex.EncodeToString(types.HashBlock(block))
	blockKey := fmt.Sprintf("%s_%s", blockPrefix, blockHash)
	blockID := fmt.Sprintf("block::%s", blockKey)

	_, err := cs.blockColl.Upsert(blockID, block, &gocb.UpsertOptions{})
	if err != nil {
		return fmt.Errorf("failed to store block: %v", err)
	}

	for _, tx := range block.Transactions {
		txID := fmt.Sprintf("tx::%s", hex.EncodeToString(types.HashTransactionNoSigPuK(tx)))
		_, err := cs.txColl.Upsert(txID, tx, &gocb.UpsertOptions{})
		if err != nil {
			return fmt.Errorf("failed to store transaction: %v", err)
		}
	}

	return nil
}

func (cs *CouchbaseService) SearchTransactionByPayload(payload []byte) (string, error) {
	// N1QL query to search for the transaction by its payload.
	query := fmt.Sprintf("SELECT META(tx).id, tx FROM `transactions` AS tx WHERE tx.outputs[0].payload = \"%s\"", payload)

	// Use the cluster to execute the N1QL query.
	rows, err := cs.cluster.Query(query, &gocb.QueryOptions{
		NamedParameters: map[string]interface{}{
			"$1": payload,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close() // Ensure to close the rows iterator when done

	// Check if there are any rows to process
	var transaction struct {
		ID string `json:"id"`
	}

	// Iterate over the result set
	for rows.Next() {
		// Get the next row of data
		if err := rows.Row(&transaction); err != nil {
			return "", fmt.Errorf("failed to retrieve transaction: %v", err)
		}

		return transaction.ID, nil
	}

	// Return an error if no transaction was found
	return "", fmt.Errorf("transaction not found with the given payload")
}

func (cs *CouchbaseService) SearchBlockByPayload(payload []byte) (string, error) {
	// N1QL query to search for the transaction by its payload.
	query := fmt.Sprintf("SELECT META(bx).id, bx, tx, output FROM `blocks` AS bx UNNEST bx.transactions AS tx UNNEST tx.outputs AS output WHERE output.payload = \"%s\"", payload)

	// Use the cluster to execute the N1QL query.
	rows, err := cs.cluster.Query(query, &gocb.QueryOptions{
		NamedParameters: map[string]interface{}{
			"$1": payload,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close() // Ensure to close the rows iterator when done

	// Check if there are any rows to process
	var transaction struct {
		ID string `json:"id"`
	}

	// Iterate over the result set
	for rows.Next() {
		// Get the next row of data
		if err := rows.Row(&transaction); err != nil {
			return "", fmt.Errorf("failed to retrieve transaction: %v", err)
		}

		return transaction.ID, nil
	}

	// Return an error if no transaction was found
	return "", fmt.Errorf("transaction not found with the given payload")
}

func (cs *CouchbaseService) Close() error {
	return cs.cluster.Close(nil)
}
