package bt

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"sort"
	"strconv"

	"github.com/sCrypt-Inc/go-bt/v2/bscript"
)

// FtInfo FT 代币信息
type FtInfo struct {
	CodeScript  string `json:"codeScript"`
	TapeScript  string `json:"tapeScript"`
	TotalSupply string `json:"totalSupply"`
	Decimal     uint   `json:"decimal"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
}

// FtUTXO FT UTXO，扩展了 FtBalance
type FtUTXO struct {
	TxID       string `json:"txId"`
	Vout       uint32 `json:"outputIndex"`
	Script     string `json:"script"`
	Satoshis   uint64 `json:"satoshis"`
	FtBalance  string `json:"ftBalance"`
}

// ftBalanceResponse 用于解析可能包含大数的 JSON
type ftBalanceResponse struct {
	Data struct {
		Balance json.RawMessage `json:"balance"`
	} `json:"data"`
}

type ftUtxoRaw struct {
	TxID     string          `json:"txid"`
	Index    int             `json:"index"`
	TBCValue uint64          `json:"tbc_value"`
	FTValue  json.RawMessage `json:"ft_value"`
}

type ftUtxoListResponse struct {
	Data struct {
		UTXOs []ftUtxoRaw `json:"utxos"`
	} `json:"data"`
}

type ftInfoResponse struct {
	Data struct {
		CodeScript string          `json:"code_script"`
		TapeScript string          `json:"tape_script"`
		Amount     json.RawMessage `json:"amount"`
		Decimal    uint            `json:"decimal"`
		Name       string          `json:"name"`
		Symbol     string          `json:"symbol"`
	} `json:"data"`
}

// buildAddressOrHash 将 addressOrHash 转为 combinescript 路径段（与合约 OP_RETURN 中 21 字节 recipient 一致：mode||payload）。
func buildAddressOrHash(addressOrHash string) (string, error) {
	ok, _ := bscript.ValidateAddress(addressOrHash)
	if ok {
		addr, err := bscript.NewAddressFromString(addressOrHash)
		if err != nil {
			return "", err
		}
		return "00" + addr.PublicKeyHash, nil
	}
	if len(addressOrHash) == 40 && isHex(addressOrHash) {
		return "01" + addressOrHash, nil
	}
	return "", fmt.Errorf("Invalid address or hash")
}

// buildAddressOrHashLegacy 旧式布局（与 tbc-contract lib/api/api.ts 一致）：pkh||00、hash||01；多数索引仍按此键返回数据。
func buildAddressOrHashLegacy(addressOrHash string) (string, error) {
	ok, _ := bscript.ValidateAddress(addressOrHash)
	if ok {
		addr, err := bscript.NewAddressFromString(addressOrHash)
		if err != nil {
			return "", err
		}
		return addr.PublicKeyHash + "00", nil
	}
	if len(addressOrHash) == 40 && isHex(addressOrHash) {
		return addressOrHash + "01", nil
	}
	return "", fmt.Errorf("Invalid address or hash")
}

func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// enrichFtUtxoScriptsFromChain 用链上父交易对应输出的 locking script 与金额覆盖 FtUTXO.Script / Satoshis，
// 与 GetPreTxdata 及 BIP143 sighash 一致。
func enrichFtUtxoScriptsFromChain(list []*FtUTXO, network string) error {
	txCache := make(map[string]*Tx)
	for _, u := range list {
		if u == nil {
			continue
		}
		tx, ok := txCache[u.TxID]
		if !ok {
			var err error
			tx, err = FetchTXRaw(u.TxID, network)
			if err != nil {
				return fmt.Errorf("enrich FtUTXO script: fetch tx %s: %w", u.TxID, err)
			}
			txCache[u.TxID] = tx
		}
		if int(u.Vout) >= len(tx.Outputs) {
			return fmt.Errorf("enrich FtUTXO script: vout %d out of range for tx %s (outputs=%d)", u.Vout, u.TxID, len(tx.Outputs))
		}
		out := tx.Outputs[u.Vout]
		ls := out.LockingScript
		if ls == nil {
			return fmt.Errorf("enrich FtUTXO script: tx %s vout %d has nil locking script", u.TxID, u.Vout)
		}
		u.Script = hex.EncodeToString(ls.Bytes())
		u.Satoshis = out.Satoshis
	}
	return nil
}

// parseBigIntOrUint64 解析 JSON 中的大整数（可能是 number 或 string）
func parseBigIntOrUint64(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "0", nil
	}
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return "", err
		}
		return s, nil
	}
	var n uint64
	if err := json.Unmarshal(raw, &n); err != nil {
		var f float64
		if err2 := json.Unmarshal(raw, &f); err2 != nil {
			return "", err
		}
		n = uint64(f)
	}
	return strconv.FormatUint(n, 10), nil
}

func getFTBalanceByHash(contractTxID, hash, network string) (string, error) {
	baseURL := getBaseURL(network)
	url := fmt.Sprintf("%sft/tokenbalance/combinescript/%s/contract/%s", baseURL, hash, contractTxID)

	resp, err := defaultHTTPClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var r ftBalanceResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("解析 FT 余额响应失败: %w", err)
	}
	return parseBigIntOrUint64(r.Data.Balance)
}

func ftBalanceStringPositive(s string) bool {
	b, ok := new(big.Int).SetString(s, 10)
	return ok && b.Sign() > 0
}

// GetFTBalance 获取 FT 余额
// 对应 JS API.getFTbalance
func GetFTBalance(contractTxID, addressOrHash, network string) (string, error) {
	hash, err := buildAddressOrHash(addressOrHash)
	if err != nil {
		return "", err
	}
	bal, err := getFTBalanceByHash(contractTxID, hash, network)
	if err != nil {
		return "", err
	}
	if ftBalanceStringPositive(bal) {
		return bal, nil
	}
	hashLegacy, err2 := buildAddressOrHashLegacy(addressOrHash)
	if err2 != nil || hashLegacy == hash {
		return bal, nil
	}
	bal2, err3 := getFTBalanceByHash(contractTxID, hashLegacy, network)
	if err3 != nil {
		return bal, nil
	}
	return bal2, nil
}

func fetchFtUTXOListResponse(contractTxID, hash, network string) (ftUtxoListResponse, error) {
	baseURL := getBaseURL(network)
	url := fmt.Sprintf("%sft/utxo/combinescript/%s/contract/%s", baseURL, hash, contractTxID)

	resp, err := defaultHTTPClient.Get(url)
	if err != nil {
		return ftUtxoListResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ftUtxoListResponse{}, err
	}

	var r ftUtxoListResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return ftUtxoListResponse{}, fmt.Errorf("解析 FT UTXO 响应失败: %w", err)
	}
	return r, nil
}

// FetchFtUTXOList 获取 FT UTXO 列表
// 对应 JS API.fetchFtUTXOList
func FetchFtUTXOList(contractTxID, addressOrHash, codeScript, network string) ([]*FtUTXO, error) {
	hash, err := buildAddressOrHash(addressOrHash)
	if err != nil {
		return nil, err
	}
	r, err := fetchFtUTXOListResponse(contractTxID, hash, network)
	if err != nil {
		return nil, err
	}
	if len(r.Data.UTXOs) == 0 {
		hashLegacy, err2 := buildAddressOrHashLegacy(addressOrHash)
		if err2 == nil && hashLegacy != hash {
			r, err = fetchFtUTXOListResponse(contractTxID, hashLegacy, network)
			if err != nil {
				return nil, err
			}
		}
	}
	if len(r.Data.UTXOs) == 0 {
		return nil, fmt.Errorf("The ft balance in the account is zero.")
	}

	result := make([]*FtUTXO, 0, len(r.Data.UTXOs))
	for i := range r.Data.UTXOs {
		fv, err := parseBigIntOrUint64(r.Data.UTXOs[i].FTValue)
		if err != nil {
			return nil, err
		}
		result = append(result, &FtUTXO{
			TxID:      r.Data.UTXOs[i].TxID,
			Vout:      uint32(r.Data.UTXOs[i].Index),
			Script:    codeScript,
			Satoshis:  r.Data.UTXOs[i].TBCValue,
			FtBalance: fv,
		})
	}
	if err := enrichFtUtxoScriptsFromChain(result, network); err != nil {
		return nil, err
	}
	return result, nil
}

// FetchFtUTXO 获取满足 amount 的单个 FT UTXO
// 对应 JS API.fetchFtUTXO
func FetchFtUTXO(contractTxID, addressOrHash, codeScript, network string, amount *big.Int) (*FtUTXO, error) {
	list, err := FetchFtUTXOList(contractTxID, addressOrHash, codeScript, network)
	if err != nil {
		return nil, err
	}
	var selected *FtUTXO
	for _, u := range list {
		ub, _ := new(big.Int).SetString(u.FtBalance, 10)
		if ub != nil && ub.Cmp(amount) >= 0 {
			selected = u
			break
		}
	}
	if selected == nil {
		selected = list[0]
	}
	sb, _ := new(big.Int).SetString(selected.FtBalance, 10)
	if sb != nil && sb.Cmp(amount) < 0 {
		totalStr, err := GetFTBalance(contractTxID, addressOrHash, network)
		if err != nil {
			return nil, err
		}
		tb, _ := new(big.Int).SetString(totalStr, 10)
		if tb != nil && tb.Cmp(amount) >= 0 {
			return nil, fmt.Errorf("Insufficient FTbalance, please merge FT UTXOs")
		}
		return nil, fmt.Errorf("FTbalance not enough!")
	}
	return selected, nil
}

// FetchFtUTXOs 获取 FT UTXO 列表，按 ftBalance 降序，最多 5 个；若指定 amount 则保证合计 >= amount
// 对应 JS API.fetchFtUTXOs
func FetchFtUTXOs(contractTxID, addressOrHash, codeScript, network string, amount *big.Int) ([]*FtUTXO, error) {
	list, err := FetchFtUTXOList(contractTxID, addressOrHash, codeScript, network)
	if err != nil {
		return nil, err
	}
	// 按 ftBalance 降序
	sort.Slice(list, func(i, j int) bool {
		a, _ := new(big.Int).SetString(list[i].FtBalance, 10)
		b, _ := new(big.Int).SetString(list[j].FtBalance, 10)
		if a == nil {
			return true
		}
		if b == nil {
			return false
		}
		return a.Cmp(b) > 0
	})

	if amount == nil || amount.Sign() == 0 {
		max := 5
		if len(list) < max {
			max = len(list)
		}
		return list[:max], nil
	}

	sum := new(big.Int)
	var result []*FtUTXO
	for i := 0; i < len(list) && i < 5; i++ {
		ub, _ := new(big.Int).SetString(list[i].FtBalance, 10)
		if ub != nil {
			sum.Add(sum, ub)
		}
		result = append(result, list[i])
		if sum.Cmp(amount) >= 0 {
			return result, nil
		}
	}
	totalStr, err := GetFTBalance(contractTxID, addressOrHash, network)
	if err != nil {
		return nil, err
	}
	tb, _ := new(big.Int).SetString(totalStr, 10)
	if tb != nil && tb.Cmp(amount) >= 0 {
		return nil, fmt.Errorf("Insufficient FTbalance, please merge FT UTXOs")
	}
	return nil, fmt.Errorf("FTbalance not enough!")
}

// FetchFtUTXOsForPool 获取满足 amount 的 number 个 FT UTXO（至少 2 个才可提前结束）
// 对应 JS API.fetchFtUTXOsforPool
func FetchFtUTXOsForPool(contractTxID, addressOrHash, codeScript, network string, amount *big.Int, number int) ([]*FtUTXO, error) {
	if number <= 0 {
		return nil, fmt.Errorf("Number must be a positive integer greater than 0")
	}
	list, err := FetchFtUTXOList(contractTxID, addressOrHash, codeScript, network)
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		a, _ := new(big.Int).SetString(list[i].FtBalance, 10)
		b, _ := new(big.Int).SetString(list[j].FtBalance, 10)
		if a == nil {
			return true
		}
		if b == nil {
			return false
		}
		return a.Cmp(b) > 0
	})

	sum := new(big.Int)
	var result []*FtUTXO
	for i := 0; i < len(list) && i < number; i++ {
		ub, _ := new(big.Int).SetString(list[i].FtBalance, 10)
		if ub != nil {
			sum.Add(sum, ub)
		}
		result = append(result, list[i])
		if i >= 1 && sum.Cmp(amount) >= 0 {
			break
		}
	}
	if sum.Cmp(amount) < 0 {
		totalStr, err := GetFTBalance(contractTxID, addressOrHash, network)
		if err != nil {
			return nil, err
		}
		tb, _ := new(big.Int).SetString(totalStr, 10)
		if tb != nil && tb.Cmp(amount) >= 0 {
			return nil, fmt.Errorf("Insufficient FTbalance, please merge FT UTXOs")
		}
		return nil, fmt.Errorf("FTbalance not enough!")
	}
	return result, nil
}

// FetchFtUTXOsMultiSig 获取多签 FT UTXO 列表（按 ft_value 升序）
func FetchFtUTXOsMultiSig(contractTxID, addressOrHash, codeScript, network string) ([]*FtUTXO, error) {
	list, err := FetchFtUTXOList(contractTxID, addressOrHash, codeScript, network)
	if err != nil {
		return nil, err
	}
	// 按 ftBalance 升序（与 JS fetchFtUTXOS_multiSig 一致）
	sort.Slice(list, func(i, j int) bool {
		a, _ := new(big.Int).SetString(list[i].FtBalance, 10)
		b, _ := new(big.Int).SetString(list[j].FtBalance, 10)
		if a == nil {
			return true
		}
		if b == nil {
			return false
		}
		return a.Cmp(b) < 0
	})
	return list, nil
}

// findMinFiveSum 从 balances 中找 5 个数的组合，使其和 >= target 且最小
// 返回 5 个下标，找不到返回 nil
func findMinFiveSum(balances []*big.Int, target *big.Int) []int {
	n := len(balances)
	if n < 5 {
		return nil
	}
	minSum := new(big.Int).SetUint64(^uint64(0))
	var result []int
	for i := 0; i <= n-5; i++ {
		for j := i + 1; j <= n-4; j++ {
			left := j + 1
			right := n - 1
			for left < right-1 {
				sum := new(big.Int)
				sum.Add(sum, balances[i])
				sum.Add(sum, balances[j])
				sum.Add(sum, balances[left])
				sum.Add(sum, balances[right-1])
				sum.Add(sum, balances[right])
				if sum.Cmp(target) >= 0 && sum.Cmp(minSum) < 0 {
					minSum.Set(sum)
					result = []int{i, j, left, right - 1, right}
				}
				if sum.Cmp(target) < 0 {
					left++
				} else {
					right--
				}
			}
		}
	}
	return result
}

// GetFtUTXOsMultiSig 获取满足 amount 的多签 FT UTXO，最多 5 个
func GetFtUTXOsMultiSig(contractTxID, addressOrHash, codeScript, network string, amount *big.Int) ([]*FtUTXO, error) {
	list, err := FetchFtUTXOsMultiSig(contractTxID, addressOrHash, codeScript, network)
	if err != nil {
		return nil, err
	}
	balances := make([]*big.Int, len(list))
	total := new(big.Int)
	for i := range list {
		b, _ := new(big.Int).SetString(list[i].FtBalance, 10)
		balances[i] = b
		if b != nil {
			total.Add(total, b)
		}
	}
	if total.Cmp(amount) < 0 {
		return nil, fmt.Errorf("Insufficient FT balance")
	}
	if len(list) <= 5 {
		return list, nil
	}
	indices := findMinFiveSum(balances, amount)
	if indices == nil {
		return nil, fmt.Errorf("Please merge MultiSig UTXO")
	}
	return []*FtUTXO{
		list[indices[0]],
		list[indices[1]],
		list[indices[2]],
		list[indices[3]],
		list[indices[4]],
	}, nil
}

// FetchFtInfo 获取 FT 合约信息
// 对应 JS API.fetchFtInfo
func FetchFtInfo(contractTxID, network string) (*FtInfo, error) {
	baseURL := getBaseURL(network)
	url := fmt.Sprintf("%sft/info/contract/%s", baseURL, contractTxID)

	resp, err := defaultHTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r ftInfoResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("解析 FT Info 响应失败: %w", err)
	}
	amountStr, err := parseBigIntOrUint64(r.Data.Amount)
	if err != nil {
		return nil, fmt.Errorf("解析 FT amount: %w", err)
	}
	return &FtInfo{
		CodeScript:  r.Data.CodeScript,
		TapeScript:  r.Data.TapeScript,
		TotalSupply: amountStr,
		Decimal:     r.Data.Decimal,
		Name:        r.Data.Name,
		Symbol:      r.Data.Symbol,
	}, nil
}

// FetchFtPrePreTxData 获取 preTX 的 pre-pre 交易数据，用于 FT 合约解锁
// 对应 JS API.fetchFtPrePreTxData
func FetchFtPrePreTxData(preTX *Tx, preTxVout int, network string) (string, error) {
	if preTxVout+1 >= len(preTX.Outputs) {
		return "", fmt.Errorf("preTxVout+1 out of range")
	}
	tapeScript := preTX.Outputs[preTxVout+1].LockingScript.Bytes()
	if len(tapeScript) < 51 {
		return "", fmt.Errorf("tape script too short")
	}
	// preTXtape = script[3:51]
	tapeSlice := tapeScript[3:51]
	tapeHex := hex.EncodeToString(tapeSlice)

	var prepretxdata string
	// 从后往前每 16 字符一块
	for i := len(tapeHex) - 16; i >= 0; i -= 16 {
		chunk := tapeHex[i : i+16]
		if chunk != "0000000000000000" {
			inputIndex := i / 16
			if inputIndex >= len(preTX.Inputs) {
				return "", fmt.Errorf("input index out of range")
			}
			// PreviousTxID 已与 TxID()/浏览器 txid 同序，勿再 Reverse；否则 txraw 404+空 raw 会报 ErrTxTooShort。
			prevTxID := hex.EncodeToString(preTX.Inputs[inputIndex].PreviousTxID())
			prepreTX, err := FetchTXRaw(prevTxID, network)
			if err != nil {
				return "", err
			}
			data, err := GetPrePreTxdata(prepreTX, int(preTX.Inputs[inputIndex].PreviousTxOutIndex))
			if err != nil {
				return "", err
			}
			prepretxdata += data
		}
	}
	return "57" + prepretxdata, nil
}
