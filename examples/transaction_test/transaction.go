package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/libsv/go-bk/base58"
	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bk/wif"
	bt "github.com/sCrypt-Inc/go-bt/v2"
	"github.com/sCrypt-Inc/go-bt/v2/bscript"
	"github.com/sCrypt-Inc/go-bt/v2/unlocker"
)

// 网络类型常量
const (
	NetworkTestnet = "testnet"
	NetworkMainnet = "mainnet"
)

func main() {
	fmt.Println("=== TBC Transaction 测试程序 ===\n")

	// 示例1: 基本交易创建
	basicTransaction()

	// 示例2: 使用 UTXO 创建交易
	transactionFromUTXOs()

	// 示例3: 带找零地址的交易
	transactionWithChange()

	// 示例4: 获取输入和输出总额
	transactionAmounts()

	// 示例5: 交易序列化
	transactionSerialization()

	// 示例6: 手续费相关
	feeExamples()

	// 示例7: 时间锁定交易
	timeLockedTransaction()

	// 示例8: OP_RETURN 输出
	transactionWithOpReturn()

	// 示例9: 多签交易（2-of-3）
	multisigTransaction()

	fmt.Println("\n=== 所有测试完成 ===")
}

// 示例1: 基本交易创建
func basicTransaction() {
	fmt.Println("--- 示例1: 基本交易创建 ---")

	tx := bt.NewTx()

	// 添加输入 - 从之前的交易输出
	addressStr := "143KgKGcse57nXBnXyJwtQrf2KP4KWto59"
	network := NetworkTestnet
	utxo, err := bt.FetchUTXO(addressStr, 0.1, network)
	if err != nil {
		log.Printf("获取 UTXO 失败: %v", err)
		return
	}
	// err := tx.From(
	// 	"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d", // 之前的交易ID
	// 	0,                                                                    // 输出索引
	// 	"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",                 // 锁定脚本（标准 P2PKH 格式）
	// 	1500,                                                                 // satoshis 数量
	// )
	// if err != nil {
	// 	log.Printf("添加输入失败: %v", err)
	// 	return
	// }

	// 添加输出 - 支付到地址
	err = tx.PayToAddress("1Cxi449KxjkXxPUfQhB65vs3wzErMuGh6F", 1000)
	if err != nil {
		log.Printf("添加输出失败: %v", err)
		return
	}

	// 设置找零地址并计算手续费
	feeQuote := bt.NewFeeQuote()
	err = tx.ChangeToAddress(addressStr, feeQuote)
	if err != nil {
		log.Printf("设置找零地址失败: %v (这是正常的，因为示例数据可能不完整)", err)
	} else {
		fmt.Printf("找零地址设置成功\n")
	}

	fmt.Printf("输入数量: %d\n", tx.InputCount())
	fmt.Printf("输出数量: %d\n", tx.OutputCount())
	fmt.Printf("输入总额: %d satoshis\n", tx.TotalInputSatoshis())
	fmt.Printf("输出总额: %d satoshis\n", tx.TotalOutputSatoshis())

	// 签名交易（注意：示例中的 WIF 和锁定脚本可能不匹配，签名可能失败）,替换为你的WIF私钥
	decodedWif, err := wif.DecodeWIF("L1u2TmR7hMMMSV9Bx2Lyt3sujbboqEFqnKygnPRnQERhKB4qptuK")
	if err != nil {
		log.Printf("解码 WIF 失败: %v", err)
		fmt.Println()
		return
	}

	err = tx.FillAllInputs(context.Background(), &unlocker.Getter{PrivateKey: decodedWif.PrivKey})
	if err != nil {
		log.Printf("签名失败: %v (这是正常的，因为示例中的私钥和锁定脚本可能不匹配)", err)
		fmt.Printf("未签名的交易ID: %s\n", tx.TxID())
		fmt.Printf("未签名的交易序列化: %s\n\n", tx.String())
	} else {
		fmt.Printf("交易签名成功\n")
		fmt.Printf("交易ID: %s\n", tx.TxID())
		fmt.Printf("序列化交易: %s\n\n", tx.String())
	}
}

