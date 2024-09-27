package indexer

import (
	"bytes"
	"log"

	"github.com/unisat-wallet/libbrc20-indexer-fractal/conf"
	"github.com/unisat-wallet/libbrc20-indexer-fractal/constant"
	"github.com/unisat-wallet/libbrc20-indexer-fractal/model"
)

func isJson(contentBody []byte) bool {
	if len(contentBody) < 40 {
		return false
	}

	content := bytes.TrimSpace(contentBody)
	if !bytes.HasPrefix(content, []byte("{")) {
		return false
	}
	if !bytes.HasSuffix(content, []byte("}")) {
		return false
	}

	return true
}

// ProcessUpdateLatestBRC20Loop
func (g *BRC20ModuleIndexer) ProcessUpdateLatestBRC20Loop(brc20Datas []*model.InscriptionBRC20Data, latestHeight int) {
	totalDataCount := len(brc20Datas)
	log.Printf("process swap update. total %d", totalDataCount)
	lastPercent := 0
	for idx, data := range brc20Datas {
		percent := idx * 100 / totalDataCount
		if lastPercent+3 < percent {
			lastPercent = percent
			log.Printf("(%d %%) brc20 processing...", percent)
		}
		// is sending transfer
		if data.IsTransfer {
			// not first move
			if data.Sequence != 1 {
				continue
			}

			// transfer
			if transferInfo, isInvalid := g.GetTransferInfoByKey(data.CreateIdxKey); transferInfo != nil {
				if err := g.ProcessTransfer(data, transferInfo, latestHeight, isInvalid); err != nil {
					log.Printf("process transfer move failed: %s", err)
				}
				continue
			}

			// module approve
			if approveInfo, isInvalid := g.GetApproveInfoByKey(data.CreateIdxKey); approveInfo != nil {
				if err := g.ProcessApprove(data, approveInfo, latestHeight, isInvalid); err != nil {
					log.Printf("process approve move failed: %s", err)
				}
				continue
			}

			// module withdraw
			if withdrawInfo := g.GetWithdrawInfoByKey(data.CreateIdxKey); withdrawInfo != nil {
				if err := g.ProcessWithdraw(data, withdrawInfo); err != nil {
					log.Printf("process withdraw move failed: %s", err)
				}
				continue
			}

			// module commit
			if commitFrom, isInvalid := g.GetCommitInfoByKey(data.CreateIdxKey); commitFrom != nil {
				if err := g.ProcessCommit(commitFrom, data, isInvalid); err != nil {
					log.Printf("process commit move failed: %s", err)
				}
				continue
			}

			continue
		}

		// inscribe as fee
		if data.Satoshi == 0 {
			continue
		}

		if ok := isJson(data.ContentBody); !ok {
			// log.Println("not json")
			continue
		}

		// protocal, lower case only
		body := new(model.InscriptionBRC20ProtocalContent)
		if err := body.Unmarshal(data.ContentBody); err != nil {
			// log.Println("Unmarshal failed", err, string(data.ContentBody))
			continue
		}

		// is inscribe deploy/mint/transfer
		if body.Proto != constant.BRC20_P &&
			body.Proto != constant.BRC20_P_MODULE &&
			body.Proto != constant.BRC20_P_SWAP {
			// log.Println("not proto")
			continue
		}

		var process func(*model.InscriptionBRC20Data) error
		if body.Proto == constant.BRC20_P && body.Operation == constant.BRC20_OP_DEPLOY {
			process = g.ProcessDeploy
		} else if body.Proto == constant.BRC20_P && body.Operation == constant.BRC20_OP_MINT {
			process = g.ProcessMint
		} else if body.Proto == constant.BRC20_P && body.Operation == constant.BRC20_OP_TRANSFER {
			process = g.ProcessInscribeTransfer
		} else if body.Proto == constant.BRC20_P_MODULE && body.Operation == constant.BRC20_OP_MODULE_DEPLOY {
			process = g.ProcessCreateModule
		} else if body.Proto == constant.BRC20_P_MODULE && body.Operation == constant.BRC20_OP_MODULE_WITHDRAW {
			process = g.ProcessInscribeWithdraw
		} else if body.Proto == constant.BRC20_P_SWAP && body.Operation == constant.BRC20_OP_SWAP_APPROVE {
			process = g.ProcessInscribeApprove
		} else if body.Proto == constant.BRC20_P_SWAP && body.Operation == constant.BRC20_OP_SWAP_COMMIT {
			process = g.ProcessInscribeCommit
		} else {
			continue
		}

		if err := process(data); err != nil {
			if conf.DEBUG {
				log.Printf("(%d %%) process failed: %s", percent, err)
			}
		}
	}

	for _, holdersBalanceMap := range g.TokenUsersBalanceData {
		for key, balance := range holdersBalanceMap {
			if balance.AvailableBalance.Sign() == 0 && balance.TransferableBalance.Sign() == 0 {
				delete(holdersBalanceMap, key)
			}
		}
	}

	log.Printf("process swap finish. ticker: %d, users: %d, tokens: %d, validInscription: %d, validTransfer: %d, invalidTransfer: %d, history: %d",
		len(g.InscriptionsTickerInfoMap),
		len(g.UserTokensBalanceData),
		len(g.TokenUsersBalanceData),

		len(g.InscriptionsValidBRC20DataMap),

		len(g.InscriptionsValidTransferMap),
		len(g.InscriptionsInvalidTransferMap),

		g.HistoryCount,
	)

	nswap := 0
	for _, m := range g.ModulesInfoMap {
		nswap += len(m.SwapPoolTotalBalanceDataMap)
	}

	nuser := 0
	for _, m := range g.ModulesInfoMap {
		nuser += len(m.UsersTokenBalanceDataMap)
	}

	log.Printf("process swap finish. module: %d, swap: %d, users: %d, validApprove: %d, invalidApprove: %d, validCommit: %d, invalidCommit: %d",
		len(g.ModulesInfoMap),
		nswap,
		nuser,

		len(g.InscriptionsValidApproveMap),
		len(g.InscriptionsInvalidApproveMap),

		len(g.InscriptionsValidCommitMap),
		len(g.InscriptionsInvalidCommitMap),
	)

}

// ProcessUpdateLatestBRC20Init
func (g *BRC20ModuleIndexer) ProcessUpdateLatestBRC20Init(brc20Datas []*model.InscriptionBRC20Data, latestHeight int) {
	log.Printf("process swap init. total %d", len(brc20Datas))

	g.initBRC20()
	g.initModule()

	g.ProcessUpdateLatestBRC20Loop(brc20Datas, latestHeight)
}
