package indexer

import (
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/unisat-wallet/libbrc20-indexer-fractal/constant"
	"github.com/unisat-wallet/libbrc20-indexer-fractal/model"
)

var (
	mutex = sync.Mutex{}
)

type BRC20ModuleIndexer struct {
	EnableHistory bool

	HistoryCount uint32
	HistoryData  [][]byte

	// history height
	FirstHistoryByHeight map[uint32]uint32
	LastHistoryHeight    uint32
	FirstMempoolHistory  uint32

	// brc20 base
	UserAllHistory map[string]*model.BRC20UserHistory

	InscriptionsTickerInfoMap     map[string]*model.BRC20TokenInfo
	UserTokensBalanceData         map[string]map[string]*model.BRC20TokenBalance // [address][ticker]balance
	TokenUsersBalanceData         map[string]map[string]*model.BRC20TokenBalance // [ticker][address]balance
	InscriptionsValidBRC20DataMap map[uint64]*model.InscriptionBRC20InfoResp

	TokenUsersBalanceDataSortedCache map[string][]*model.BRC20TokenBalance // [ticker][]balance cache

	// inner valid transfer
	InscriptionsValidTransferMap map[uint64]*model.InscriptionBRC20TickInfo
	// inner invalid transfer
	InscriptionsInvalidTransferMap map[uint64]*model.InscriptionBRC20TickInfo

	// module
	// all modules info
	ModulesInfoMap map[string]*model.BRC20ModuleSwapInfo

	// module of users [address]moduleid
	UsersModuleWithTokenMap map[string]string

	// module lp of users [address]moduleid
	UsersModuleWithLpTokenMap map[string]string

	// runtime for approve
	InscriptionsValidApproveMap   map[uint64]*model.InscriptionBRC20SwapInfo // inner valid approve [create_key]
	InscriptionsInvalidApproveMap map[uint64]*model.InscriptionBRC20SwapInfo

	// runtime for commit
	InscriptionsValidCommitMap   map[uint64]*model.InscriptionBRC20Data // inner valid commit by key
	InscriptionsInvalidCommitMap map[uint64]*model.InscriptionBRC20Data

	InscriptionsValidCommitMapById map[string]*model.InscriptionBRC20Data // inner valid commit by id

	// runtime for withdraw
	InscriptionsWithdrawMap map[uint64]*model.InscriptionBRC20SwapInfo // inner all ready to withdraw by key
}

func (g *BRC20ModuleIndexer) GetBRC20HistoryByUser(pkScript string) (userHistory *model.BRC20UserHistory) {
	if history, ok := g.UserAllHistory[pkScript]; !ok {
		userHistory = &model.BRC20UserHistory{}
		g.UserAllHistory[pkScript] = userHistory
	} else {
		userHistory = history
	}
	return userHistory
}

func (g *BRC20ModuleIndexer) GetBRC20HistoryByUserForAPI(pkScript string) (userHistory *model.BRC20UserHistory) {
	if history, ok := g.UserAllHistory[pkScript]; !ok {
		userHistory = &model.BRC20UserHistory{}
	} else {
		userHistory = history
	}
	return userHistory
}

func (g *BRC20ModuleIndexer) GetBRC20TokenUsersBalanceDataSortedCacheForAPI(ticker string) (holdersBalance []*model.BRC20TokenBalance) {
	mutex.Lock()
	defer mutex.Unlock()

	holdersBalance, ok := g.TokenUsersBalanceDataSortedCache[ticker]
	if ok {
		return holdersBalance
	}

	tokenUsers, ok := g.TokenUsersBalanceData[ticker]
	if !ok {
		return make([]*model.BRC20TokenBalance, 0)
	}

	for _, balance := range tokenUsers {
		holdersBalance = append(holdersBalance, balance)
	}

	sort.Slice(holdersBalance, func(i, j int) bool {
		return strings.Compare(holdersBalance[i].PkScript, holdersBalance[j].PkScript) > 0
	})

	sort.SliceStable(holdersBalance, func(i, j int) bool {
		return holdersBalance[i].OverallBalance().Cmp(holdersBalance[j].OverallBalance()) > 0
	})

	g.TokenUsersBalanceDataSortedCache[ticker] = holdersBalance

	return holdersBalance
}