// 示例2: 使用 UTXO 创建交易
func transactionFromUTXOs() {
	fmt.Println("--- 示例2: 使用 UTXO 创建交易 ---")

	tx := bt.NewTx()

	// 创建 UTXO
	utxo1 := &bt.UTXO{
		TxID:          mustDecodeHex("11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d"),
		Vout:          0,
		Satoshis:      1000,
		LockingScript: mustLockingScript("76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac"),
	}

	utxo2 := &bt.UTXO{
		TxID:          mustDecodeHex("b7b0650a7c3a1bd4716369783876348b59f5404784970192cec1996e86950576"),
		Vout:          0,
		Satoshis:      2000,
		LockingScript: mustLockingScript("76a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac"),
	}

	// 从多个 UTXO 添加输入
	err := tx.FromUTXOs(utxo1, utxo2)
	if err != nil {
		log.Printf("从 UTXO 添加输入失败: %v", err)
		return
	}

	// 添加输出
	err = tx.PayToAddress("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 2500)
	if err != nil {
		log.Printf("添加输出失败: %v", err)
		return
	}

	fmt.Printf("输入数量: %d\n", tx.InputCount())
	fmt.Printf("输出数量: %d\n", tx.OutputCount())
	fmt.Printf("输入总额: %d satoshis\n", tx.TotalInputSatoshis())
	fmt.Printf("输出总额: %d satoshis\n\n", tx.TotalOutputSatoshis())
}

// 示例3: 带找零地址的交易
func transactionWithChange() {
	fmt.Println("--- 示例3: 带找零地址的交易 ---")

	tx := bt.NewTx()

	// 添加输入 (1500 satoshis)
	err := tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		1500,
	)
	if err != nil {
		log.Printf("添加输入失败: %v", err)
		return
	}

	// 添加输出 (1000 satoshis)
	err = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)
	if err != nil {
		log.Printf("添加输出失败: %v", err)
		return
	}

	// 设置找零地址 - 会自动计算手续费并将剩余金额作为找零
	feeQuote := bt.NewFeeQuote()
	err = tx.ChangeToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", feeQuote)
	if err != nil {
		log.Printf("设置找零地址失败: %v", err)
	} else {
		fmt.Printf("找零已自动添加到输出中\n")
	}

	fmt.Printf("输入总额: %d satoshis\n", tx.TotalInputSatoshis())
	fmt.Printf("输出总额: %d satoshis\n", tx.TotalOutputSatoshis())
	fmt.Printf("输出数量: %d\n\n", tx.OutputCount())
}

// 示例4: 获取输入和输出总额
func transactionAmounts() {
	fmt.Println("--- 示例4: 获取输入和输出总额 ---")

	tx := bt.NewTx()

	// 添加多个输入
	err := tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		1500,
	)
	if err == nil {
		err = tx.From(
			"b7b0650a7c3a1bd4716369783876348b59f5404784970192cec1996e86950576",
			0,
			"76a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac",
			2000,
		)
	}

	if err != nil {
		log.Printf("添加输入失败: %v", err)
		return
	}

	// 添加多个输出
	_ = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)
	_ = tx.PayToAddress("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 1500)

	inputAmount := tx.TotalInputSatoshis()
	outputAmount := tx.TotalOutputSatoshis()

	fmt.Printf("输入总额 (inputAmount): %d satoshis\n", inputAmount)
	fmt.Printf("输出总额 (outputAmount): %d satoshis\n", outputAmount)
	fmt.Printf("差额: %d satoshis\n\n", inputAmount-outputAmount)
}

// 示例5: 交易序列化
func transactionSerialization() {
	fmt.Println("--- 示例5: 交易序列化 ---")

	tx := bt.NewTx()

	_ = tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		1500,
	)

	_ = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)

	// toString() - 返回十六进制序列化字符串
	serializedHex := tx.String()
	fmt.Printf("序列化 (toString): %s\n", serializedHex)

	// Bytes() - 返回字节数组
	serializedBytes := tx.Bytes()
	fmt.Printf("序列化 (Bytes): %d 字节\n", len(serializedBytes))

	// NodeJSON() - 返回 JSON 格式的对象表示
	nodeJSON := tx.NodeJSON()
	fmt.Printf("JSON 格式可用: %v\n", nodeJSON != nil)

	// 交易ID
	fmt.Printf("交易ID: %s\n\n", tx.TxID())
}

