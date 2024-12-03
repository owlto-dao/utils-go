package util

import "github.com/gagliardetto/solana-go"

type WrappedInstruction struct {
	program solana.PublicKey
	inner   solana.Instruction
}

func NewWrappedInstruction(program solana.PublicKey, instruction solana.Instruction) *WrappedInstruction {
	return &WrappedInstruction{
		program: program,
		inner:   instruction,
	}
}

func (instruction *WrappedInstruction) ProgramID() solana.PublicKey {
	return instruction.program
}

func (instruction *WrappedInstruction) Accounts() []*solana.AccountMeta {
	return instruction.inner.Accounts()
}

func (instruction *WrappedInstruction) Data() ([]byte, error) {
	return instruction.inner.Data()
}
