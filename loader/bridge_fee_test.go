package loader

import (
	"math/big"
	"testing"
)

func TestUSDCBridgeFee(t *testing.T) {
	t.Log("test sol...")
	mgr := NewBridgeFeeManager(nil, nil)

	t.Log(mgr.FromUiString(big.NewInt(1000000), 21111111, 6, 2))

}
