package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/unisat-wallet/libbrc20-indexer-fractal/conf"
	"github.com/unisat-wallet/libbrc20-indexer-fractal/indexer"
	"github.com/unisat-wallet/libbrc20-indexer-fractal/loader"
	"github.com/unisat-wallet/libbrc20-indexer-fractal/utils"
)

var (
	inputfile        string
	outputfile       string
	outputModulefile string
	outputTickerfile string
	testnet          bool
)

func init() {
	flag.BoolVar(&testnet, "testnet", false, "testnet")
	flag.StringVar(&inputfile, "input", "./data/brc20.input.txt", "the filename of input data, default(./data/brc20.input.txt)")
	flag.StringVar(&outputfile, "output", "./data/brc20.output.txt", "the filename of output data, default(./data/brc20.output.txt)")
	flag.StringVar(&outputModulefile, "output_module", "./data/module.output.txt", "the filename of output data, default(./data/module.output.txt)")
	flag.StringVar(&outputTickerfile, "output_ticker", "./data/ticker.output.txt", "the filename of output data, default(./data/ticker.output.txt)")

	flag.Parse()

	if testnet {
		conf.GlobalNetParams = &chaincfg.TestNet3Params
	}
	conf.DEBUG = os.Getenv("DEBUG") == "true"
	if moduleSwapSourceInscriptionId := os.Getenv("MODULE_SWAP_SOURCE_INSCRIPTION_ID"); moduleSwapSourceInscriptionId != "" {
		conf.MODULE_SWAP_SOURCE_INSCRIPTION_ID = moduleSwapSourceInscriptionId
	}
	if ticksEnabled := os.Getenv("TICKS_ENABLED"); ticksEnabled != "" {
		conf.TICKS_ENABLED = ticksEnabled
	}
	if tickMinLen := os.Getenv("TICK_MIN_LEN"); tickMinLen != "" {
		tickMinLenInt, err := strconv.Atoi(tickMinLen)
		if err != nil {
			panic(err)
		}
		conf.TICK_MIN_LEN = tickMinLenInt
	}
	if tickMaxLen := os.Getenv("TICK_MAX_LEN"); tickMaxLen != "" {
		tickMaxLenInt, err := strconv.Atoi(tickMaxLen)
		if err != nil {
			panic(err)
		}
		conf.TICK_MAX_LEN = tickMaxLenInt
	}
	if brc20ModuleSafeConfirmation := os.Getenv("BRC20_MODULE_SAFE_CONFIRMATION"); brc20ModuleSafeConfirmation != "" {
		brc20ModuleSafeConfirmationInt, err := strconv.Atoi(brc20ModuleSafeConfirmation)
		if err != nil {
			panic(err)
		}
		conf.BRC20_MODULE_SAFE_CONFIRMATION = brc20ModuleSafeConfirmationInt
	}
	if enableSelfMintHeight := os.Getenv("ENABLE_SELF_MINT_HEIGHT"); enableSelfMintHeight != "" {
		enableSelfMintHeightInt, err := strconv.Atoi(enableSelfMintHeight)
		if err != nil {
			panic(err)
		}
		conf.ENABLE_SELF_MINT_HEIGHT = enableSelfMintHeightInt
	}
}

func main() {
	brc20Data := loader.LoadBRC20InputData(inputfile)
	latestHeight := utils.GetLatestHeight()

	g := &indexer.BRC20ModuleIndexer{}
	g.ProcessUpdateLatestBRC20Init(brc20Data, latestHeight)

	loader.DumpBRC20InputData(outputfile, brc20Data, true)
	loader.DumpModuleInfoMap(outputModulefile, g.ModulesInfoMap)
	loader.DumpTickerInfoMap(outputTickerfile, g.HistoryData, g.InscriptionsTickerInfoMap, g.UserTokensBalanceData, g.TokenUsersBalanceData)
}
