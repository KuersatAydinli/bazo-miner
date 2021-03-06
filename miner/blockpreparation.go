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

	//Keep track of transactions from assigned for my shard and which are valid. Only consider these ones when filling a block
	//Otherwhise we would also count invalid transactions from my shard, this prevents well-filled blocks.
	txFromThisShard := 0

	for _, tx := range opentxs {
		/*When fetching and adding Txs from the MemPool, first check if it belongs to my shard. Only if so, then add tx to the block*/
		txAssignedShard := assignTransactionToShard(tx)

		if int(txAssignedShard) == ValidatorShardMap.ValMapping[validatorAccAddress]{
			FileLogger.Printf("---- Transaction (%x) in shard: %d\n", tx.Hash(),txAssignedShard)
			//Prevent block size to overflow.
			if int(block.GetSize()+10)+(txFromThisShard*int(len(tx.Hash()))) > int(activeParameters.Block_size){
				break
			}

			switch tx.(type) {
			case *protocol.StakeTx:
				//Add StakeTXs only when preparing the last block before the next epoch block
				if (int(lastBlock.Height) == int(lastEpochBlock.Height) + int(activeParameters.epoch_length) - 1) {
					err := addTx(block, tx)
					if err == nil {
						txFromThisShard += 1
					}
				}
			case *protocol.ContractTx, *protocol.FundsTx, *protocol.ConfigTx:
				err := addTx(block, tx)
				if err != nil {
					//If the tx is invalid, we remove it completely, prevents starvation in the mempool.
					//storage.DeleteOpenTx(tx)
					storage.WriteINVALIDOpenTx(tx)
					//storage.DeleteOpenTx(tx)
				} else {
					txFromThisShard += 1
				}
			}

		}
	}
}

/**
	Transactions are sharded based on the public address of the sender
 */
func assignTransactionToShard(transaction protocol.Transaction) (shardNr int) {
	//Convert Address/Issuer ([64]bytes) included in TX to bigInt for the modulo operation to determine the assigned shard ID.
	switch transaction.(type) {
		case *protocol.ContractTx:
			var byteToConvert [64]byte
			byteToConvert = transaction.(*protocol.ContractTx).Issuer
			var calculatedInt int
			calculatedInt = int(binary.BigEndian.Uint64(byteToConvert[:8]))
			return int((Abs(int32(calculatedInt)) % int32(NumberOfShards)) + 1)
		case *protocol.FundsTx:
			var byteToConvert [64]byte
			byteToConvert = transaction.(*protocol.FundsTx).From
			var calculatedInt int
			calculatedInt = int(binary.BigEndian.Uint64(byteToConvert[:8]))
			return int((Abs(int32(calculatedInt)) % int32(NumberOfShards)) + 1)
		case *protocol.ConfigTx:
			var byteToConvert [64]byte
			byteToConvert = transaction.(*protocol.ConfigTx).Sig
			var calculatedInt int
			calculatedInt = int(binary.BigEndian.Uint64(byteToConvert[:8]))
			return int((Abs(int32(calculatedInt)) % int32(NumberOfShards)) + 1)
		case *protocol.StakeTx:
			var byteToConvert [64]byte
			byteToConvert = transaction.(*protocol.StakeTx).Account
			var calculatedInt int
			calculatedInt = int(binary.BigEndian.Uint64(byteToConvert[:8]))
			return int((Abs(int32(calculatedInt)) % int32(NumberOfShards)) + 1)
		default:
			return 1 // default shard ID
		}
}

func Abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
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

/**
	During the synchronisation phase at every block height, the validator also receives the transaction hashes which were validated
	by the other shards. To avoid starvation, delete those transactions from the mempool
 */
func DeleteTransactionFromMempool(contractData [][32]byte, fundsData [][32]byte, configData [][32]byte, stakeData [][32]byte) {
	for _,fundsTX := range fundsData{
		if(storage.ReadOpenTx(fundsTX) != nil){
			storage.DeleteOpenTx(storage.ReadOpenTx(fundsTX))
			FileLogger.Printf("Deleted transaction (%x) from the MemPool.\n",fundsTX)
		}
	}

	for _,configTX := range configData{
		if(storage.ReadOpenTx(configTX) != nil){
			storage.DeleteOpenTx(storage.ReadOpenTx(configTX))
			FileLogger.Printf("Deleted transaction (%x) from the MemPool.\n",configTX)
		}
	}

	for _,stakeTX := range stakeData{
		if(storage.ReadOpenTx(stakeTX) != nil){
			storage.DeleteOpenTx(storage.ReadOpenTx(stakeTX))
			FileLogger.Printf("Deleted transaction (%x) from the MemPool.\n",stakeTX)
		}
	}

	for _,contractTX := range contractData{
		if(storage.ReadOpenTx(contractTX) != nil){
			storage.DeleteOpenTx(storage.ReadOpenTx(contractTX))
			FileLogger.Printf("Deleted transaction (%x) from the MemPool.\n",contractTX)
		}
	}

	//logger.Printf("Deleted transaction count: %d - New Mempool Size: %d\n",len(txPayload.FundsTxData)+len(txPayload.StakeTxData)+len(txPayload.ContractTxData)+ len(txPayload.ConfigTxData),storage.GetMemPoolSize())
	FileLogger.Printf("Deleted transaction count: %d - New Mempool Size: %d\n",len(contractData)+len(fundsData)+len(configData)+ len(stakeData),storage.GetMemPoolSize())
}
