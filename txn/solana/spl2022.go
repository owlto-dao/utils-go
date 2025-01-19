package sol

import (
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	solana_token "github.com/gagliardetto/solana-go/programs/token"
)

type Spl2022CreateAtaInstruction struct {
	token                   solana.PublicKey
	recipient               solana.PublicKey
	splCreateAtaInstruction solana.Instruction
}

func (spl2022 *Spl2022CreateAtaInstruction) ProgramID() solana.PublicKey {
	return spl2022.splCreateAtaInstruction.ProgramID()
}

func (spl2022 *Spl2022CreateAtaInstruction) Accounts() []*solana.AccountMeta {
	fakeAtaAccount, _, _ := solana.FindAssociatedTokenAddress(spl2022.recipient, spl2022.token)
	accounts := spl2022.splCreateAtaInstruction.Accounts()
	for _, account := range accounts {
		if account.PublicKey.Equals(solana.TokenProgramID) {
			account.PublicKey = solana.MustPublicKeyFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")
		} else if account.PublicKey.Equals(fakeAtaAccount) {
			spl2022AtaAccount, _ := Get2022AtaFromPk(spl2022.recipient, spl2022.token)
			account.PublicKey = spl2022AtaAccount
		}
	}
	return accounts
}

func (spl2022 *Spl2022CreateAtaInstruction) Data() ([]byte, error) {
	return spl2022.splCreateAtaInstruction.Data()
}

func NewSpl2022CreateAtaInstruction(payer solana.PublicKey, wallet solana.PublicKey, token solana.PublicKey) *Spl2022CreateAtaInstruction {
	return &Spl2022CreateAtaInstruction{
		token:                   token,
		recipient:               wallet,
		splCreateAtaInstruction: associatedtokenaccount.NewCreateInstruction(payer, wallet, token).Build(),
	}
}

type Spl2022TransferInstruction struct {
	splTransferInstruction solana.Instruction
}

func (spl2022 *Spl2022TransferInstruction) ProgramID() solana.PublicKey {
	return solana.MustPublicKeyFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")
}

func (spl2022 *Spl2022TransferInstruction) Accounts() []*solana.AccountMeta {
	return spl2022.splTransferInstruction.Accounts()
}

func (spl2022 *Spl2022TransferInstruction) Data() ([]byte, error) {
	return spl2022.splTransferInstruction.Data()
}

func NewSpl2022TransferInstruction(amount uint64, source solana.PublicKey, destination solana.PublicKey, owner solana.PublicKey) *Spl2022TransferInstruction {
	return &Spl2022TransferInstruction{
		splTransferInstruction: solana_token.NewTransferInstruction(amount, source, destination, owner, []solana.PublicKey{}).Build(),
	}
}
