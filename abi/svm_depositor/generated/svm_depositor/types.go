// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package svm_depositor

import ag_binary "github.com/gagliardetto/binary"

type TransferData struct {
	Amount      uint64
	TargetAddr  string
	Maker       [32]uint8
	Token       [32]uint8
	Destination uint32
	Channel     uint32
	Extra       string
}

func (obj TransferData) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Amount` param:
	err = encoder.Encode(obj.Amount)
	if err != nil {
		return err
	}
	// Serialize `TargetAddr` param:
	err = encoder.Encode(obj.TargetAddr)
	if err != nil {
		return err
	}
	// Serialize `Maker` param:
	err = encoder.Encode(obj.Maker)
	if err != nil {
		return err
	}
	// Serialize `Token` param:
	err = encoder.Encode(obj.Token)
	if err != nil {
		return err
	}
	// Serialize `Destination` param:
	err = encoder.Encode(obj.Destination)
	if err != nil {
		return err
	}
	// Serialize `Channel` param:
	err = encoder.Encode(obj.Channel)
	if err != nil {
		return err
	}
	// Serialize `Extra` param:
	err = encoder.Encode(obj.Extra)
	if err != nil {
		return err
	}
	return nil
}

func (obj *TransferData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Amount`:
	err = decoder.Decode(&obj.Amount)
	if err != nil {
		return err
	}
	// Deserialize `TargetAddr`:
	err = decoder.Decode(&obj.TargetAddr)
	if err != nil {
		return err
	}
	// Deserialize `Maker`:
	err = decoder.Decode(&obj.Maker)
	if err != nil {
		return err
	}
	// Deserialize `Token`:
	err = decoder.Decode(&obj.Token)
	if err != nil {
		return err
	}
	// Deserialize `Destination`:
	err = decoder.Decode(&obj.Destination)
	if err != nil {
		return err
	}
	// Deserialize `Channel`:
	err = decoder.Decode(&obj.Channel)
	if err != nil {
		return err
	}
	// Deserialize `Extra`:
	err = decoder.Decode(&obj.Extra)
	if err != nil {
		return err
	}
	return nil
}