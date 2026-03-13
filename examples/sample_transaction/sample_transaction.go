package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/libsv/go-bk/wif"
	bt "github.com/sCrypt-Inc/go-bt/v2"
	"github.com/sCrypt-Inc/go-bt/v2/bscript"
	"github.com/sCrypt-Inc/go-bt/v2/bscript/interpreter"
	"github.com/sCrypt-Inc/go-bt/v2/unlocker"
)

// 网络类型（直接复用 api.go 里的字符串约定）
const (
	NetworkTestnet = "testnet"
	NetworkMainnet = "mainnet"
)

func main() {
	fmt.Println("=== TBC Transaction 构建和广播测试程序 ===")

	// ========= 0. 显式设置网络 =========
	network := NetworkTestnet // 可选: NetworkTestnet 或 NetworkMainnet
	fmt.Printf("当前网络: %s\n", network)

	// ========= 1. 配置私钥 & 地址 =========
	// TODO: 替换为你的 WIF 私钥
	wifStr := "L1u2TmR7hMMMSV9Bx2Lyt3sujbboqEFqnKygnPRnQERhKB4qptuK"
	if wifStr == "" || strings.HasPrefix(wifStr, "<") {
		log.Fatal("请先在 sample_transaction.go 中填写你的 WIF 私钥 (wifStr 变量)")
	}

	decodedWif, err := wif.DecodeWIF(wifStr)
	if err != nil {
		log.Fatalf("解码 WIF 失败: %v", err)
	}

	privateKey := decodedWif.PrivKey
	// 根据网络类型确定 mainnet 参数（testnet = false, mainnet = true）
	// isMainnet := network == NetworkMainnet
	// address, err := bscript.NewAddressFromPublicKey(decodedWif.PrivKey.PubKey(), isMainnet)
	// if err != nil {
	// 	log.Fatalf("从私钥生成地址失败: %v", err)
	// }

	addressStr := "143KgKGcse57nXBnXyJwtQrf2KP4KWto59"
	fmt.Printf("发送方地址(from): %s\n", addressStr)

	// ========= 2. 获取 UTXO =========
	// 方式1: 通过 API 自动获取 UTXO（推荐）
	utxo, err := bt.FetchUTXO(addressStr, 0.1, network)
	if err != nil {
		log.Printf("⚠️  从 API 获取 UTXO 失败: %v", err)
		log.Println("尝试使用手动指定的 UTXO...")

		// 方式2: 手动指定 UTXO（如果 API 失败，使用手动指定的 UTXO）
		// TODO: 请从区块浏览器查询该地址的真实 UTXO，并替换下面的值
		// 注意：txid 需要是十六进制字符串（大端序，不需要反转）
		// script 是该 UTXO 的锁定脚本（scriptPubKey）
		// utxo = &bt.UTXO{
		// 	TxID:          mustDecodeHex("7dc77da876c1c6765ea3a8a234274328f545168d61f9e7bcbed487bee6a93e9b"),
		// 	Vout:          0,
		// 	Satoshis:      100000,
		// 	LockingScript: mustLockingScript("76a9142158ccfe3dc673b74e67c1ffd77842fd8bc4361c88ac"),
		// }

		if utxo == nil {
			log.Fatal("无法获取 UTXO，请检查网络连接或手动指定 UTXO")
		}
		log.Println("✅ 使用手动指定的 UTXO")
	}

	fmt.Printf("UTXO 信息: txid=%s, vout=%d, satoshis=%d\n",
		hex.EncodeToString(utxo.TxID), utxo.Vout, utxo.Satoshis)

	// ========= 3. 配置收款地址 & 金额 =========
	toAddressStr := "1Cxi449KxjkXxPUfQhB65vs3wzErMuGh6F" // TODO: 替换为接收方地址
	if toAddressStr == "" || strings.HasPrefix(toAddressStr, "<") {
		log.Fatal("请先在 sample_transaction.go 中填写接收方地址 (toAddressStr)")
	}

	sendAmount := uint64(50000) // 要转给收款地址的金额 (satoshis)
	fee := uint64(1000)         // 手续费 (satoshis)

	if sendAmount+fee > utxo.Satoshis {
		log.Fatalf("发送金额 (%d) + 手续费 (%d) 大于 UTXO 金额 (%d)，请调整数值",
			sendAmount, fee, utxo.Satoshis)
	}

	// ========= 4. 开始构建 Transaction =========
	tx := bt.NewTx()

	// 4.1 添加输入（from UTXO）
	err = tx.FromUTXOs(utxo)
	if err != nil {
		log.Fatalf("添加输入失败: %v", err)
	}

	// 4.2 添加主输出：转给接收地址
	err = tx.PayToAddress(toAddressStr, sendAmount)
	if err != nil {
		log.Fatalf("添加输出失败: %v", err)
	}

	// 4.3 设置找零地址（多余的钱回到 fromAddress）
	// 注意: go-bt 的 ChangeToAddress 会自动计算手续费，但我们可以手动设置手续费
	// 这里我们先手动计算找零
	changeAmount := utxo.Satoshis - sendAmount - fee
	if changeAmount >= 42 { // DUST_AMOUNT = 42
		err = tx.PayToAddress(addressStr, changeAmount)
		if err != nil {
			log.Fatalf("添加找零输出失败: %v", err)
		}
	}

	fmt.Printf("输入金额 (satoshis): %d\n", tx.TotalInputSatoshis())
	fmt.Printf("输出金额 (satoshis): %d\n", tx.TotalOutputSatoshis())
	fmt.Printf("预计找零金额 (satoshis): %d\n", changeAmount)

	// ========= 5. 使用私钥对交易进行签名 =========
	err = tx.FillAllInputs(context.Background(), &unlocker.Getter{PrivateKey: privateKey})
	if err != nil {
		log.Fatalf("签名失败: %v", err)
	}

	fmt.Printf("是否已全部签名: 是\n")

	// ========= 6. 序列化并输出结果 =========
	rawHex := tx.String() // 返回十六进制序列化字符串
	// fmt.Println("\ntxid:", tx.TxID)
	fmt.Println("\n原始交易十六进制(raw tx hex):")
	fmt.Println(rawHex)

	// ========= 6.1 使用解释器对构建好的交易进行脚本执行验证 =========
	if err := verifyWithInterpreter(tx, utxo); err != nil {
		log.Fatalf("解释器执行失败: %v", err)
	}
	fmt.Println("✅ 解释器执行通过（脚本验证成功）")

	// ========= 7. 通过 API 广播真实交易 =========
	fmt.Println("\n开始通过官方 API 广播交易...")

	txid, err := bt.BroadcastTXRaw(rawHex, network)
	if err != nil {
		log.Fatalf("❌ 交易广播失败: %v", err)
	}

	fmt.Printf("✅ 交易广播成功，txid: %s\n", txid)

	// ========= 8. 检查交易是否在链上 =========
	fmt.Println("\n检查交易是否在链上...")

	// 等待几秒让交易被处理
	time.Sleep(3 * time.Second)

	isOnChain, err := bt.IsTxOnChain(txid, network)
	if err != nil {
		log.Printf("⚠️  检查交易状态失败: %v", err)
	} else if isOnChain {
		fmt.Printf("✅ 交易已在链上，txid: %s\n", txid)
	} else {
		fmt.Printf("⏳ 交易尚未在链上，txid: %s（可能还在等待确认）\n", txid)
	}
}