// 示例6: 手续费相关
func feeExamples() {
	fmt.Println("--- 示例6: 手续费相关 ---")

	tx := bt.NewTx()

	_ = tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		1500,
	)

	_ = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)

	// 创建手续费报价
	feeQuote := bt.NewFeeQuote()

	// 使用 ChangeToAddress 会自动计算手续费并将剩余作为找零
	err := tx.ChangeToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", feeQuote)
	if err != nil {
		log.Printf("设置找零失败: %v", err)
		return
	}

	// 估算交易大小
	size, err := tx.EstimateSize()
	if err != nil {
		log.Printf("估算大小失败: %v", err)
	} else {
		fmt.Printf("估算交易大小: %d 字节\n", size)
	}

	// 检查手续费是否足够
	isEnough, err := tx.EstimateIsFeePaidEnough(feeQuote)
	if err != nil {
		log.Printf("检查手续费失败: %v", err)
	} else {
		fmt.Printf("手续费是否足够: %v\n", isEnough)
	}

	// 估算手续费
	fees, err := tx.EstimateFeesPaid(feeQuote)
	if err != nil {
		log.Printf("估算手续费失败: %v", err)
	} else {
		fmt.Printf("估算手续费: %d satoshis\n", fees.TotalFeePaid)
	}

	fmt.Printf("输入总额: %d satoshis\n", tx.TotalInputSatoshis())
	fmt.Printf("输出总额: %d satoshis\n\n", tx.TotalOutputSatoshis())
}

// 示例7: 时间锁定交易
func timeLockedTransaction() {
	fmt.Println("--- 示例7: 时间锁定交易 ---")

	tx := bt.NewTx()

	// 设置锁定时间 - 使用未来的日期
	futureDate := time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC)

	// 将时间戳转换为 locktime (Unix 时间戳)
	// LockTime 字段如果是时间戳，必须 >= 500000000
	locktime := uint32(futureDate.Unix())
	if locktime < 500000000 {
		locktime = uint32(futureDate.Unix()) + 500000000
	}
	tx.LockTime = locktime

	_ = tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		1500,
	)

	_ = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)

	fmt.Printf("锁定时间 (LockTime): %d\n", tx.LockTime)
	fmt.Printf("对应日期: %s\n", futureDate.Format(time.RFC3339))

	// 检查 LockTime 是否表示时间戳 (> 500000000) 还是区块高度
	if tx.LockTime >= 500000000 {
		lockTimeDate := time.Unix(int64(tx.LockTime), 0)
		fmt.Printf("解析后的锁定日期: %s\n", lockTimeDate.Format(time.RFC3339))
	} else {
		fmt.Printf("锁定区块高度: %d\n", tx.LockTime)
	}

	fmt.Println()
}

