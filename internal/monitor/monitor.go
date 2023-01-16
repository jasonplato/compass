package monitor

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/compass/internal/chain"
	"github.com/mapprotocol/compass/internal/constant"
	"github.com/mapprotocol/compass/mapprotocol"
	"github.com/mapprotocol/compass/pkg/util"
	"math/big"
	"time"
)

type Monitor struct {
	*chain.CommonSync
	balance, syncedHeight *big.Int
	timestamp             int64
}

func New(cs *chain.CommonSync) *Monitor {
	return &Monitor{
		CommonSync:   cs,
		balance:      new(big.Int),
		syncedHeight: new(big.Int),
	}
}

func (m *Monitor) Sync() error {
	m.Log.Debug("Starting listener...")
	go func() {
		err := m.sync()
		if err != nil {
			m.Log.Error("Polling Account balance failed", "err", err)
		}
	}()

	return nil
}

// sync function of Monitor will poll for the latest block and listen the log information of transactions in the block
// Polling begins at the block defined in `m.Cfg.startBlock`. Failed attempts to fetch the latest block or parse
// a block will be retried up to BlockRetryLimit times before continuing to the next block.
// However，an error in synchronizing the log will cause the entire program to block
func (m *Monitor) sync() error {
	addr := common.HexToAddress(m.Cfg.From)
	waterLine, ok := new(big.Int).SetString(m.Cfg.WaterLine, 10)
	if !ok {
		m.SysErr <- fmt.Errorf("%s waterLine Not Number", m.Cfg.Name)
		return nil
	}
	changeInterval, ok := new(big.Int).SetString(m.Cfg.ChangeInterval, 10)
	if !ok {
		m.SysErr <- fmt.Errorf("%s changeInterval Not Number", m.Cfg.Name)
		return nil
	}
	var heightCount int64
	for {
		select {
		case <-m.Stop:
			return errors.New("polling terminated")
		default:
			balance, err := m.Conn.Client().BalanceAt(context.Background(), addr, nil)
			if err != nil {
				m.Log.Error("Unable to get user balance failed", "from", addr, "err", err)
				time.Sleep(constant.RetryLongInterval)
				continue
			}

			m.Log.Info("Get balance result", "account", addr, "balance", balance)

			if balance.Cmp(m.balance) != 0 {
				m.balance = balance
				m.timestamp = time.Now().Unix()
			}

			if balance.Cmp(waterLine) == -1 {
				// alarm
				util.Alarm(context.Background(),
					fmt.Sprintf("Balance Less than %0.4f Balance,\nchain=%s addr=%s balance=%0.4f",
						float64(new(big.Int).Div(waterLine, constant.Wei).Int64())/float64(constant.Wei.Int64()), m.Cfg.Name, m.Cfg.From,
						float64(balance.Div(balance, constant.Wei).Int64())/float64(constant.Wei.Int64())))
			}

			if (time.Now().Unix() - m.timestamp) > changeInterval.Int64() {
				time.Sleep(time.Second * 5)
				// alarm
				util.Alarm(context.Background(),
					fmt.Sprintf("No transaction occurred in addr in the last %d seconds,\n"+
						"chain=%s addr=%s balance=%0.4f", changeInterval.Int64(), m.Cfg.Name, m.Cfg.From,
						float64(balance.Div(balance, constant.Wei).Int64())/float64(constant.Wei.Int64())))
			}

			if m.Cfg.Id == m.Cfg.MapChainID {

			} else {
				height, err := mapprotocol.Get2MapHeight(m.Cfg.Id)
				m.Log.Info("Check Height", "syncHeight", height)
				if err != nil {
					m.Log.Error("get2MapHeight failed", "err", err)
				} else {
					if m.syncedHeight == height {
						heightCount++
						if heightCount == 20 {
							util.Alarm(context.Background(),
								fmt.Sprintf("Maintainer Sync Height No change within 15 minutes chain=%s, height=%d",
									m.Cfg.Name, height.Uint64()))
						}
					} else {
						heightCount = 0
					}
					m.syncedHeight = height
				}
			}

			time.Sleep(constant.BalanceRetryInterval)
		}
	}
}
