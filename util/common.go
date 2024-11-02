package util

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

func IsHexStringZero(hexString string) bool {
	// Remove "0x" prefix if it exists
	hexString = strings.TrimSpace(hexString)
	if len(hexString) >= 2 && (hexString[:2] == "0x" || hexString[:2] == "0X") {
		hexString = hexString[2:]
	}

	// Check if all characters in the hex string are '0'
	for _, ch := range hexString {
		if ch != '0' {
			return false
		}
	}
	return true
}

func IsHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !IsHexCharacter(c) {
			return false
		}
	}
	return true
}

func IsHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

func GetJsonBigInt(itf interface{}) *big.Int {
	switch itf := itf.(type) {
	case float64:
		return big.NewInt(int64(itf))
	case string:
		bi := new(big.Int)
		bi, success := bi.SetString(strings.TrimSpace(itf), 0)
		if success {
			return bi
		} else {
			return new(big.Int)
		}
	default:
		return new(big.Int)
	}
}

func FromUiString(amount string, decimals int32) (*big.Int, error) {
	v, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, err
	}
	return v.Shift(decimals).BigInt(), nil
}

func FromUiFloat(amount float64, decimals int32) *big.Int {
	return decimal.NewFromFloat(amount).Shift(decimals).BigInt()
}

func StringToUi(amountStr string, decimals int32) (*big.Float, error) {
	v, err := decimal.NewFromString(amountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid amount string: %s", amountStr)
	}
	return v.Shift(-decimals).BigFloat(), nil
}

func BigIntToUi(amount *big.Int, decimals int32) *big.Float {
	return decimal.NewFromBigInt(amount, -decimals).BigFloat()
}

func IsEvmAddress(address string, chainID int32) bool {
	return common.IsHexAddress(address) && chainID != 666666666 && chainID != 83797601
}

func MaskEVMAddress(address string) string {
	if len(address) < 12 {
		return address
	}
	return address[:6] + "****" + address[len(address)-4:]
}
