package transaction_test

import (
	"testing"

	"github.com/LoongYearMeta/tbc-lib-go/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Golden: tbc-lib-js Transaction._estimateSize — 1×P2PKH 输入 180B，1×P2PKH 输出 8+1+25。
func TestJSEstimateSize_oneP2PKHInputOneP2PKHOutput(t *testing.T) {
	tx := transaction.NewTx()
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

// Golden: change() 隐式手续费 = ceil(estimateSize * feePerKb / 1000)（与 JS _estimateFee 一致），
// 此处仅 1 输入、将新增 1 笔找零，estimate 含找零脚本（与 JS 临时 0 sat 输出一致）。
func TestChangeToAddress_implicitFee_matchesCeilJSFormula(t *testing.T) {
	tx := transaction.NewTx()
	require.NoError(t, tx.From(
		"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
		0,
		"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
		4000000,
	))
	require.NoError(t, tx.ChangeToAddress("mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M", transaction.NewFeeQuote()))

	std, err := transaction.NewFeeQuote().Fee(transaction.FeeTypeStandard)
	require.NoError(t, err)
	// 224B，默认 MAPI 5 sat/10B → feePerKb=500 → ceil(224*500/1000)=112
	wantEst := 224
	wantFee := transaction.CeilMiningFeeFromEstimatedBytes(wantEst, std.MiningFee)

	fee := tx.TotalInputSatoshis() - tx.TotalOutputSatoshis()
	assert.Equal(t, wantFee, fee)
	assert.Equal(t, uint64(4000000)-wantFee, tx.Outputs[0].Satoshis)
}

// Golden: CeilMiningFeeFromEstimatedBytes 与 JS Math.ceil(size/1000 * feePerKb)；
// Transaction.FEE_PER_KB 默认 100 → MiningFee 100 sat / 1000 bytes。
func TestCeilMiningFeeFromEstimatedBytes_golden(t *testing.T) {
	assert.Equal(t, uint64(112), transaction.CeilMiningFeeFromEstimatedBytes(224, transaction.FeeUnit{Satoshis: 5, Bytes: 10}))
	assert.Equal(t, uint64(23), transaction.CeilMiningFeeFromEstimatedBytes(224, transaction.FeeUnit{Satoshis: 100, Bytes: 1000}))
	assert.Equal(t, uint64(1), transaction.CeilMiningFeeFromEstimatedBytes(1, transaction.FeeUnit{Satoshis: 100, Bytes: 1000}))
}

func TestAdjustImplicitFeeToTarget(t *testing.T) {
	tx := transaction.NewTx()
	require.NoError(t, tx.From(
		"a4c76f8a7c05a91dcf5699b95b54e856298e50c1ceca9a8a5569c8532c500c11",
		0, "76a91455b61be43392125d127f1780fb038437cd67ef9c88ac", 1000,
	))
	require.NoError(t, tx.AddP2PKHOutputFromAddress("mtestD3vRB7AoYWK2n6kLdZmAMLbLhDsLr", 800))
	require.NoError(t, tx.AdjustImplicitFeeToTarget(50))
	assert.Equal(t, uint64(50), tx.TotalInputSatoshis()-tx.TotalOutputSatoshis())
	assert.Equal(t, uint64(950), tx.Outputs[len(tx.Outputs)-1].Satoshis)
}
