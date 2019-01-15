package p2p

import (
	"errors"
)

//Both block and tx requests are handled asymmetricaly, using channels as inter-communication
//All the request in this file are specifically initiated by the miner package
func BlockReq(hash [32]byte) error {

	//p := peers.getRandomPeer(PEERTYPE_MINER)
	//logger.Printf("BLOCK_REQ for Block (%x) to miner (1 of %v) with IP-Port: %v", hash[0:8], peers.len(PEERTYPE_MINER), p.getIPPort())
	//
	//if p == nil {
	//	return errors.New("Couldn't get a connection, request not transmitted.")
	//}
	//
	//packet := BuildPacket(BLOCK_REQ, hash[:])
	//sendData(p, packet)

	//Try Block Request with Broadcast
	logger.Printf("BLOCK_REQ for Block (%x) to %v miners", hash[0:8], peers.len(PEERTYPE_MINER))
	for p := range peers.minerConns {
		//Write to the channel, which the peerBroadcast(*peer) running in a seperate goroutine consumes right away.

		if p == nil {
			return errors.New("Couldn't get a connection, request not transmitted.")
		}
		packet := BuildPacket(BLOCK_REQ, hash[:])
		sendData(p, packet)
	}

	return nil
}

func LastBlockReq() error {

	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(BLOCK_REQ, nil)
	sendData(p, packet)
	return nil
}

//Request specific transaction
func TxReq(hash [32]byte, reqType uint8) error {

	//p := peers.getRandomPeer(PEERTYPE_MINER)
	//
	//if reqType == FUNDSTX_REQ {
	//	logger.Printf("TX_REQ: %x Number of miners: %v, selected: %v", hash, peers.len(PEERTYPE_MINER), p.getIPPort())
	//}

	//if p == nil {
	//	return errors.New("Couldn't get a connection, request not transmitted.")
	//}

	//packet := BuildPacket(reqType, hash[:])
	//sendData(p, packet)

	//Try TXRequest broadcast
	logger.Printf("TX_REQ for TX (%x) to %v miners", hash, peers.len(PEERTYPE_MINER))
	for p := range peers.minerConns {
		//Write to the channel, which the peerBroadcast(*peer) running in a seperate goroutine consumes right away.

		if p == nil {
			return errors.New("Couldn't get a connection, request not transmitted.")
		}
		packet := BuildPacket(reqType, hash[:])
		sendData(p, packet)
	}

	return nil
}
