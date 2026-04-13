package bt

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sCrypt-Inc/go-bt/v2/bscript"
)

// defaultHTTPClient 默认 HTTP 客户端，带超时
var defaultHTTPClient = &http.Client{Timeout: 30 * time.Second}

// 对应 tbc-contract/lib/api/api.ts 的基础 URL 配置
const (
	mainnetAPIURL = "https://api.turingbitchain.io/api/tbc/"
	testnetAPIURL = "https://api.tbcdev.org/api/tbc/"
)

// getBaseURL 获取指定网络的基础 URL
// network 可以是 "testnet" / "mainnet" / 自定义完整 URL
func getBaseURL(network string) string {
	switch network {
	case "testnet":
		return testnetAPIURL
	case "mainnet", "":
		// 默认为主网
		return mainnetAPIURL
	default:
		if network[len(network)-1] == '/' {
			return network
		}
		return network + "/"
	}
}

// ----- 公共响应结构 -----

type balanceResponse struct {
	Data struct {
		Balance uint64 `json:"balance"`
	} `json:"data"`
}

type utxoListResponse struct {
	Data struct {
		UTXOs []struct {
			TxID  string `json:"txid"`
			Index int    `json:"index"`
			Value uint64 `json:"value"`
		} `json:"utxos"`
	} `json:"data"`
}

type txrawResponse struct {
	Data struct {
		TxRaw string `json:"txraw"`
	} `json:"data"`
}

