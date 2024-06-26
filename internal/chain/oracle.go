package chain

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/mapprotocol/compass/internal/constant"
	"github.com/mapprotocol/compass/internal/proof"
	"github.com/mapprotocol/compass/internal/tx"
	"github.com/mapprotocol/compass/mapprotocol"
	"github.com/mapprotocol/compass/msg"
	"github.com/mapprotocol/compass/pkg/util"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

type Oracle struct {
	*CommonSync
}

func NewOracle(cs *CommonSync) *Oracle {
	return &Oracle{
		CommonSync: cs,
	}
}

func (m *Oracle) Sync() error {
	m.Log.Debug("Starting listener...")
	go func() {
		err := m.sync()
		if err != nil {
			m.Log.Error("Polling blocks failed", "err", err)
		}
	}()

	return nil
}

func (m *Oracle) sync() error {
	if !m.Cfg.SyncToMap && m.Cfg.Id != m.Cfg.MapChainID {
		time.Sleep(time.Hour * 2400)
		return nil
	}
	var currentBlock = m.Cfg.StartBlock

	for {
		select {
		case <-m.Stop:
			return errors.New("polling terminated")
		default:
			latestBlock, err := m.Conn.LatestBlock()
			if err != nil {
				m.Log.Error("Unable to get latest block", "block", currentBlock, "err", err)
				time.Sleep(constant.RetryLongInterval)
				continue
			}

			if big.NewInt(0).Sub(latestBlock, currentBlock).Cmp(m.BlockConfirmations) == -1 {
				m.Log.Debug("Block not ready, will retry", "currentBlock", currentBlock, "latest", latestBlock)
				time.Sleep(constant.BalanceRetryInterval)
				continue
			}

			err = m.oracleHandler(m, currentBlock)
			if err != nil {
				m.Log.Error("Failed to get events for block", "block", currentBlock, "err", err)
				time.Sleep(constant.BlockRetryInterval)
				util.Alarm(context.Background(), fmt.Sprintf("mos failed, chain=%s, err is %s", m.Cfg.Name, err.Error()))
				continue
			}

			err = m.BlockStore.StoreBlock(currentBlock)
			if err != nil {
				m.Log.Error("Failed to write latest block to blockstore", "block", currentBlock, "err", err)
			}

			currentBlock.Add(currentBlock, big.NewInt(1))
			if latestBlock.Int64()-currentBlock.Int64() <= m.Cfg.BlockConfirmations.Int64() {
				time.Sleep(constant.MessengerInterval)
			}
		}
	}
}

func DefaultOracleHandler(m *Oracle, latestBlock *big.Int) error {
	m.Log.Debug("Querying block for events", "block", latestBlock)
	count := 0
	query := m.BuildQuery(m.Cfg.OracleNode, m.Cfg.Events, latestBlock, latestBlock)
	// querying for logs
	logs, err := m.Conn.Client().FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("oracle unable to Filter Logs: %w", err)
	}
	header, err := m.Conn.Client().HeaderByNumber(context.Background(), latestBlock)
	if err != nil {
		return fmt.Errorf("oracle get header failed, err: %w", err)
	}
	if len(logs) == 0 {
		return nil
	}
	if m.Cfg.Id == constant.MerlinChainId {
		txsHash, err := mapprotocol.GetMapTransactionsHashByBlockNumber(m.Conn.Client(), latestBlock)
		if err != nil {
			return fmt.Errorf("unable to get tx hashes Logs: %w", err)
		}
		receipts, err := tx.GetReceiptsByTxsHash(m.Conn.Client(), txsHash)
		if err != nil {
			return fmt.Errorf("unable to get receipts hashes Logs: %w", err)
		}
		tr, _ := trie.New(common.Hash{}, trie.NewDatabase(memorydb.New()))
		tr = proof.DeriveTire(types.Receipts(receipts), tr)
		m.Log.Info("oracle merlin receipt", "blockNumber", latestBlock, "hash", tr.Hash())
		header.ReceiptHash = tr.Hash()
	}
	m.Log.Info("Find log", "block", latestBlock, "logs", len(logs))
	var input []byte
	if m.Cfg.ApiUrl == "" {
		input, err = mapprotocol.OracleAbi.Methods[mapprotocol.MethodOfPropose].Inputs.Pack(header.Number, header.ReceiptHash)
	} else { // todo
		proof, err := mapprotocol.GetZkProof(m.Cfg.ApiUrl, m.Cfg.Id, latestBlock.Uint64())
		if err != nil {
			return err
		}
		validators, err := mapprotocol.GetCurValidators(m.Conn.Client(), latestBlock)
		if err != nil {
			return err
		}
		input, err = mapprotocol.OracleAbi.Methods[mapprotocol.MethodOfPropose].Inputs.Pack(validators, header.Number, header.ReceiptHash, proof)
	}
	if err != nil {
		return err
	}

	if m.Cfg.Id == m.Cfg.MapChainID {
		data, err := mapprotocol.PackInput(mapprotocol.LightManger, mapprotocol.MethodUpdateBlockHeader, big.NewInt(int64(m.Cfg.Id)), input)
		if err != nil {
			return err
		}

		for _, cid := range m.Cfg.SyncChainIDList {
			message := msg.NewSyncFromMap(m.Cfg.MapChainID, cid, []interface{}{data}, m.MsgCh)
			err = m.Router.Send(message)
			if err != nil {
				m.Log.Error("subscription error: failed to route message", "err", err)
				return nil
			}
			count++
		}
	} else {
		id := big.NewInt(0).SetUint64(uint64(m.Cfg.Id))
		message := msg.NewSyncToMap(m.Cfg.Id, m.Cfg.MapChainID, []interface{}{id, input}, m.MsgCh)
		err = m.Router.Send(message)
		if err != nil {
			m.Log.Error("subscription error: failed to route message", "err", err)
			return nil
		}
		count++
	}

	err = m.WaitUntilMsgHandled(count)
	if err != nil {
		return err
	}
	return nil
}
