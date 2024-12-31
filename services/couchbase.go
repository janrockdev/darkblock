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

func (cs *CouchbaseService) Close() error {
	return cs.cluster.Close(nil)
}
