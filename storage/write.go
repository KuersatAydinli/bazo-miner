package storage

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/boltdb/bolt"
)

func WriteOpenBlock(block *protocol.Block) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(OPENBLOCKS_BUCKET))
		return b.Put(block.Hash[:], block.Encode())
	})
}

func WriteOpenEpochBlock(epochBlock *protocol.EpochBlock) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(OPENEPOCHBLOCK_BUCKET))
		return b.Put(epochBlock.Hash[:], epochBlock.Encode())
	})
}

func WriteClosedBlock(block *protocol.Block) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLOSEDBLOCKS_BUCKET))
		return b.Put(block.Hash[:], block.Encode())
	})
}

func WriteClosedEpochBlock(epochBlock *protocol.EpochBlock) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLOSEDEPOCHBLOCK_BUCKET))
		return b.Put(epochBlock.Hash[:], epochBlock.Encode())
	})
}

func WriteFirstEpochBlock(epochBlock *protocol.EpochBlock) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLOSEDEPOCHBLOCK_BUCKET))
		return b.Put([]byte("firstepochblock"), epochBlock.Encode())
	})
}

func WriteINVALIDOpenTx(transaction protocol.Transaction) {
	txINVALIDMemPool[transaction.Hash()] = transaction
}

func WriteToReceivedStash(block *protocol.Block) {
	ReceivedBlockStash = append(ReceivedBlockStash, block)
	//When lenght of stash is > 50 --> Remove first added Block
	if len(ReceivedBlockStash) > 50 {
		ReceivedBlockStash = append(ReceivedBlockStash[:0], ReceivedBlockStash[1:]...)
	}

}

func BlockAlreadyInStash(slice []*protocol.Block, newBlockHash [32]byte) bool {
	for _, blockInStash := range slice {
		if blockInStash.Hash == newBlockHash {
			return true
		}
	}
	return false
}

func WriteLastClosedEpochBlock(epochBlock *protocol.EpochBlock) (err error) {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LASTCLOSEDEPOCHBLOCK_BUCKET))
		return b.Put(epochBlock.Hash[:], epochBlock.Encode())
	})
}

func WriteLastClosedBlock(block *protocol.Block) (err error) {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LASTCLOSEDBLOCK_BUCKET))
		return b.Put(block.Hash[:], block.Encode())
	})
}

//Changing the "tx" shortcut here and using "transaction" to distinguish between bolt's transactions
func WriteOpenTx(transaction protocol.Transaction) {
	memPoolMutex.Lock()
	defer memPoolMutex.Unlock()
	txMemPool[transaction.Hash()] = transaction
}

func WriteClosedTx(transaction protocol.Transaction) error {
	var bucket string
	switch transaction.(type) {
	case *protocol.FundsTx:
		bucket = CLOSEDFUNDS_BUCKET
	case *protocol.ContractTx:
		bucket = CLOSEDACCS_BUCKET
	case *protocol.ConfigTx:
		bucket = CLOSEDCONFIGS_BUCKET
	case *protocol.StakeTx:
		bucket = CLOSEDSTAKES_BUCKET
	}

	hash := transaction.Hash()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Put(hash[:], transaction.Encode())
	})
}

func WriteAccount(account *protocol.Account) {
	State[account.Address] = account
}

func WriteGenesis(genesis *protocol.Genesis) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(GENESIS_BUCKET))
		return b.Put([]byte("genesis"), genesis.Encode())
	})
}

func WriteToOwnBlockStash(block *protocol.Block) {
	OwnBlockStash = append(OwnBlockStash,block)

	if(len(OwnBlockStash) > 20){
		OwnBlockStash = append(OwnBlockStash[:0], OwnBlockStash[1:]...)
	}
}

func WriteToOwnStateTransitionkStash(st *protocol.StateTransition) {
	OwnStateTransitionStash = append(OwnStateTransitionStash,st)

	if(len(OwnStateTransitionStash) > 20){
		OwnStateTransitionStash = append(OwnStateTransitionStash[:0], OwnStateTransitionStash[1:]...)
	}
}

