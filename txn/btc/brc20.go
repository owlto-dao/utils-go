package btc

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/owlto-dao/utils-go/convert"
)

func BRC20TransferBody(receiverAddr string, tokenName string, amount *big.Int) ([]byte, error) {
	receiverAddr = strings.TrimSpace(receiverAddr)
	data := map[string]interface{}{
		"method":   "transfer",
		"token":    tokenName,
		"amount":   amount.Int64(),
		"receiver": receiverAddr,
	}
	dataStr := convert.ConvertToJsonString(data)
	m := map[string]interface{}{
		"tx_type": "BRC20",
		"data":    dataStr,
	}
	return json.Marshal(m)
}
