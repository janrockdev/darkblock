# DarkBlock - Metadata Blockchain

[![Go](https://github.com/janrockdev/darkblock/actions/workflows/go.yml/badge.svg)](https://github.com/janrockdev/darkblock/actions/workflows/go.yml)

(in progress)

language	files	total
Go	        24  	4,143

## Description
This project aims to create a custom blockchain tailored for metadata use cases. The focus is on building a flexible and modular architecture that allows for easy customization and integration with various network protocols. The blockchain will support efficient metadata storage, retrieval, and management, making it ideal for applications that require robust and fast metadata tracing proof.

## Modules
- crypto/keys (ED25519)
ED25519 is a public-key signature system designed to provide high performance while maintaining robust security. It's part of the EdDSA (Edwards-curve Digital Signature Algorithm) family and is built upon the Twisted Edwards curve known as Curve25519.

## Flows
- create cryto with keys and address generators and signers
- create initial protobuffer type definition and compile 
- create block builder and hasher just for header
note: sign block = take privKey > take hash of block (32 bytes long) > sign the bytes with privKey > it will create 64 bytes long signature > verification = public key and hash of block
- transaction hashing, signing and verification (single hashing, basic for now) / UTXO prep
- gRPC networking setup
- gRPC interconnection with bootstrap setup
note: peer discovery process = run an initial set of bootstap nodes > run new node > add bootstap nodes to the connected peers list > connect to them and receive their lists of connected nodes > and so on... > eventually in few seconds we will have all nodes in the network (not necessary active > that is another service > something like periodic version exchange, handshake etc.) 
- peer discovery and bootstrap network
- extended blockchain data structure (in-memory blocks store)
- added broadcasting transaction to all nodes, store to mempool, simulation of building blocks, selected node to be validator
- added basic genesis block and small validation for blocks
- genesis block and first transaction
- storing UTXO for future validation
- UTXO funds check, additional validations
- node split to independent nodes with publishing client
- consolidated verification for transactions with Merkle tree

- (in progress/plan) complete transaction and block validation, persist blocks, storing more details, small refactoring, consensus, and testing/benchmarking

## TODO
- [ ] test replacement of grpc.Dial to grpc.NewClient (performance, params etc.)
- [ ] remove failed node from distribution list (no retry, because node need to synchronise with the latest state)
ERR node/node.go:168 > failed to broadcast transaction: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing: dial tcp :5000: connect: connection refused

- create a separated project to cover the distributed networking and discovery

## Install

### Protobuf compiler
```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

### UML generator
```shell
go install github.com/jfeliu007/goplantuml/cmd/goplantuml
~/go/bin/goplantuml -recursive . > diagram.puml
# https://www.plantuml.com/
```

### Badger
```shell
go get github.com/dgraph-io/badger/v4
```

### Block example
```log
block print: [
    header:{
        version:1
        height:1
        prevHash:"\xc9R\x9f\xb1ౣm\xc6iKTO\x06\xf0d\xd0\xcceى\xab5\x8d2\x04\x8c0\xef\xf2\x1e\xf7"
        rootHash:"\x8e[\xf7]v\xa6\xe9v\xe5\x9d\xf5\x89r\xf5\xdeB\xca9\x98\xf5\xfalO\xc6\xe4\xd8d\xe1m\x95\x1fT"
        timestamp:1733569183867933564
    }
    transactions:{
        version:1
        inputs:{
            prevTxHash:"1cb11d27e5d37b19cf3f54b9ede935677c462c25860539252ab911088a4f69c4"
            prevOutIndex:1
            publicKey:"\xe1\xf7\x0e<\xaa\xbep\x80+Of\xfa\xa3\xdd\xc6M\xaa+\xbb\xa5\x92H\x8be\xc4e\xa1\xc1\x16\xb3G\x8d"
            signature:"Օ${\xa7\xa9gBHTƁl\x9fa\x0e\x03\x1f\x13\xb4\xd2k\x9c\xefa=\x86\xa5\x12'\xa3\xadW\x87X\xf7\x9f\xabBC\xdb\xf3S)\xcb\xdd۟Y\xcbQ*\xabk\x13]\x9a\xcdo\xd5`\xe7\xc8\x0c"
        } outputs:{
            amount:1
            address:"\xd8k\xb5p\xe3W\xa9\x9f \xb7evw%\x0c\x91?\xcc\xf7u"
            payload:"{\"metadata\": \"sims_d4c08a06-7bd9-47b2-a39f-9dae3c0b220c\"}"
        }
    }
    transactions:{
        version:1
        inputs:{
            prevTxHash:"5881389f0b100136e9c19d3c6afa913cc24490700aa68d648d0170bcd0f865ec"
            prevOutIndex:2
            publicKey:"\xe1\xf7\x0e<\xaa\xbep\x80+Of\xfa\xa3\xdd\xc6M\xaa+\xbb\xa5\x92H\x8be\xc4e\xa1\xc1\x16\xb3G\x8d"
            signature:"\xa6\xa40fn\xd1\xd1\xdbמt\xa5\xdf[\xf3w\xefE\xa7\xe1\x1c\xce\xf21\x94\x94к\xe8\x9aؙV\xe8\xf8\xc2\xd0\xf6\xb3\xb8\xbfyT\x88\x08y\xbf\xf7;\x89\xab\xd8yH\xcf\xdbXcc\x04\x01\xed\xe3\x05"
        } 
        outputs:{
            amount:1
            address:"\xd8k\xb5p\xe3W\xa9\x9f \xb7evw%\x0c\x91?\xcc\xf7u"
            payload:"{\"metadata\": \"sims_4ff2cfa8-e6c0-4d5e-be0c-1dc39aaec74c\"}"
        }
    }
    publicKey:"\xe1\xf7\x0e<\xaa\xbep\x80+Of\xfa\xa3\xdd\xc6M\xaa+\xbb\xa5\x92H\x8be\xc4e\xa1\xc1\x16\xb3G\x8d"
    signature:"\xb9Y5\x92\x89'3|?6\xbdƒ\xd94f\x88\x91gz\x18\xf0\x90~C\xe7e\x02:543\xa6\x19@'\xdd_\x1b|\x92\xcf2(\x7f\xb1،\xe7I\x0e\x90\xbci\x96\x95\x19\x04C*BgV\x00"
]
```