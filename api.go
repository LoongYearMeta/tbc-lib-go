package bt

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sCrypt-Inc/go-bt/v2/bscript"
)

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
		TxID  string `json:"txid"`
		Error string `json:"error"`
	} `json:"data"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// ----- 导出函数：与 TS API 类似的接口 -----

// GetTBCBalance 获取指定地址的 TBC 余额（单位：satoshis）
// network: "testnet" / "mainnet" / 自定义 URL，空字符串默认 mainnet
func GetTBCBalance(address, network string) (uint64, error) {
	baseURL := getBaseURL(network)
	url := fmt.Sprintf("%sbalance/address/%s", baseURL, address)

	resp, err := http.Get(url)
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
	fmt.Printf("[Go FetchUTXO] network=%s url=%s\n", network, url)

	resp, err := http.Get(url)
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

	// 构造锁定脚本（P2PKH）
	lockingScript, err := bscript.NewP2PKHFromAddress(address)
	if err != nil {
		return nil, fmt.Errorf("创建锁定脚本失败: %w", err)
	}

	txidBytes, err := hex.DecodeString(selected.TxID)
	if err != nil {
		return nil, fmt.Errorf("解码 txid 失败: %w", err)
	}

	return &UTXO{
		TxID:          txidBytes,                 // 大端序；内部会在序列化时处理反转
		Vout:          uint32(selected.Index),    // 输出索引
		Satoshis:      selected.Value,           // 金额
		LockingScript: lockingScript,           // P2PKH 锁定脚本
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

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 TXRaw 接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TXRaw 接口返回状态码 %d: %s", resp.StatusCode, string(body))
	}

	var tr txrawResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("解析 TXRaw 响应失败: %w", err)
	}

	return NewTxFromString(tr.Data.TxRaw)
}

// IsTxOnChain 判断交易是否在链上
// 通过尝试获取交易数据来判断：如果能成功获取，说明交易在链上；否则不在链上
// network: "testnet" / "mainnet" / 自定义 URL，空字符串默认 mainnet
// 返回: (是否在链上, 错误信息)
func IsTxOnChain(txid, network string) (bool, error) {
	baseURL := getBaseURL(network)
	url := fmt.Sprintf("%stxraw/txid/%s", baseURL, txid)

	resp, err := http.Get(url)
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


