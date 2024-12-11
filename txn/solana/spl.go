package sol

import (
	"bytes"
	"math/big"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
)

func SplApproveBody(senderAddr string, tokenAddr string, spenderAddr string, amount *big.Int, decimals int32) ([]byte, error) {
	senderAddr = strings.TrimSpace(senderAddr)
	tokenAddr = strings.TrimSpace(tokenAddr)
	spenderAddr = strings.TrimSpace(spenderAddr)

	senderpk, err := solana.PublicKeyFromBase58(senderAddr)
	if err != nil {
		return nil, err
	}
	mintpk, err := solana.PublicKeyFromBase58(tokenAddr)
	if err != nil {
		return nil, err
	}
	spenderpk, err := solana.PublicKeyFromBase58(spenderAddr)
	if err != nil {
		return nil, err
	}

	senderAta, err := GetAtaFromPk(senderpk, mintpk)
	if err != nil {
		return nil, err
	}

	spenderAta, err := GetAtaFromPk(spenderpk, mintpk)
	if err != nil {
		return nil, err
	}

	inst := token.NewApproveCheckedInstruction(
		amount.Uint64(),
		uint8(decimals),
		senderAta,
		mintpk,
		spenderAta,
		senderpk,
		[]solana.PublicKey{},
	).Build()

	return ToBody([]solana.Instruction{inst}, nil)

}

func Spl2022ApproveBody(senderAddr string, tokenAddr string, spenderAddr string, amount *big.Int, decimals int32) ([]byte, error) {
	senderAddr = strings.TrimSpace(senderAddr)
	tokenAddr = strings.TrimSpace(tokenAddr)
	spenderAddr = strings.TrimSpace(spenderAddr)

	senderpk, err := solana.PublicKeyFromBase58(senderAddr)
	if err != nil {
		return nil, err
	}
	mintpk, err := solana.PublicKeyFromBase58(tokenAddr)
	if err != nil {
		return nil, err
	}
	spenderpk, err := solana.PublicKeyFromBase58(spenderAddr)
	if err != nil {
		return nil, err
	}

	senderAta, err := Get2022AtaFromPk(senderpk, mintpk)
	if err != nil {
		return nil, err
	}

	spenderAta, err := Get2022AtaFromPk(spenderpk, mintpk)
	if err != nil {
		return nil, err
	}

	inst := token.NewApproveCheckedInstruction(
		amount.Uint64(),
		uint8(decimals),
		senderAta,
		mintpk,
		spenderAta,
		senderpk,
		[]solana.PublicKey{},
	).Build()
	data, err := ToBody([]solana.Instruction{inst}, nil)
	return bytes.ReplaceAll(data, []byte("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"), []byte("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")), err
}

func SqlTransferBody(senderAddr string, tokenAddr string, receiverAddr string, amount *big.Int, decimals int32) ([]byte, error) {
	senderAddr = strings.TrimSpace(senderAddr)
	tokenAddr = strings.TrimSpace(tokenAddr)
	receiverAddr = strings.TrimSpace(receiverAddr)

	senderpk, err := solana.PublicKeyFromBase58(senderAddr)
	if err != nil {
		return nil, err
	}
	mintpk, err := solana.PublicKeyFromBase58(tokenAddr)
	if err != nil {
		return nil, err
	}
	receiverpk, err := solana.PublicKeyFromBase58(receiverAddr)
	if err != nil {
		return nil, err
	}

	senderAta, err := GetAtaFromPk(senderpk, mintpk)
	if err != nil {
		return nil, err
	}

	receiverAta, err := GetAtaFromPk(receiverpk, mintpk)
	if err != nil {
		return nil, err
	}

	if decimals >= 0 {
		inst := token.NewTransferCheckedInstruction(
			amount.Uint64(),
			uint8(decimals),
			senderAta,
			mintpk,
			receiverAta,
			senderpk,
			[]solana.PublicKey{},
		).Build()
		return ToBody([]solana.Instruction{inst}, nil)
	} else {
		inst := token.NewTransferInstruction(
			amount.Uint64(),
			senderAta,
			receiverAta,
			senderpk,
			[]solana.PublicKey{},
		).Build()
		return ToBody([]solana.Instruction{inst}, nil)
	}

}

func Sql2022TransferBody(senderAddr string, tokenAddr string, receiverAddr string, amount *big.Int, decimals int32) ([]byte, error) {
	senderAddr = strings.TrimSpace(senderAddr)
	tokenAddr = strings.TrimSpace(tokenAddr)
	receiverAddr = strings.TrimSpace(receiverAddr)

	senderpk, err := solana.PublicKeyFromBase58(senderAddr)
	if err != nil {
		return nil, err
	}
	mintpk, err := solana.PublicKeyFromBase58(tokenAddr)
	if err != nil {
		return nil, err
	}
	receiverpk, err := solana.PublicKeyFromBase58(receiverAddr)
	if err != nil {
		return nil, err
	}

	senderAta, err := Get2022AtaFromPk(senderpk, mintpk)
	if err != nil {
		return nil, err
	}

	receiverAta, err := Get2022AtaFromPk(receiverpk, mintpk)
	if err != nil {
		return nil, err
	}

	if decimals >= 0 {
		inst := token.NewTransferCheckedInstruction(
			amount.Uint64(),
			uint8(decimals),
			senderAta,
			mintpk,
			receiverAta,
			senderpk,
			[]solana.PublicKey{},
		).Build()
		data, err := ToBody([]solana.Instruction{inst}, nil)
		return bytes.ReplaceAll(data, []byte("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"), []byte("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")), err
	} else {
		inst := token.NewTransferInstruction(
			amount.Uint64(),
			senderAta,
			receiverAta,
			senderpk,
			[]solana.PublicKey{},
		).Build()
		data, err := ToBody([]solana.Instruction{inst}, nil)
		return bytes.ReplaceAll(data, []byte("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"), []byte("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")), err
	}

}
