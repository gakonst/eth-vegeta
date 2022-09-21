package main

import (
    "os"
    "bufio"
    "fmt"
    "time"

  vegeta "github.com/tsenart/vegeta/v12/lib"
)

// Load the RLP-encoded transaction which were produced
// by the binary.
func readTxs(path string) []string {
    readFile, err := os.Open(path)
    if err != nil {
        panic(err)
    }
    defer readFile.Close()

    fileScanner := bufio.NewScanner(readFile)
    fileScanner.Split(bufio.ScanLines)

    txs := []string{}
    for fileScanner.Scan() {
        tx := fileScanner.Text()
        txs = append(txs, tx)
    }

    return txs
}

func NewEthSendRawTransactionTargeter() vegeta.Targeter {
    // Read file with all the RLP encoded transactions
    fname := "./txs"
    txs := readTxs(fname)
    i := 0

    return func(tgt *vegeta.Target) error {
        if tgt == nil {
            return vegeta.ErrNilTarget
        }

        tgt.Method = "POST"
        tgt.URL = "http://localhost:8545"

        var header = map[string][]string{"Content-type":[]string{"application/json"}}
        tgt.Header = header


        // pick one tx from the list
        tx := txs[i % len(txs)]

        payload := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["%s"],"id":1}`, tx)
        // payload := `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`
        tgt.Body = []byte(payload)
        fmt.Println("Sending", payload)
        return nil
    }
}

func file(name string, create bool) (*os.File, error) {
	switch name {
	case "stdin":
		return os.Stdin, nil
	case "stdout":
		return os.Stdout, nil
	default:
		if create {
			return os.Create(name)
		}
		return os.Open(name)
	}
}

func main() {
  rate := vegeta.Rate{Freq: 1000, Per: time.Second}
  duration := 10 * time.Second
  targeter := NewEthSendRawTransactionTargeter();
  attacker := vegeta.NewAttacker()

  out, err := file("attack-out", true)
  if err != nil {
      panic(err)
  }
  defer out.Close()

  enc := vegeta.NewEncoder(out)

  for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
      if err = enc.Encode(res); err != nil {
          panic(err)
      }
  }
}
