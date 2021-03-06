// Package test implements unit test and regtest utitilies for attestation.
package test

import (
    "os/exec"
    "os"
    "log"

    "mainstay/config"

    "github.com/btcsuite/btcd/btcjson"
)

// For regtest attestation demonstration
const DEMO_INIT_PATH = "/src/mainstay/test/demo-init.sh"

// For unit-testing
const TEST_INIT_PATH = "/src/mainstay/test/test-init.sh"

var testConf = []byte(`
{
    "main": {
        "rpcurl": "localhost:18443",
        "rpcuser": "user",
        "rpcpass": "pass",
        "chain": "regtest"
    },
    "ocean": {
        "rpcurl": "localhost:18010",
        "rpcuser": "bitcoinrpc",
        "rpcpass": "acc1e7a299bc49449912e235b54dbce5",
        "chain": "main"
    },
    "misc": {
        "multisignodes": "127.0.0.1:5001"
    }
}
`)

// Test structure
// Set up testing environment for use by regtest demo or unit tests
type Test struct {
    Config      *config.Config
}

// NewTest returns a pointer to a Test instance
func NewTest(logOutput bool, isRegtest bool) *Test {
    // Run init test script that sets up bitcoin and ocean
    var initPath string
    if (isRegtest) { // for running the demon in regtest mode along with ocean demo
        initPath = os.Getenv("GOPATH") + DEMO_INIT_PATH
    } else { // for running unit tests
        initPath = os.Getenv("GOPATH") + TEST_INIT_PATH
    }

    cmd := exec.Command("/bin/sh", initPath)
    output, err := cmd.Output()
    if err != nil {
        log.Fatal(err)
    }
    if (logOutput) {
        log.Println(string(output))
    }

    // if not a regtest, then unittest
    config := config.NewConfig(!isRegtest, testConf)

    /*
    addr1 := "2N9z6a8BQB1xWmesCJcBWZm1R3f1PZcwrGz"
    pub1 := "03e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b33"
    priv1 := "cQca2KvrBnJJUCYa2tD4RXhiQshWLNMSK2A96ZKWo1SZkHhh3YLz"

    addr2 := "2MyC1i1FGy6MZWyMgmZXku4gdWZxWCRa6RL"
    pub2 := "02f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e75"
    priv2 := "cSS9R4XPpajhqy28hcfHEzEzAbyWDqBaGZR4xtV7Jg8TixSWee1x"

    address := "2MxBi6eodnuoVCw8McGrf1nuoVhastqoBXB"
    script := "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b332102f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e7552ae"
    */
    priv1 := "cQca2KvrBnJJUCYa2tD4RXhiQshWLNMSK2A96ZKWo1SZkHhh3YLz"
    script := "512103e52cf15e0a5cf6612314f077bb65cf9a6596b76c0fcb34b682f673a8314c7b332102f3a78a7bd6cf01c56312e7e828bef74134dfb109e59afd088526212d96518e7552ae"

    // Get first unspent as initial TX for attestation chain
    unspent, errUnspent := config.MainClient().ListUnspent()
    if errUnspent != nil {
        log.Fatal(errUnspent)
    }
    var tx0 btcjson.ListUnspentResult
    for _, vout := range unspent {
        if (vout.Amount > 50) { // skip regtest txs
            tx0 = vout
        }
    }

    config.SetInitTX(tx0.TxID)
    config.SetInitPK(priv1)
    config.SetMultisigScript(script)

    return &Test{config}
}