// 示例8: OP_RETURN 输出
func transactionWithOpReturn() {
	fmt.Println("--- 示例8: OP_RETURN 输出 ---")

	tx := bt.NewTx()

	_ = tx.From(
		"b7b0650a7c3a1bd4716369783876348b59f5404784970192cec1996e86950576",
		0,
		"76a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac",
		1000,
	)

	_ = tx.PayToAddress("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 900)

	// 添加 OP_RETURN 输出（数据输出）
	data := []byte("You are using go-bt!")
	err := tx.AddOpReturnOutput(data)
	if err != nil {
		log.Printf("添加 OP_RETURN 输出失败: %v", err)
		return
	}

	fmt.Printf("输出数量: %d\n", tx.OutputCount())
	fmt.Printf("包含数据输出: %v\n", tx.HasDataOutputs())
	fmt.Printf("交易序列化: %s\n\n", tx.String())
}

// 示例9: 多签交易（2-of-3 多签）
// 创建 2-of-3 多签地址，需要 3 个公钥中的任意 2 个签名才能花费
func multisigTransaction() {
	fmt.Println("--- 示例9: 多签交易（2-of-3） ---")

	// 创建 3 个私钥（用于演示，实际应用中这些私钥应该由不同的人保管）
	wif1, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
	if err != nil {
		log.Printf("解码 WIF1 失败: %v", err)
		return
	}

	wif2, err := wif.DecodeWIF("L1u2TmR7hMMMSV9Bx2Lyt3sujbboqEFqnKygnPRnQERhKB4qptuK")
	if err != nil {
		log.Printf("解码 WIF2 失败: %v", err)
		return
	}

	wif3, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
	if err != nil {
		log.Printf("解码 WIF3 失败: %v", err)
		return
	}

	// 获取公钥（压缩格式）
	pubKey1 := wif1.PrivKey.PubKey().SerialiseCompressed()
	pubKey2 := wif2.PrivKey.PubKey().SerialiseCompressed()
	pubKey3 := wif3.PrivKey.PubKey().SerialiseCompressed()

	fmt.Printf("公钥1: %s\n", hex.EncodeToString(pubKey1))
	fmt.Printf("公钥2: %s\n", hex.EncodeToString(pubKey2))
	fmt.Printf("公钥3: %s\n", hex.EncodeToString(pubKey3))

	// 构建多签脚本 (2-of-3): OP_2 <pubkey1> <pubkey2> <pubkey3> OP_3 OP_CHECKMULTISIG
	// 注意：OP_CHECKMULTISIG 需要一个额外的 OP_0 在签名前（这是 Bitcoin 的一个 bug/特性）
	multisigScript := bscript.NewFromBytes([]byte{})
	multisigScript.AppendOpcodes(bscript.Op2) // 需要 2 个签名
	multisigScript.AppendPushData(pubKey1)
	multisigScript.AppendPushData(pubKey2)
	multisigScript.AppendPushData(pubKey3)
	multisigScript.AppendOpcodes(bscript.Op3) // 总共 3 个公钥
	multisigScript.AppendOpcodes(bscript.OpCHECKMULTISIG)

	fmt.Printf("多签脚本 (redeem script): %s\n", multisigScript.String())

	// 创建 P2SH 地址（从多签脚本的哈希）
	scriptHash := crypto.Hash160([]byte(*multisigScript))
	p2shScript := bscript.NewFromBytes([]byte{})
	p2shScript.AppendOpcodes(bscript.OpHASH160)
	p2shScript.AppendPushData(scriptHash)
	p2shScript.AppendOpcodes(bscript.OpEQUAL)

	fmt.Printf("P2SH 锁定脚本: %s\n", p2shScript.String())

	// 手动创建 P2SH 地址（主网版本字节 0x05，测试网 0xc4）
	// 格式：版本字节 + 脚本哈希 + 校验和
	versionByte := byte(0x05) // 主网 P2SH，测试网使用 0xc4
	addressBytes := make([]byte, 0, 21)
	addressBytes = append(addressBytes, versionByte)
	addressBytes = append(addressBytes, scriptHash...)
	
	// 计算校验和
	checksum := crypto.Sha256d(addressBytes)[:4]
	addressBytes = append(addressBytes, checksum...)
	
	// Base58 编码
	p2shAddress := base58.Encode(addressBytes)
	
	fmt.Printf("多签 P2SH 地址: %s\n", p2shAddress)

	// 创建一个从多签地址花费的交易
	// 注意：这只是一个演示，实际的多签 UTXO 需要先创建并发送到多签地址
	tx := bt.NewTx()

	// 假设有一个来自多签地址的 UTXO（实际使用时需要替换为真实的 UTXO）
	multisigUTXO := &bt.UTXO{
		TxID:          mustDecodeHex("11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d"),
		Vout:          0,
		Satoshis:      50000,
		LockingScript: p2shScript, // P2SH 锁定脚本
	}

	err = tx.FromUTXOs(multisigUTXO)
	if err != nil {
		log.Printf("添加多签输入失败: %v", err)
		return
	}

	// 添加输出
	err = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 40000)
	if err != nil {
		log.Printf("添加输出失败: %v", err)
		return
	}

	fmt.Printf("输入数量: %d\n", tx.InputCount())
	fmt.Printf("输出数量: %d\n", tx.OutputCount())

	// 对多签输入进行签名（需要至少 2 个签名）
	// 注意：go-bt 的多签签名需要手动构建解锁脚本
	// 这里只是演示结构，实际的多签签名更复杂

	// 构建解锁脚本：OP_0 <sig1> <sig2> <redeem_script>
	// 注意：OP_0 是 CHECKMULTISIG 的一个已知 bug/特性
	unlockingScript := bscript.NewFromBytes([]byte{})
	unlockingScript.AppendOpcodes(bscript.Op0) // CHECKMULTISIG bug 需要这个

	// 这里应该添加两个签名（使用私钥1和私钥2）
	// 实际实现需要使用 sighash 计算并签名
	// 为了演示，我们只展示结构
	fmt.Printf("多签解锁脚本结构: OP_0 <sig1> <sig2> <redeem_script>\n")
	fmt.Printf("注意：实际的多签签名需要计算正确的 sighash 并使用对应的私钥签名\n")

	// 设置解锁脚本到输入（这里只是演示，实际需要正确的签名）
	// tx.Inputs[0].UnlockingScript = unlockingScript

	fmt.Printf("多签交易构建完成（未签名，需要正确的签名才能广播）\n")
	fmt.Printf("交易ID: %s\n\n", tx.TxID())
}

// 辅助函数：解码十六进制字符串
func mustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// 辅助函数：从十六进制字符串创建锁定脚本
func mustLockingScript(hexStr string) *bscript.Script {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return bscript.NewFromBytes(b)
}
