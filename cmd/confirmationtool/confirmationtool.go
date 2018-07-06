// Confirmation Tool
package main

import (
    "os"
    "log"
    "time"
    "flag"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcd/chaincfg"
    "ocean-attestation/conf"
    "ocean-attestation/staychain"
)

const MAIN_NAME     = "bitcoin"
const SIDE_NAME     = "ocean"
const CONF_PATH     = "/src/ocean-attestation/cmd/confirmationtool/conf.json"
const DEFAULT_TX    = "bf41c0da8047b1416d5ca680e2643967b27537cdf9a41527034698c336b55313"

var (
    tx              string
    showDetails     bool
    mainClient      *rpcclient.Client
    sideClient      *rpcclient.Client
    mainChainCfg    *chaincfg.Params
)

func init() {
    flag.BoolVar(&showDetails, "detailed", false, "Detailed information on attestation transaction")
    flag.StringVar(&tx, "tx", "", "Tx id for genesis attestation transaction")
    flag.Parse()
    if (tx == "") {
        tx = DEFAULT_TX
    }

    confFile := conf.GetConfFile(os.Getenv("GOPATH") + CONF_PATH)
    mainClient = conf.GetRPC(MAIN_NAME, confFile)
    sideClient = conf.GetRPC(SIDE_NAME, confFile)
    mainChainCfg = conf.GetChainCfgParams(MAIN_NAME, confFile)
}

func main() {
    defer mainClient.Shutdown()
    defer sideClient.Shutdown()

    txhash, errHash := chainhash.NewHashFromStr(tx)
    if errHash != nil {
        log.Fatal("Invalid tx id provided")
    }
    txraw, errGet := mainClient.GetRawTransactionVerbose(txhash)
    if errGet != nil {
        log.Fatal("Inititial transcaction does not exist")
    }

    fetcher := staychain.NewChainFetcher(mainClient, staychain.Tx(*txraw))
    chain := staychain.NewChain(fetcher)
    verifier := staychain.NewChainVerifier(mainChainCfg, sideClient)

    time.AfterFunc(5*time.Minute, func() {
        log.Println("Exit: ", chain.Close())
    })

    for tx := range chain.Updates() {
        log.Println("Verifying attestation")
        log.Printf("txid: %s\n", tx.Txid)
        info, err := verifier.Verify(tx)
        if err != nil {
            log.Println(err)
        } else {
            printAttestation(tx, info)
        }
    }
}

func printAttestation(tx staychain.Tx, info staychain.ChainVerifierInfo) {
    log.Println("Attestation Verified")
    if showDetails {
        log.Printf("%+v\n", tx)
    } else {
        log.Printf("%s blockhash: %s\n", MAIN_NAME, tx.BlockHash)
    }
    log.Printf("%s blockhash: %s\n", SIDE_NAME, info.Hash())
    log.Printf("%s blockheight: %d\n", SIDE_NAME, info.Height())
    log.Printf("\n")
}