package transaction_test

import (
	"math"
	"testing"

	"github.com/LoongYearMeta/tbc-lib-go/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	feeSafetyAddress = "mwV3YgnowbJJB3LcyCuqiKpdivvNNFiK7M"
	feeSafetyScript  = "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac"
	expectedSDKDust  = uint64(42)
)

func TestChangeToAddressEnforcesMinimumTransactionFee(t *testing.T) {
	tx := transaction.NewTx()
	require.NoError(t, tx.From(
		"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
		0,
		feeSafetyScript,
		1000,
	))

	require.NoError(t, tx.ChangeToAddress(feeSafetyAddress, transaction.NewFeeQuote()))

	require.Equal(t, 1, tx.OutputCount())
	assert.Equal(t, uint64(1000)-transaction.MinimumTransactionFee, tx.Outputs[0].Satoshis)
	assert.Equal(t, transaction.MinimumTransactionFee,
		tx.TotalInputSatoshis()-tx.TotalOutputSatoshis())
}

func feeSafetyQuote() *transaction.FeeQuote {
	q := transaction.NewFeeQuote()
	q.AddQuote(transaction.FeeTypeStandard, &transaction.Fee{
		FeeType:   transaction.FeeTypeStandard,
		MiningFee: transaction.FeeUnit{Satoshis: 1000, Bytes: 1000},
		RelayFee:  transaction.FeeUnit{Satoshis: 1000, Bytes: 1000},
	})
	return q
}

func feeSafetyTx(t *testing.T, available uint64) *transaction.Tx {
	t.Helper()

	tx := transaction.NewTx()
	require.NoError(t, tx.From(
		"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
		0,
		feeSafetyScript,
		1000,
	))
	require.NoError(t, tx.PayToAddress(feeSafetyAddress, 1000-available))
	return tx
}

func TestChangeToAddressFeeAndDustBoundaries(t *testing.T) {
	// 1 P2PKH input + 1 payment + 1 prospective change output estimates to
	// 258 bytes. At 1000 sat/kB the target fee is therefore exactly 258 sat.
	const targetFee uint64 = 258

	t.Run("one satoshi below target fee is rejected", func(t *testing.T) {
		tx := feeSafetyTx(t, targetFee-1)

		err := tx.ChangeToAddress(feeSafetyAddress, feeSafetyQuote())

		require.ErrorIs(t, err, transaction.ErrInsufficientInputs)
		assert.Equal(t, 1, tx.OutputCount())
	})

	t.Run("exact target fee succeeds without change", func(t *testing.T) {
		tx := feeSafetyTx(t, targetFee)

		require.NoError(t, tx.ChangeToAddress(feeSafetyAddress, feeSafetyQuote()))

		assert.Equal(t, 1, tx.OutputCount())
		assert.Equal(t, targetFee, tx.TotalInputSatoshis()-tx.TotalOutputSatoshis())
	})

	t.Run("remainder below sdk dust is donated to fee", func(t *testing.T) {
		tx := feeSafetyTx(t, targetFee+expectedSDKDust-1)

		require.NoError(t, tx.ChangeToAddress(feeSafetyAddress, feeSafetyQuote()))

		assert.Equal(t, 1, tx.OutputCount())
		assert.Equal(t, targetFee+expectedSDKDust-1, tx.TotalInputSatoshis()-tx.TotalOutputSatoshis())
	})

	t.Run("remainder at sdk dust creates change", func(t *testing.T) {
		tx := feeSafetyTx(t, targetFee+expectedSDKDust)

		require.NoError(t, tx.ChangeToAddress(feeSafetyAddress, feeSafetyQuote()))

		require.Equal(t, 2, tx.OutputCount())
		assert.Equal(t, expectedSDKDust, tx.Outputs[1].Satoshis)
		assert.Equal(t, targetFee, tx.TotalInputSatoshis()-tx.TotalOutputSatoshis())
	})
}

func TestChangeToAddressRejectsAmountSumOverflow(t *testing.T) {
	t.Run("input sum overflow", func(t *testing.T) {
		tx := transaction.NewTx()
		require.NoError(t, tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			feeSafetyScript,
			math.MaxUint64,
		))
		require.NoError(t, tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			1,
			feeSafetyScript,
			math.MaxUint64,
		))
		require.NoError(t, tx.PayToAddress(feeSafetyAddress, 42))

		err := tx.ChangeToAddress(feeSafetyAddress, feeSafetyQuote())

		assert.ErrorIs(t, err, transaction.ErrAmountOverflow)
	})

	t.Run("output sum overflow", func(t *testing.T) {
		tx := transaction.NewTx()
		require.NoError(t, tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			feeSafetyScript,
			math.MaxUint64,
		))
		require.NoError(t, tx.PayToAddress(feeSafetyAddress, math.MaxUint64))
		require.NoError(t, tx.PayToAddress(feeSafetyAddress, 1))

		err := tx.ChangeToAddress(feeSafetyAddress, feeSafetyQuote())

		assert.ErrorIs(t, err, transaction.ErrAmountOverflow)
	})
}
