// Package merge 提供 UTXO 合并等高级 API，对应 tbc-lib-js 的 API.mergeUTXO 和 API.fetchUTXO。
// 使用方式：import "github.com/sCrypt-Inc/go-bt/v2/merge"
package merge

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"github.com/libsv/go-bk/bec"
	"github.com/sCrypt-Inc/go-bt/v2"
	"github.com/sCrypt-Inc/go-bt/v2/bscript"
	"github.com/sCrypt-Inc/go-bt/v2/unlocker"
)

// MergeUTXO 合并地址下的所有 UTXO 为一个
// 对应 JS API.mergeUTXO(privateKey, network)
func MergeUTXO(privKey *bec.PrivateKey, network string) error {
	mainnet := network != "testnet"
	addr, err := bscript.NewAddressFromPublicKey(privKey.PubKey(), mainnet)
	if err != nil {
		return err
	}
	address := addr.AddressString

	utxos, err := bt.FetchUTXOs(address, network)
	if err != nil {
		return err
	}
	if len(utxos) == 0 {
		return fmt.Errorf("no UTXO available")
	}
	if len(utxos) == 1 {
		return nil
	}

	var sumAmount uint64
	for _, u := range utxos {
		sumAmount += u.Satoshis
	}

	tx := bt.NewTx()
	if err := tx.FromUTXOs(utxos...); err != nil {
		return err
	}

	txSize, err := tx.EstimateSize()
	if err != nil {
		return err
	}
	txSize += 100
	fee := uint64(80)
	if txSize >= 1000 {
		fee = uint64(math.Ceil(float64(txSize) / 1000 * 80))
	}

	changeAmount := sumAmount - fee
	if changeAmount < 1 {
		return fmt.Errorf("insufficient amount for fee")
	}

	if err := tx.PayToAddress(address, changeAmount); err != nil {
		return err
	}

	ctx := context.Background()
	if err := tx.FillAllInputs(ctx, &unlocker.Getter{PrivateKey: privKey}); err != nil {
		return err
	}

	txraw := hex.EncodeToString(tx.Bytes())
	if _, err := bt.BroadcastTXRaw(txraw, network); err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	return MergeUTXO(privKey, network)
}

// FetchUTXOWithPrivateKey 从私钥对应地址获取满足 amount 的 UTXO，不足时自动 Merge 后重试
// 对应 JS API.fetchUTXO(privateKey, amount, network)
func FetchUTXOWithPrivateKey(privKey *bec.PrivateKey, amountTBC float64, network string) (*bt.UTXO, error) {
	mainnet := network != "testnet"
	addr, err := bscript.NewAddressFromPublicKey(privKey.PubKey(), mainnet)
	if err != nil {
		return nil, err
	}
	address := addr.AddressString

	utxos, err := bt.FetchUTXOs(address, network)
	if err != nil {
		return nil, err
	}
	if len(utxos) == 0 {
		return nil, fmt.Errorf("the tbc balance in the account is zero")
	}

	amountSatoshis := uint64(amountTBC * 1e6)

	if len(utxos) == 1 {
		if utxos[0].Satoshis > amountSatoshis {
			return utxos[0], nil
		}
		return nil, fmt.Errorf("insufficient tbc balance")
	}

	var selected *bt.UTXO
	for _, u := range utxos {
		if u.Satoshis > amountSatoshis {
			selected = u
			break
		}
	}
	if selected == nil {
		selected = utxos[0]
	}

	if selected.Satoshis < amountSatoshis {
		totalBalance, err := bt.GetTBCBalance(address, network)
		if err != nil {
			return nil, err
		}
		if totalBalance <= amountSatoshis {
			return nil, fmt.Errorf("insufficient tbc balance")
		}
		if err := MergeUTXO(privKey, network); err != nil {
			return nil, err
		}
		time.Sleep(3 * time.Second)
		return FetchUTXOWithPrivateKey(privKey, amountTBC, network)
	}

	return selected, nil
}
