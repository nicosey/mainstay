package main

import (
	"context"
	"flag"
    "os"
    "os/signal"
    "sync"
    "log"
    "time"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"ocean-attestation/attestation"
	"ocean-attestation/conf"
	"ocean-attestation/models"
	"ocean-attestation/requestapi"
	"ocean-attestation/test"
)

const DEFAULT_API_HOST = "localhost:8080"
const MAIN_NAME = "main"
const OCEAN_NAME = "ocean"

var (
	genesisTX               string
	mainClient, oceanClient *rpcclient.Client
	mainChainCfg            *chaincfg.Params
	isRegtest               bool
    apiHost                 string
)

func parseFlags() {
	flag.BoolVar(&isRegtest, "regtest", false, "Use regtest wallet configuration instead of user wallet")
	flag.StringVar(&genesisTX, "tx", "", "Tx id for genesis attestation transaction")
	flag.Parse()

	if (genesisTX == "") && !isRegtest {
		flag.PrintDefaults()
		log.Fatalf("Need to provide -tx argument. To use test configuration set the -regtest flag.")
	}
}

func init() {
	parseFlags()
	if isRegtest { // Use configuration applied in unit tests
		test := test.NewTest(true, true)
		mainClient = test.Btc
		oceanClient = test.Ocean
		mainChainCfg = test.BtcConfig
        genesisTX = test.Tx0hash
        log.Printf("Running regtest mode with -tx=%s\n", genesisTX)
	} else {
		mainClient = conf.GetRPC(MAIN_NAME)
		oceanClient = conf.GetRPC(OCEAN_NAME)
		mainChainCfg = conf.GetChainCfgParams(MAIN_NAME)
	}

    apiHost = os.Getenv("API_HOST")
    if apiHost == "" {
        apiHost = DEFAULT_API_HOST
    }
}

func main() {
	defer mainClient.Shutdown()
	defer oceanClient.Shutdown()

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	channel := models.NewChannel()
	requestService := requestapi.NewRequestService(ctx, wg, channel, apiHost)
	attestService := attestation.NewAttestService(ctx, wg, channel, mainClient, oceanClient, mainChainCfg, genesisTX)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	wg.Add(1)
	go func() {
		defer cancel()
		defer wg.Done()
		select {
		case sig := <-c:
			log.Printf("Got %s signal. Aborting...\n", sig)
		case <-ctx.Done():
			signal.Stop(c)
		}
	}()

    if isRegtest { // In regtest demo mode generate main client blocks automatically
        wg.Add(1)
        go func() {
            waitTime := time.Now()
            defer wg.Done()
            for {
                select {
                    case <-ctx.Done():
                        return
                    default:
                        if time.Since(waitTime).Seconds() > 60 {
                            mainClient.Generate(1)
                            waitTime = time.Now()
                        }
                }
            }
        }()
    }

	wg.Add(1)
	go requestService.Run()
	wg.Add(1)
	go attestService.Run()
	wg.Wait()
}