type broadcastResponse struct {
	Code    string `json:"code"`
	Data    struct {
		TxID    string `json:"txid"`
		Error   string `json:"error"`
		Success int    `json:"success"`
		Failed  int    `json:"failed"`
	} `json:"data"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// BlockHeaderInfo 区块头信息，对应 JS API 返回格式
type BlockHeaderInfo struct {
	Hash             string `json:"hash"`
	Confirmations    int    `json:"confirmations"`
	Height           int    `json:"height"`
	Version          int    `json:"version"`
	VersionHex       string `json:"versionHex"`
	MerkleRoot       string `json:"merkleroot"`
	Time             int64  `json:"time"`
	Nonce            uint32 `json:"nonce"`
	Bits             string `json:"bits"`
	Difficulty       string `json:"difficulty"`
	PreviousBlockHash string `json:"previoushash"`
	NextBlockHash    string `json:"nexthash"`
}

type blockHeadersResponse struct {
	Data []BlockHeaderInfo `json:"data"`
}

// BroadcastTXsRequestItem 批量广播时单条请求格式
type BroadcastTXsRequestItem struct {
	TxRaw string `json:"txraw"`
}

// ----- 导出函数：与 TS API 类似的接口 -----

// GetTBCBalance 获取指定地址的 TBC 余额（单位：satoshis）
// network: "testnet" / "mainnet" / 自定义 URL，空字符串默认 mainnet
func GetTBCBalance(address, network string) (uint64, error) {
	baseURL := getBaseURL(network)
	url := fmt.Sprintf("%sbalance/address/%s", baseURL, address)

	resp, err := defaultHTTPClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("请求余额接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("余额接口返回状态码 %d: %s", resp.StatusCode, string(body))
	}

	var br balanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&br); err != nil {
		return 0, fmt.Errorf("解析余额响应失败: %w", err)
	}

	return br.Data.Balance, nil
}

// FetchUTXO 根据地址和需要的金额，选择一个满足条件的 UTXO。
// amountTBC 为 TBC 数量（例如 0.1），内部会转换为 satoshis（1e6）。
// 返回值为 *bt.UTXO，可直接用于 tx.FromUTXOs。
func FetchUTXO(address string, amountTBC float64, network string) (*UTXO, error) {
	baseURL := getBaseURL(network)
	url := fmt.Sprintf("%sutxo/address/%s", baseURL, address)

	resp, err := defaultHTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 UTXO 接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("UTXO 接口返回状态码 %d: %s", resp.StatusCode, string(body))
	}

	var ur utxoListResponse
	if err := json.NewDecoder(resp.Body).Decode(&ur); err != nil {
		return nil, fmt.Errorf("解析 UTXO 响应失败: %w", err)
	}

	if len(ur.Data.UTXOs) == 0 {
		return nil, fmt.Errorf("该地址没有可用的 UTXO")
	}

	amountSatoshis := uint64(amountTBC * 1e6)

	// 简单选择一个 >= 目标金额的 UTXO，如果没有则退回第一个
	var selected = &ur.Data.UTXOs[0]
	for i := range ur.Data.UTXOs {
		if ur.Data.UTXOs[i].Value >= amountSatoshis {
			selected = &ur.Data.UTXOs[i]
			break
		}
	}

	txidBytes, err := hex.DecodeString(selected.TxID)
	if err != nil {
		return nil, fmt.Errorf("解码 txid 失败: %w", err)
	}

	chainTx, err := FetchTXRaw(selected.TxID, network)
	if err != nil {
		return nil, fmt.Errorf("拉取 UTXO 父交易以校准 script/金额失败: %w", err)
	}
	if selected.Index < 0 || selected.Index >= len(chainTx.Outputs) {
		return nil, fmt.Errorf("UTXO vout %d 超出父交易输出数 %d", selected.Index, len(chainTx.Outputs))
	}
	out := chainTx.Outputs[selected.Index]
	if out.LockingScript == nil {
		return nil, fmt.Errorf("链上输出 %s:%d 无 locking script", selected.TxID, selected.Index)
	}

	return &UTXO{
		TxID:          txidBytes,
		Vout:          uint32(selected.Index),
		Satoshis:      out.Satoshis,
		LockingScript: out.LockingScript,
	}, nil
}

// BroadcastTXRaw 将原始交易十六进制串广播到网络。
// network: "testnet" / "mainnet" / 自定义 URL，空字符串默认 mainnet。
// 返回 txid。
func BroadcastTXRaw(txraw, network string) (string, error) {
	baseURL := getBaseURL(network)
	url := baseURL + "broadcasttx"

	payload := map[string]string{"txraw": txraw}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var br broadcastResponse
	if err := json.Unmarshal(body, &br); err != nil {
		return "", fmt.Errorf("解析广播响应失败: %w, 内容: %s", err, string(body))
	}

	if br.Code == "200" {
		return br.Data.TxID, nil
	}

	// 拼接更详细的错误信息
	errMsg := br.Message
	if br.Data.Error != "" {
		errMsg = br.Data.Error
	}
	if br.Error != "" {
		errMsg = br.Error
	}
	if errMsg == "" {
		errMsg = fmt.Sprintf("广播失败，code=%s", br.Code)
	}

	return "", fmt.Errorf("%s", errMsg)
}

// FetchTXRaw 根据 txid 拉取原始交易并解析为 *bt.Tx
func FetchTXRaw(txid, network string) (*Tx, error) {
	baseURL := getBaseURL(network)
	url := fmt.Sprintf("%stxraw/txid/%s", baseURL, txid)

	resp, err := defaultHTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 TXRaw 接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TXRaw 接口返回状态码 %d: %s", resp.StatusCode, string(body))
	}

	var tr struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			TxRaw string `json:"txraw"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("解析 TXRaw 响应失败: %w", err)
	}
	if tr.Code != "" && tr.Code != "200" {
		msg := tr.Message
		if msg == "" {
			msg = "unknown error"
		}
		return nil, fmt.Errorf("TXRaw 接口业务失败 code=%s: %s", tr.Code, msg)
	}
	raw := strings.TrimSpace(tr.Data.TxRaw)
	if raw == "" {
		return nil, fmt.Errorf("TXRaw 响应缺少 txraw (txid=%s)", txid)
	}

	return NewTxFromString(raw)
}

// IsTxOnChain 判断交易是否在链上
// 通过尝试获取交易数据来判断：如果能成功获取，说明交易在链上；否则不在链上
// network: "testnet" / "mainnet" / 自定义 URL，空字符串默认 mainnet
// 返回: (是否在链上, 错误信息)
func IsTxOnChain(txid, network string) (bool, error) {
	baseURL := getBaseURL(network)
	url := fmt.Sprintf("%stxraw/txid/%s", baseURL, txid)

	resp, err := defaultHTTPClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("请求 TXRaw 接口失败: %w", err)
	}
	defer resp.Body.Close()

	// 如果状态码是 200，说明交易存在（在链上）
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	// 如果状态码是 404，说明交易不存在（不在链上）
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	// 其他状态码，返回错误
	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("TXRaw 接口返回状态码 %d: %s", resp.StatusCode, string(body))
}

