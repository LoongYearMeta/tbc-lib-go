package transaction_test

import "github.com/LoongYearMeta/tbc-lib-go/transaction"

// legacyDefaultFeeQuote returns the historical pre-100-sat/KB default fee
// quote (5 sat / 10 bytes = 500 sat/KB). Many golden tests in this package
// bake in raw transaction hex / change-output amounts / fee assertions that
// were computed under that old default. Since the project default has since
// been aligned with tbc-lib-js Transaction.FEE_PER_KB = 100, those golden
// tests now opt into the legacy rate explicitly via this helper to preserve
// their byte-for-byte expectations.
func legacyDefaultFeeQuote() *transaction.FeeQuote {
	q := transaction.NewFeeQuote()
	q.AddQuote(transaction.FeeTypeStandard, &transaction.Fee{
		FeeType:   transaction.FeeTypeStandard,
		MiningFee: transaction.FeeUnit{Satoshis: 5, Bytes: 10},
		RelayFee:  transaction.FeeUnit{Satoshis: 5, Bytes: 10},
	})
	q.AddQuote(transaction.FeeTypeData, &transaction.Fee{
		FeeType:   transaction.FeeTypeData,
		MiningFee: transaction.FeeUnit{Satoshis: 5, Bytes: 10},
		RelayFee:  transaction.FeeUnit{Satoshis: 5, Bytes: 10},
	})
	return q
}
