package writer

import (
	"context"
	"fmt"
	"github.com/mapprotocol/compass/internal/constant"
	"github.com/mapprotocol/compass/msg"
	"github.com/mapprotocol/compass/pkg/util"
	"strings"
	"time"
)

// execMap2OtherMsg executes sync msg, and send tx to the destination blockchain
func (w *Writer) execMap2OtherMsg(m msg.Message) bool {
	var (
		errorCount int64
		needNonce  = true
	)
	for {
		select {
		case <-w.stop:
			return false
		default:
			err := w.conn.LockAndUpdateOpts(needNonce)
			if err != nil {
				w.log.Error("Failed to update nonce", "err", err)
				return false
			}
			// These store the gas limit and price before a transaction is sent for logging in case of a failure
			// This is necessary as tx will be nil in the case of an error when sending VoteProposal()
			tx, err := w.sendTx(&w.cfg.LightNode, nil, m.Payload[0].([]byte))
			w.conn.UnlockOpts()
			if err == nil {
				// message successfully handled
				w.log.Info("Sync Map Header to other chain tx execution", "tx", tx.Hash(), "src", m.Source, "dst", m.Destination, "needNonce", needNonce, "nonce", w.conn.Opts().Nonce)
				err = w.txStatus(tx.Hash())
				if err != nil {
					w.log.Warn("TxHash Status is not successful, will retry", "err", err)
				} else {
					m.DoneCh <- struct{}{}
					return true
				}
			} else if strings.Index(err.Error(), constant.EthOrderExist) != -1 {
				w.log.Info(constant.EthOrderExistPrint, "id", m.Destination, "err", err)
				m.DoneCh <- struct{}{}
				return true
			} else if strings.Index(err.Error(), constant.HeaderIsHave) != -1 {
				w.log.Info(constant.HeaderIsHavePrint, "id", m.Destination, "err", err)
				m.DoneCh <- struct{}{}
				return true
			} else if strings.Index(err.Error(), "EOF") != -1 {
				w.log.Error("Sync Header to map encounter EOF, will retry", "id", m.Destination)
			} else if err.Error() == constant.ErrNonceTooLow.Error() || err.Error() == constant.ErrTxUnderpriced.Error() {
				w.log.Error("Sync Map Header to other chain Nonce too low, will retry", "id", m.Destination)
			} else if strings.Index(err.Error(), constant.NotEnoughGas) != -1 {
				w.log.Error(constant.NotEnoughGasPrint, "id", m.Destination)
			} else {
				w.log.Warn("Sync Map Header to other chain Execution failed, header may already been synced", "id", m.Destination, "err", err)
			}
			needNonce = false
			errorCount++
			if errorCount >= 10 {
				util.Alarm(context.Background(), fmt.Sprintf("writer map to other(%d) header failed, err is %s", m.Destination, err.Error()))
				errorCount = 0
			}
			time.Sleep(constant.TxRetryInterval)
		}
	}
	//w.log.Error("Sync Map Header to other chain Submission of Sync MapHeader transaction failed", "source", m.Source,
	//	"dest", m.Destination, "depositNonce", m.DepositNonce)
	//w.sysErr <- constant.ErrFatalTx
	//return false
}