// verifyWithInterpreter 使用 go-bt 的脚本解释器，对构建好的交易执行脚本验证。
// 这里利用我们刚刚使用的 utxo 作为 prevTxOut（只需要金额和锁定脚本即可）。
func verifyWithInterpreter(tx *bt.Tx, utxo *bt.UTXO) error {
	if tx == nil || utxo == nil {
		return fmt.Errorf("tx 或 utxo 为空")
	}

	if tx.InputCount() == 0 {
		return fmt.Errorf("交易没有任何输入，无法执行脚本验证")
	}

	// 只验证第 0 个输入；如果有多个输入，可以按需循环。
	inputIdx := 0

	// 构造一个“前序输出”对象，包含金额和锁定脚本。
	prevOutput := &bt.Output{
		Satoshis:      utxo.Satoshis,
		LockingScript: utxo.LockingScript,
	}

	return interpreter.NewEngine().Execute(
		interpreter.WithTx(tx, inputIdx, prevOutput),
		interpreter.WithForkID(),
		interpreter.WithAfterGenesis(),
	)
}

// mustDecodeHex 辅助函数：解码十六进制字符串
func mustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// mustLockingScript 辅助函数：从十六进制字符串创建锁定脚本
func mustLockingScript(hexStr string) *bscript.Script {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return bscript.NewFromBytes(b)
}
