package util

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/NethermindEth/starknet.go/curve"
	"github.com/ethereum/go-ethereum/common"
	"github.com/owlto-dao/utils-go/owlconsts"
)

func IsNativeAddress(address string) bool {
	return IsHexStringZero(address) || address == owlconsts.SolanaZeroAddress || address == owlconsts.BFCZeroAddress ||
		address == owlconsts.ZksyncEraAddress
}

func GetChecksumAddress(address string) (string, error) {
	address = strings.TrimSpace(address)
	if len(address) == 0 {
		return "", fmt.Errorf("empty address")
	}
	if !strings.HasPrefix(address, "0x") && !strings.HasPrefix(address, "0X") {
		return address, nil
	}
	if len(address) == 42 {
		return GetChecksumAddress40(address)
	} else if len(address) == 66 {
		return GetChecksumAddress64(address)
	} else {
		return "", fmt.Errorf("unsupport address len: %s", address)
	}
}

func GetFuelChecksumAddress(address string) (string, error) {
	// 检查地址是否符合 B256 格式
	if !isB256(address) {
		return "", fmt.Errorf("invalid B256 address format")
	}

	address = strings.TrimSpace(address)
	if len(address) == 0 {
		return "", fmt.Errorf("empty address")
	}
	if len(address) == 66 {
		return toFuelChecksumAddress(address), nil
	} else {
		return "", fmt.Errorf("unsupport address len: %s", address)
	}
}

func isB256(address string) bool {
	matched, _ := regexp.MatchString(`^(0x)[0-9a-fA-F]{64}$`, address)
	return len(address) == 66 && matched
}

func toFuelChecksumAddress(address string) string {
	// 转换地址为小写，并去掉前缀 "0x"
	addressHex := strings.ToLower(strings.TrimPrefix(address, "0x"))

	// 计算地址的 SHA-256 哈希
	hasher := sha256.New()
	hasher.Write([]byte(addressHex))
	checksum := hasher.Sum(nil)

	// 根据哈希值设置字符的大小写
	ret := "0x"
	for i := 0; i < 32; i++ {
		ha := addressHex[i*2]
		hb := addressHex[i*2+1]
		// 根据哈希的高 4 位和低 4 位设置字符大小写
		if checksum[i]&0xf0 >= 0x80 {
			ret += strings.ToUpper(string(ha))
		} else {
			ret += string(ha)
		}
		if checksum[i]&0x0f >= 0x08 {
			ret += strings.ToUpper(string(hb))
		} else {
			ret += string(hb)
		}
	}

	return ret
}

func GetChecksumAddress40(address string) (string, error) {
	address = strings.TrimSpace(address)
	return common.HexToAddress(address).Hex(), nil
}

func GetChecksumAddress64(address string) (string, error) {
	address = strings.TrimSpace(address)
	address = strings.TrimLeft(strings.TrimPrefix(strings.ToLower(address), "0x"), "0")
	if len(address)%2 != 0 {
		address = "0" + address
	}
	if len(address) > 64 {
		return "", errors.New("address too long")
	}
	address64 := strings.Repeat("0", 64-len(address)) + address
	chars := strings.Split(address64, "")
	byteSlice, err := hex.DecodeString(address)
	if err != nil {
		return "", err
	}
	h, err := curve.Curve.StarknetKeccak(byteSlice)
	if err != nil {
		return "", err
	}
	hs := strings.TrimLeft(strings.TrimPrefix(h.String(), "0x"), "0")
	if len(hs) > 64 {
		return "", errors.New("hs too long")
	}
	hashed, err := hex.DecodeString(strings.Repeat("0", 64-len(hs)) + hs)
	if err != nil {
		return "", err
	}

	for i := 0; i < len(chars); i += 2 {
		if hashed[i>>1]>>4 >= 8 {
			chars[i] = strings.ToUpper(chars[i])
		}
		if (hashed[i>>1] & 0x0f) >= 8 {
			chars[i+1] = strings.ToUpper(chars[i+1])
		}
	}
	return "0x" + strings.Join(chars, ""), nil
}
