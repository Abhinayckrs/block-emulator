package pbft_all

import (
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"encoding/json"
	"log"
	"strconv"
	"time"
)

type RawSLCPbftExtraHandleMod struct {
	pbftNode *PbftConsensusNode
}

func (slc *RawSLCPbftExtraHandleMod) HandleinPropose() (bool, *message.Request) {
	// Leader generates a block
	block := slc.pbftNode.CurChain.GenerateBlock(int32(slc.pbftNode.NodeID))
	r := &message.Request{
		RequestType: message.BlockRequest,
		ReqTime:     time.Now(),
	}
	r.Msg.Content = block.Encode()
	return true, r
}

func (slc *RawSLCPbftExtraHandleMod) HandleinPrePrepare(ppmsg *message.PrePrepare) bool {
	// Only leader validates the block
	if slc.pbftNode.NodeID == 0 { // Static leader
		if err := slc.pbftNode.CurChain.IsValidBlock(core.DecodeB(ppmsg.RequestMsg.Msg.Content)); err != nil {
			slc.pbftNode.pl.Plog.Printf("S%dN%d: invalid block\n", slc.pbftNode.ShardID, slc.pbftNode.NodeID)
			return false
		}
	}
	slc.pbftNode.pl.Plog.Printf("S%dN%d: accepted proposal\n", slc.pbftNode.ShardID, slc.pbftNode.NodeID)
	return true
}

func (slc *RawSLCPbftExtraHandleMod) HandleinPrepare(pmsg *message.Prepare) bool {
	// No prepare phase in SLC
	return true
}

func (slc *RawSLCPbftExtraHandleMod) HandleinCommit(cmsg *message.Commit) bool {
	r := slc.pbftNode.requestPool[string(cmsg.Digest)]
	block := core.DecodeB(r.Msg.Content)
	slc.pbftNode.pl.Plog.Printf("S%dN%d: adding block %d\n", slc.pbftNode.ShardID, slc.pbftNode.NodeID, block.Header.Number)
	slc.pbftNode.CurChain.AddBlock(block)

	// Relay transactions (leader only)
	if slc.pbftNode.NodeID == 0 {
		slc.pbftNode.pl.Plog.Printf("S%dN%d: sending relay txs at height %d\n", slc.pbftNode.ShardID, slc.pbftNode.NodeID, block.Header.Number)
		slc.pbftNode.CurChain.Txpool.RelayPool = make(map[uint64][]*core.Transaction)
		interShardTxs := make([]*core.Transaction, 0)
		relay1Txs := make([]*core.Transaction, 0)
		relay2Txs := make([]*core.Transaction, 0)
		for _, tx := range block.Body {
			ssid := slc.pbftNode.CurChain.Get_PartitionMap(tx.Sender)
			rsid := slc.pbftNode.CurChain.Get_PartitionMap(tx.Recipient)
			if !tx.Relayed && ssid != slc.pbftNode.ShardID {
				log.Panic("incorrect tx")
			}
			if tx.Relayed && rsid != slc.pbftNode.ShardID {
				log.Panic("incorrect tx")
			}
			if rsid != slc.pbftNode.ShardID {
				relay1Txs = append(relay1Txs, tx)
				tx.Relayed = true
				slc.pbftNode.CurChain.Txpool.AddRelayTx(tx, rsid)
			} else {
				if tx.Relayed {
					relay2Txs = append(relay2Txs, tx)
				} else {
					interShardTxs = append(interShardTxs, tx)
				}
			}
		}

		if params.RelayWithMerkleProof == 1 {
			slc.pbftNode.RelayWithProofSend(block)
		} else {
			slc.pbftNode.RelayMsgSend()
		}

		// Send metrics to supervisor
		bim := message.BlockInfoMsg{
			BlockBodyLength: len(block.Body),
			InnerShardTxs:   interShardTxs,
			Epoch:           0,
			Relay1Txs:       relay1Txs,
			Relay2Txs:       relay2Txs,
			SenderShardID:   slc.pbftNode.ShardID,
			ProposeTime:     r.ReqTime,
			CommitTime:      time.Now(),
		}
		bByte, err := json.Marshal(bim)
		if err != nil {
			log.Panic()
		}
		msg_send := message.MergeMessage(message.CBlockInfo, bByte)
		go networks.TcpDial(msg_send, slc.pbftNode.ip_nodeTable[params.SupervisorShard][0])
		slc.pbftNode.pl.Plog.Printf("S%dN%d: sent executed txs\n", slc.pbftNode.ShardID, slc.pbftNode.NodeID)
		slc.pbftNode.CurChain.Txpool.GetLocked()
		metricName := []string{
			"Block Height",
			"EpochID of this block",
			"TxPool Size",
			"# of all Txs in this block",
			"# of Relay1 Txs in this block",
			"# of Relay2 Txs in this block",
			"TimeStamp - Propose (unixMill)",
			"TimeStamp - Commit (unixMill)",
			"SUM of confirm latency (ms, All Txs)",
			"SUM of confirm latency (ms, Relay1 Txs)",
			"SUM of confirm latency (ms, Relay2 Txs)",
		}
		metricVal := []string{
			strconv.Itoa(int(block.Header.Number)),
			strconv.Itoa(bim.Epoch),
			strconv.Itoa(len(slc.pbftNode.CurChain.Txpool.TxQueue)),
			strconv.Itoa(len(block.Body)),
			strconv.Itoa(len(relay1Txs)),
			strconv.Itoa(len(relay2Txs)),
			strconv.FormatInt(bim.ProposeTime.UnixMilli(), 10),
			strconv.FormatInt(bim.CommitTime.UnixMilli(), 10),
			strconv.FormatInt(computeTCL(block.Body, bim.CommitTime), 10),
			strconv.FormatInt(computeTCL(relay1Txs, bim.CommitTime), 10),
			strconv.FormatInt(computeTCL(relay2Txs, bim.CommitTime), 10),
		}
		slc.pbftNode.writeCSVline(metricName, metricVal)
		slc.pbftNode.CurChain.Txpool.GetUnlocked()
	}
	return true
}

func (slc *RawSLCPbftExtraHandleMod) HandleReqestforOldSeq(*message.RequestOldMessage) bool {
	return true
}

func (slc *RawSLCPbftExtraHandleMod) HandleforSequentialRequest(som *message.SendOldMessage) bool {
	if int(som.SeqEndHeight-som.SeqStartHeight+1) != len(som.OldRequest) {
		slc.pbftNode.pl.Plog.Printf("S%dN%d: incomplete SendOldMessage\n", slc.pbftNode.ShardID, slc.pbftNode.NodeID)
		return false
	}
	for height := som.SeqStartHeight; height <= som.SeqEndHeight; height++ {
		r := som.OldRequest[height-som.SeqStartHeight]
		if r.RequestType == message.BlockRequest {
			b := core.DecodeB(r.Msg.Content)
			slc.pbftNode.CurChain.AddBlock(b)
		}
	}
	slc.pbftNode.sequenceID = som.SeqEndHeight + 1
	slc.pbftNode.CurChain.PrintBlockChain()
	return true
}
