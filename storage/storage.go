package storage

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/boltdb/bolt"
	"log"
	"sync"
	"time"
)

var (
	db                 *bolt.DB
	logger             *log.Logger
	State              = make(map[[64]byte]*protocol.Account)
	//This map keeps track of the relative account adjustments within a shard, such as balance, txcount and stakingheight
	RelativeState                     = make(map[[64]byte]*protocol.RelativeAccount)
	RootKeys                          = make(map[[64]byte]*protocol.Account)
	txMemPool                         = make(map[[32]byte]protocol.Transaction)
	ReceivedStateStash                      = protocol.NewStateStash()
	OwnBlockStash           []*protocol.Block
	OwnStateTransitionStash []*protocol.StateTransition
	AllClosedBlocksAsc      []*protocol.Block
	BootstrapServer         string
	Buckets                 []string
	memPoolMutex                                  	   = &sync.Mutex{}
	ThisShardID             int // ID of the shard this validator is assigned to
	txINVALIDMemPool        = make(map[[32]byte]protocol.Transaction)
	ReceivedBlockStash      = make([]*protocol.Block, 0)
)

const (
	ERROR_MSG 				= "Storage initialization aborted. Reason: "
	OPENBLOCKS_BUCKET 		= "openblocks"
	CLOSEDBLOCKS_BUCKET 	= "closedblocks"
	CLOSEDFUNDS_BUCKET 		= "closedfunds"
	CLOSEDACCS_BUCKET 		= "closedaccs"
	CLOSEDSTAKES_BUCKET 	= "closedstakes"
	CLOSEDCONFIGS_BUCKET	= "closedconfigs"
	LASTCLOSEDBLOCK_BUCKET 	= "lastclosedblock"
	GENESIS_BUCKET			= "genesis"
	CLOSEDEPOCHBLOCK_BUCKET = "closedepochblocks"
	LASTCLOSEDEPOCHBLOCK_BUCKET = "lastclosedepochblocks"
	OPENEPOCHBLOCK_BUCKET	= "openepochblock"
)

//Entry function for the storage package
func Init(dbname string, bootstrapIpport string) error {
	BootstrapServer = bootstrapIpport
	logger = InitLogger()

	Buckets = []string {
		OPENBLOCKS_BUCKET,
		CLOSEDBLOCKS_BUCKET,
		CLOSEDFUNDS_BUCKET,
		CLOSEDACCS_BUCKET,
		CLOSEDSTAKES_BUCKET,
		CLOSEDCONFIGS_BUCKET,
		LASTCLOSEDBLOCK_BUCKET,
		GENESIS_BUCKET,
		CLOSEDEPOCHBLOCK_BUCKET,
		LASTCLOSEDEPOCHBLOCK_BUCKET,
		OPENEPOCHBLOCK_BUCKET,
	}

	var err error
	db, err = bolt.Open(dbname, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		logger.Fatal(ERROR_MSG, err)
		return err
	}
	for _, bucket := range Buckets {
		err = db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucket))
			if b == nil {
				return fmt.Errorf("Bucket not found")
			}
			return nil
		})

		if(err == nil){
			err = clearBucket(bucket)
			logger.Printf("Bucket cleared: %v", bucket)
			if err != nil {
				return err
			}
		} else {
			err = CreateBucket(bucket, db)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CreateBucket(bucketName string, db *bolt.DB) (err error) {
	return db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte(bucketName))
		if err != nil {
			return fmt.Errorf(ERROR_MSG + " %s", err)
		}
		return nil
	})
}

func DeleteBucket(bucketName string, db *bolt.DB) (err error) {
	return db.Update(func(tx *bolt.Tx) error {
		err = tx.DeleteBucket([]byte(bucketName))
		if err != nil {
			return fmt.Errorf(ERROR_MSG + " %s", err)
		}
		return nil
	})
}

func TearDown() {
	db.Close()
}
