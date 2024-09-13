package btc

import (
	"encoding/json"
	"math/big"
	"strings"
)

func TransferBody(receiverAddr string, amount *big.Int) ([]byte, error) {
	receiverAddr = strings.TrimSpace(receiverAddr)
	data := map[string]interface{}{
		"amount":   amount.Int64(),
		"receiver": receiverAddr,
	}
	return json.Marshal(data)
}