func (g *BRC20ModuleIndexer) UpdateHistoryHeightAndGetHistoryIndex(historyObj *model.BRC20History) uint32 {
	height := historyObj.Height
	history := g.HistoryCount
	g.HistoryData = append(g.HistoryData, historyObj.Marshal())
	g.HistoryCount += 1

	if height == g.LastHistoryHeight {
		return history
	}

	if height == constant.MEMPOOL_HEIGHT {
		if g.FirstMempoolHistory == 0 {
			g.FirstMempoolHistory = history
		}
		return history
	}

	if g.LastHistoryHeight == 0 {
		g.FirstHistoryByHeight[height] = history
	} else {
		for h := g.LastHistoryHeight + 1; h <= height; h++ {
			g.FirstHistoryByHeight[h] = history
		}
	}
	g.LastHistoryHeight = height

	return history
}

func (g *BRC20ModuleIndexer) initBRC20() {
	g.EnableHistory = true

	g.HistoryCount = 0
	g.HistoryData = make([][]byte, 0)

	g.FirstHistoryByHeight = make(map[uint32]uint32, 0)
	g.LastHistoryHeight = 0
	g.FirstMempoolHistory = 0

	// user history
	g.UserAllHistory = make(map[string]*model.BRC20UserHistory, 0)

	// all ticker info
	g.InscriptionsTickerInfoMap = make(map[string]*model.BRC20TokenInfo, 0)

	// ticker of users
	g.UserTokensBalanceData = make(map[string]map[string]*model.BRC20TokenBalance, 0)

	// ticker holders
	g.TokenUsersBalanceData = make(map[string]map[string]*model.BRC20TokenBalance, 0)

	// ticker holders sorted cache
	g.TokenUsersBalanceDataSortedCache = make(map[string][]*model.BRC20TokenBalance, 0)

	// valid brc20 inscriptions
	g.InscriptionsValidBRC20DataMap = make(map[uint64]*model.InscriptionBRC20InfoResp, 0)

	// inner valid transfer
	g.InscriptionsValidTransferMap = make(map[uint64]*model.InscriptionBRC20TickInfo, 0)
	// inner invalid transfer
	g.InscriptionsInvalidTransferMap = make(map[uint64]*model.InscriptionBRC20TickInfo, 0)
}

func (g *BRC20ModuleIndexer) initModule() {
	// all modules info
	g.ModulesInfoMap = make(map[string]*model.BRC20ModuleSwapInfo, 0)

	// module of users [address]moduleid
	g.UsersModuleWithTokenMap = make(map[string]string, 0)

	// swap
	// module of users [address]moduleid
	g.UsersModuleWithLpTokenMap = make(map[string]string, 0)

	// runtime for approve
	g.InscriptionsValidApproveMap = make(map[uint64]*model.InscriptionBRC20SwapInfo, 0)
	g.InscriptionsInvalidApproveMap = make(map[uint64]*model.InscriptionBRC20SwapInfo, 0)

	// runtime for commit
	g.InscriptionsValidCommitMap = make(map[uint64]*model.InscriptionBRC20Data, 0) // inner valid commit
	g.InscriptionsInvalidCommitMap = make(map[uint64]*model.InscriptionBRC20Data, 0)

	g.InscriptionsValidCommitMapById = make(map[string]*model.InscriptionBRC20Data, 0) // inner valid commit

	// runtime for withdraw
	g.InscriptionsWithdrawMap = make(map[uint64]*model.InscriptionBRC20SwapInfo, 0)
}

func (g *BRC20ModuleIndexer) GetUserTokenBalance(ticker, userPkScript string) (tokenBalance *model.BRC20TokenBalance) {
	uniqueLowerTicker := strings.ToLower(ticker)
	// get user's tokens to update
	var userTokens map[string]*model.BRC20TokenBalance
	if tokens, ok := g.UserTokensBalanceData[userPkScript]; !ok {
		userTokens = make(map[string]*model.BRC20TokenBalance, 0)
		g.UserTokensBalanceData[userPkScript] = userTokens
	} else {
		userTokens = tokens
	}
	// get tokenBalance to update
	if tb, ok := userTokens[uniqueLowerTicker]; !ok {
		tokenBalance = &model.BRC20TokenBalance{Ticker: ticker, PkScript: userPkScript}
		userTokens[uniqueLowerTicker] = tokenBalance
	} else {
		tokenBalance = tb
	}
	// set token's users
	tokenUsers, ok := g.TokenUsersBalanceData[uniqueLowerTicker]
	if !ok {
		log.Panicf("g.TokenUsersBalanceData[%s], not exists", uniqueLowerTicker)
	}
	tokenUsers[userPkScript] = tokenBalance

	return tokenBalance
}

