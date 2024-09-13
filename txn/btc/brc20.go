package btc

import (
	"encoding/json"
	"math/big"
	"strings"
)

func BRC20TransferBody(receiverAddr string, tokenName string, amount *big.Int) ([]byte, error) {
	receiverAddr = strings.TrimSpace(receiverAddr)
	data := map[string]interface{}{
		"method":   "transfer",
		"token":    tokenName,
		"amount":   amount.Int64(),
		"receiver": receiverAddr,
	}
	return json.Marshal(data)
}
