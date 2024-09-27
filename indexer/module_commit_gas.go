package indexer

import (
	"encoding/hex"
	"errors"
	"log"

	"github.com/unisat-wallet/libbrc20-indexer-fractal/conf"
	"github.com/unisat-wallet/libbrc20-indexer-fractal/decimal"
	"github.com/unisat-wallet/libbrc20-indexer-fractal/model"
	"github.com/unisat-wallet/libbrc20-indexer-fractal/utils"
)

func (g *BRC20ModuleIndexer) ProcessCommitFunctionGasFee(moduleInfo *model.BRC20ModuleSwapInfo, userPkScript string, gasAmt *decimal.Decimal) error {

	tokenBalance := moduleInfo.GetUserTokenBalance(moduleInfo.GasTick, userPkScript)
	// fixme: Must use the confirmed amount
	if tokenBalance.SwapAccountBalance.Cmp(gasAmt) < 0 {
		address, err := utils.GetAddressFromScript([]byte(userPkScript), conf.GlobalNetParams)
		if err != nil {
			address = hex.EncodeToString([]byte(userPkScript))
		}

		log.Printf("gas[%s] user[%s], balance %s", moduleInfo.GasTick, address, tokenBalance)
		return errors.New("gas fee: token balance insufficient")
	}

	gasToBalance := moduleInfo.GetUserTokenBalance(moduleInfo.GasTick, moduleInfo.GasToPkScript)

	// User Real-time gas Balance Update
	tokenBalance.SwapAccountBalance = tokenBalance.SwapAccountBalance.Sub(gasAmt)
	tokenBalance.SwapAccountBalanceSafe = tokenBalance.SwapAccountBalanceSafe.Sub(gasAmt)
	gasToBalance.SwapAccountBalance = gasToBalance.SwapAccountBalance.Add(gasAmt)
	gasToBalance.SwapAccountBalanceSafe = gasToBalance.SwapAccountBalanceSafe.Add(gasAmt)

	// log.Printf("gas fee[%s]: %s user: %s, gasTo: %s", moduleInfo.GasTick, gasAmt, tokenBalance.SwapAccountBalance, gasToBalance.SwapAccountBalance)
	return nil
}