func (copyDup *BRC20ModuleIndexer) deepCopyBRC20Data(base *BRC20ModuleIndexer, withData bool) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Printf("deepCopyBRC20Data history start. total: %d", base.HistoryCount)
		// history
		copyDup.EnableHistory = base.EnableHistory
		copyDup.HistoryCount = base.HistoryCount

		for height, history := range base.FirstHistoryByHeight {
			copyDup.FirstHistoryByHeight[height] = history
		}
		copyDup.LastHistoryHeight = base.LastHistoryHeight
		copyDup.FirstMempoolHistory = base.FirstMempoolHistory

		for _, h := range base.HistoryData {
			copyDup.HistoryData = append(copyDup.HistoryData, h)
		}
		log.Printf("deepCopyBRC20Data history finish. total: %d", base.HistoryCount)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Printf("deepCopyBRC20Data user history start. total: %d", len(base.UserAllHistory))

		// userhistory
		for u, userHistory := range base.UserAllHistory {
			h := &model.BRC20UserHistory{
				History: make([]uint32, len(userHistory.History)),
			}
			copy(h.History, userHistory.History)
			copyDup.UserAllHistory[u] = h
		}

		log.Printf("deepCopyBRC20Data user history finish. total: %d", len(base.UserAllHistory))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Printf("deepCopyBRC20Data tick info start. total: %d", len(base.InscriptionsTickerInfoMap))
		for k, v := range base.InscriptionsTickerInfoMap {
			tinfo := &model.BRC20TokenInfo{
				Ticker:   v.Ticker,
				SelfMint: v.SelfMint,
				Deploy:   v.Deploy.DeepCopy(),
			}

			// history
			tinfo.History = make([]uint32, len(v.History))
			copy(tinfo.History, v.History)

			tinfo.HistoryMint = make([]uint32, len(v.HistoryMint))
			copy(tinfo.HistoryMint, v.HistoryMint)

			tinfo.HistoryInscribeTransfer = make([]uint32, len(v.HistoryInscribeTransfer))
			copy(tinfo.HistoryInscribeTransfer, v.HistoryInscribeTransfer)

			tinfo.HistoryTransfer = make([]uint32, len(v.HistoryTransfer))
			copy(tinfo.HistoryTransfer, v.HistoryTransfer)

			// set info
			copyDup.InscriptionsTickerInfoMap[k] = tinfo
		}
		log.Printf("deepCopyBRC20Data tick info finish. total: %d", len(base.InscriptionsTickerInfoMap))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Printf("deepCopyBRC20Data user balance start. total: %d", len(base.UserTokensBalanceData))
		for u, userTokens := range base.UserTokensBalanceData {
			userTokensCopy := make(map[string]*model.BRC20TokenBalance, 0)
			copyDup.UserTokensBalanceData[u] = userTokensCopy
			for uniqueLowerTicker, v := range userTokens {
				tb := v.DeepCopy()
				userTokensCopy[uniqueLowerTicker] = tb

				tokenUsers, ok := copyDup.TokenUsersBalanceData[uniqueLowerTicker]
				if !ok {
					tokenUsers = make(map[string]*model.BRC20TokenBalance, 0)
					copyDup.TokenUsersBalanceData[uniqueLowerTicker] = tokenUsers
				}
				tokenUsers[u] = tb
			}
		}
		log.Printf("deepCopyBRC20Data user balance finish. total: %d", len(base.UserTokensBalanceData))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if withData {
			log.Printf("deepCopyBRC20Data valid data start. total: %d", len(base.InscriptionsValidBRC20DataMap))
			for k, v := range base.InscriptionsValidBRC20DataMap {
				copyDup.InscriptionsValidBRC20DataMap[k] = v
			}
			log.Printf("deepCopyBRC20Data valid data finish. total: %d", len(base.InscriptionsValidBRC20DataMap))
		}

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Printf("deepCopyBRC20Data valid transfer start. total: %d", len(base.InscriptionsValidTransferMap))
		// transferInfo
		for k, v := range base.InscriptionsValidTransferMap {
			copyDup.InscriptionsValidTransferMap[k] = v
		}

		// fixme: disable invalid copy
		// for k, v := range base.InscriptionsInvalidTransferMap {
		// 	copyDup.InscriptionsInvalidTransferMap[k] = v
		// }
		log.Printf("deepCopyBRC20Data valid transfer finish. total: %d", len(base.InscriptionsValidTransferMap))
	}()

	wg.Wait()
	log.Printf("deepCopyBRC20Data finish. total: %d", len(base.InscriptionsTickerInfoMap))
}

