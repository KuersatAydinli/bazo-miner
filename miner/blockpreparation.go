package miner

import (
	"encoding/binary"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"sort"
)

//The code here is needed if a new block is built. All open (not yet validated) transactions are first fetched
//from the mempool and then sorted. The sorting is important because if transactions are fetched from the mempool
//they're received in random order (because it's implemented as a map). However, if a user wants to issue more fundsTxs
//they need to be sorted according to increasing txCnt, this greatly increases throughput.

type openTxs []protocol.Transaction

func prepareBlock(block *protocol.Block) {
	//Fetch all txs from mempool (opentxs).
	opentxs := storage.ReadAllOpenTxs()

	//This copy is strange, but seems to be necessary to leverage the sort interface.
	//Shouldn't be too bad because no deep copy.
	var tmpCopy openTxs
	tmpCopy = opentxs

	sort.Sort(tmpCopy)

	for _, tx := range opentxs {
		/*When fetching and adding Txs from the MemPool, first check if it belongs to my shard. Only if so, then add tx to the block*/
		txAssignedShard := assignTransactionToShard(tx)

		if txAssignedShard == ValidatorShardMap[validatorAccAddress]{
			/*Set shard identifier in block*/
			block.ShardId = uint8(txAssignedShard)

			//Prevent block size to overflow.
			if block.GetSize()+tx.Size() > activeParameters.Block_size {
				break
			}

			err := addTx(block, tx)
			if err != nil {
				//If the tx is invalid, we remove it completely, prevents starvation in the mempool.
				storage.DeleteOpenTx(tx)
			}
		}
	}
}

func assignTransactionToShard(transaction protocol.Transaction) (shardNr int) {
	//Convert Address/Issuer ([64]bytes) included in TX to bigInt for the modulo operation to determine the assigned shard ID.
	var txSenderAddressInt uint64

	switch transaction.(type) {
		case *protocol.ContractTx:
			binary.BigEndian.PutUint64(transaction.(*protocol.ContractTx).Issuer[:], uint64(txSenderAddressInt))
			return int((int(txSenderAddressInt) % NumberOfShards) + 1)
		case *protocol.FundsTx:
			binary.BigEndian.PutUint64(transaction.(*protocol.FundsTx).From[:], uint64(txSenderAddressInt))
			return int((int(txSenderAddressInt) % NumberOfShards) + 1)
		case *protocol.ConfigTx:
			binary.BigEndian.PutUint64(transaction.(*protocol.ConfigTx).Sig[:], uint64(txSenderAddressInt))
			return int((int(txSenderAddressInt) % NumberOfShards) + 1)
		case *protocol.StakeTx:
			binary.BigEndian.PutUint64(transaction.(*protocol.StakeTx).Account[:], uint64(txSenderAddressInt))
			return int((int(txSenderAddressInt) % NumberOfShards) + 1)
		default:
			return 1
		}
}

//Implement the sort interface
func (f openTxs) Len() int {
	return len(f)
}

func (f openTxs) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f openTxs) Less(i, j int) bool {
	//Comparison only makes sense if both tx are fundsTxs.
	//Why can we only do that with switch, and not e.g., if tx.(type) == ..?
	switch f[i].(type) {
	case *protocol.ContractTx:
		//We only want to sort a subset of all transactions, namely all fundsTxs.
		//However, to successfully do that we have to place all other txs at the beginning.
		//The order between contractTxs and configTxs doesn't matter.
		return true
	case *protocol.ConfigTx:
		return true
	case *protocol.StakeTx:
		return true
	}

	switch f[j].(type) {
	case *protocol.ContractTx:
		return false
	case *protocol.ConfigTx:
		return false
	case *protocol.StakeTx:
		return false
	}

	return f[i].(*protocol.FundsTx).TxCnt < f[j].(*protocol.FundsTx).TxCnt
}
