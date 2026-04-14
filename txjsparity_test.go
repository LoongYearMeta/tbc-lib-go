package bt_test

import (
	"math"
	"testing"

	"github.com/sCrypt-Inc/go-bt/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Golden: tbc-lib-js Transaction._estimateSize — 1×P2PKH 输入 180B，1×P2PKH 输出 8+1+25。
func TestJSEstimateSize_oneP2PKHInputOneP2PKHOutput(t *testing.T) {
	tx := bt.NewTx()
	require.NoError(t, tx.From(
		"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
		0,
		"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
		4000000,
	))
	require.NoError(t, tx.PayToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", 1000))
	// 8 + vin(1) + vout(1) + 180 + (8+1+25)
	assert.Equal(t, 224, tx.JSEstimateSize())
}

// Golden: change() 隐式手续费 = ceil(JSEstimateSizeWithPendingChange * satPerKB / bytesPer)，
// 此处仅 1 输入、将新增 1 笔找零，estimate 含找零脚本（与 JS 临时 0 sat 输出一致）。
func TestChangeToAddress_implicitFee_matchesCeilJSFormula(t *testing.T) {
	tx := bt.NewTx()
	require.NoError(t, tx.From(
		"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
		0,
		"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
		4000000,
	))
	require.NoError(t, tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", bt.NewFeeQuote()))

	std, err := bt.NewFeeQuote().Fee(bt.FeeTypeStandard)
	require.NoError(t, err)
	sat, bytesPer := std.MiningFee.Satoshis, std.MiningFee.Bytes
	// 仅输入、无输出时 JSEstimateSize 不含找零；找零路径下内部用带 extra 脚本的估算，等价于 224B → fee 112 @ 5/10
	wantEst := 224
	wantFee := uint64(math.Ceil(float64(wantEst) * float64(sat) / float64(bytesPer)))

	fee := tx.TotalInputSatoshis() - tx.TotalOutputSatoshis()
	assert.Equal(t, wantFee, fee)
	assert.Equal(t, uint64(4000000)-wantFee, tx.Outputs[0].Satoshis)
}

func TestAdjustImplicitFeeToTarget(t *testing.T) {
	tx := bt.NewTx()
	require.NoError(t, tx.From(
		"a4c76f8a7c05a91dcf5699b95b54e856298e50c1ceca9a8a5569c8532c500c11",
		0, "76a91455b61be43392125d127f1780fb038437cd67ef9c88ac", 1000,
	))
	require.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 800))
	require.NoError(t, tx.AdjustImplicitFeeToTarget(50))
	assert.Equal(t, uint64(50), tx.TotalInputSatoshis()-tx.TotalOutputSatoshis())
	assert.Equal(t, uint64(950), tx.Outputs[len(tx.Outputs)-1].Satoshis)
}