func (copyDup *BRC20ModuleIndexer) cherryPickBRC20Data(base *BRC20ModuleIndexer, pickUsersPkScript, pickTokensTick map[string]bool) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for lowerTick := range pickTokensTick {
			v, ok := base.InscriptionsTickerInfoMap[lowerTick]
			if !ok {
				continue
			}

			tinfo := &model.BRC20TokenInfo{
				Ticker:   v.Ticker,
				SelfMint: v.SelfMint,
				Deploy:   v.Deploy.DeepCopy(),
			}
			copyDup.InscriptionsTickerInfoMap[lowerTick] = tinfo
		}

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for u := range pickUsersPkScript {
			userTokens, ok := base.UserTokensBalanceData[u]
			if !ok {
				continue
			}
			userTokensCopy := make(map[string]*model.BRC20TokenBalance, 0)
			for lowerTick := range pickTokensTick {
				balance, ok := userTokens[lowerTick]
				if !ok {
					continue
				}
				userTokensCopy[lowerTick] = balance.DeepCopy()
			}
			copyDup.UserTokensBalanceData[u] = userTokensCopy
		}

		for u, userTokens := range copyDup.UserTokensBalanceData {
			for uniqueLowerTicker, balance := range userTokens {
				tokenUsers, ok := copyDup.TokenUsersBalanceData[uniqueLowerTicker]
				if !ok {
					tokenUsers = make(map[string]*model.BRC20TokenBalance, 0)
					copyDup.TokenUsersBalanceData[uniqueLowerTicker] = tokenUsers
				}
				tokenUsers[u] = balance
			}
		}
	}()

	wg.Wait()

	log.Printf("cherryPickBRC20Data finish. total: %d", len(copyDup.InscriptionsTickerInfoMap))
}

func (copyDup *BRC20ModuleIndexer) deepCopyModuleData(base *BRC20ModuleIndexer) {

	for module, info := range base.ModulesInfoMap {
		copyDup.ModulesInfoMap[module] = info.DeepCopy()
	}

	// module of users
	for k, v := range base.UsersModuleWithTokenMap {
		copyDup.UsersModuleWithTokenMap[k] = v
	}

	// module lp of users
	for k, v := range base.UsersModuleWithLpTokenMap {
		copyDup.UsersModuleWithLpTokenMap[k] = v
	}

	// approveInfo
	for k, v := range base.InscriptionsValidApproveMap {
		copyDup.InscriptionsValidApproveMap[k] = v
	}
	for k, v := range base.InscriptionsInvalidApproveMap {
		copyDup.InscriptionsInvalidApproveMap[k] = v
	}

	// commitInfo
	for k, v := range base.InscriptionsValidCommitMap {
		copyDup.InscriptionsValidCommitMap[k] = v
	}
	for k, v := range base.InscriptionsInvalidCommitMap {
		copyDup.InscriptionsInvalidCommitMap[k] = v
	}

	for k, v := range base.InscriptionsValidCommitMapById {
		copyDup.InscriptionsValidCommitMapById[k] = v
	}

	// withdraw
	for k, v := range base.InscriptionsWithdrawMap {
		copyDup.InscriptionsWithdrawMap[k] = v
	}

	log.Printf("deepCopyModuleData finish. total: %d", len(base.ModulesInfoMap))
}

func (copyDup *BRC20ModuleIndexer) cherryPickModuleData(base *BRC20ModuleIndexer, module string, pickUsersPkScript, pickTokensTick, pickPoolsPair map[string]bool) {

	info, ok := base.ModulesInfoMap[module]
	if ok {
		copyDup.ModulesInfoMap[module] = info.CherryPick(pickUsersPkScript, pickTokensTick, pickPoolsPair)
	}

	// Data required for verification
	for k, v := range base.InscriptionsValidCommitMapById {
		copyDup.InscriptionsValidCommitMapById[k] = v
	}
	log.Printf("cherryPickModuleData finish. total: %d", len(base.ModulesInfoMap))
}

func (base *BRC20ModuleIndexer) DeepCopy(withData bool) (copyDup *BRC20ModuleIndexer) {
	copyDup = &BRC20ModuleIndexer{}
	copyDup.initBRC20()
	copyDup.initModule()

	copyDup.deepCopyBRC20Data(base, withData)
	copyDup.deepCopyModuleData(base)
	return copyDup
}

func (base *BRC20ModuleIndexer) CherryPick(module string, pickUsersPkScript, pickTokensTick, pickPoolsPair map[string]bool) (copyDup *BRC20ModuleIndexer) {
	copyDup = &BRC20ModuleIndexer{}
	copyDup.initBRC20()
	copyDup.initModule()

	moduleInfo, ok := base.ModulesInfoMap[module]
	if ok {
		lowerTick := strings.ToLower(moduleInfo.GasTick)
		pickTokensTick[lowerTick] = true
	}
	copyDup.cherryPickBRC20Data(base, pickUsersPkScript, pickTokensTick)
	copyDup.cherryPickModuleData(base, module, pickUsersPkScript, pickTokensTick, pickPoolsPair)
	return copyDup
}