// FetchUTXOs 获取指定地址的所有 UTXO
// 对应 JS API.fetchUTXOs(address, network)
func FetchUTXOs(address, network string) (UTXOs, error) {
	baseURL := getBaseURL(network)
	url := fmt.Sprintf("%sutxo/address/%s", baseURL, address)

	resp, err := defaultHTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 UTXO 接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("UTXO 接口返回状态码 %d: %s", resp.StatusCode, string(body))
	}

	var ur utxoListResponse
	if err := json.NewDecoder(resp.Body).Decode(&ur); err != nil {
		return nil, fmt.Errorf("解析 UTXO 响应失败: %w", err)
	}

	if len(ur.Data.UTXOs) == 0 {
		return nil, fmt.Errorf("The balance in the account is zero.")
	}

	lockingScript, err := bscript.NewP2PKHFromAddress(address)
	if err != nil {
		return nil, fmt.Errorf("创建锁定脚本失败: %w", err)
	}

	result := make(UTXOs, 0, len(ur.Data.UTXOs))
	for i := range ur.Data.UTXOs {
		txidBytes, err := hex.DecodeString(ur.Data.UTXOs[i].TxID)
		if err != nil {
			return nil, fmt.Errorf("解码 txid 失败: %w", err)
		}
		result = append(result, &UTXO{
			TxID:          txidBytes,
			Vout:          uint32(ur.Data.UTXOs[i].Index),
			Satoshis:      ur.Data.UTXOs[i].Value,
			LockingScript: lockingScript,
		})
	}
	return result, nil
}

// GetUTXOs 获取地址 UTXO 列表，并校验总余额是否满足 amountTBC
// 对应 JS API.getUTXOs(address, amount_tbc, network)
func GetUTXOs(address string, amountTBC float64, network string) (UTXOs, error) {
	utxos, err := FetchUTXOs(address, network)
	if err != nil {
		return nil, err
	}
	amountSatoshis := uint64(amountTBC * 1e6)
	var total uint64
	for _, u := range utxos {
		total += u.Satoshis
	}
	if total < amountSatoshis {
		return nil, fmt.Errorf("Insufficient tbc balance")
	}
	return utxos, nil
}

// BroadcastTXsRaw 批量广播原始交易
// txrawList: [{TxRaw: "hex"}...]，对应 JS API.broadcastTXsraw
// 返回 success 数量、failed 数量、错误
func BroadcastTXsRaw(txrawList []BroadcastTXsRequestItem, network string) (success, failed int, err error) {
	baseURL := getBaseURL(network)
	url := baseURL + "broadcasttxs"

	jsonData, err := json.Marshal(txrawList)
	if err != nil {
		return 0, 0, fmt.Errorf("序列化请求体失败: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, 0, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("读取响应失败: %w", err)
	}

	var br broadcastResponse
	if err := json.Unmarshal(body, &br); err != nil {
		return 0, 0, fmt.Errorf("解析广播响应失败: %w, 内容: %s", err, string(body))
	}

	if br.Code == "200" {
		return br.Data.Success, br.Data.Failed, nil
	}
	if br.Code == "400" && (bytes.Contains(body, []byte("partial failure")) || br.Data.Success > 0) {
		return br.Data.Success, br.Data.Failed, nil
	}
	errMsg := br.Message
	if br.Error != "" {
		errMsg = br.Error
	}
	if errMsg == "" {
		errMsg = "Broadcast failed"
	}
	return 0, 0, fmt.Errorf("%s", errMsg)
}

// FetchBlockHeaders 拉取最近区块头
// 对应 JS API.fetchBlockHeaders(network)
// start=0, end=1 表示拉取 1 个区块
func FetchBlockHeaders(network string) ([]BlockHeaderInfo, error) {
	baseURL := getBaseURL(network)
	url := fmt.Sprintf("%srecentblocks/start/0/end/1", baseURL)

	resp, err := defaultHTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求区块头接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Failed to fetch block headers: %s", string(body))
	}

	var r blockHeadersResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("解析区块头响应失败: %w", err)
	}
	return r.Data, nil
}


