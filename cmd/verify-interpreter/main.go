package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	bt "github.com/sCrypt-Inc/go-bt/v2"
	"github.com/sCrypt-Inc/go-bt/v2/bscript"
	"github.com/sCrypt-Inc/go-bt/v2/bscript/interpreter"
)

// 网络类型（直接复用 api.go 里的字符串约定）
const (
	NetworkTestnet = "testnet"
	NetworkMainnet = "mainnet"
)

func main() {
	// 从命令行参数获取 txid，如果没有提供则使用默认值
	var txid string
	var network string = NetworkMainnet

	if len(os.Args) < 2 {
		log.Println("用法: verify-interpreter <txid> [network]")
		log.Println("示例: verify-interpreter eb40327709215a77cf4c671c2574c12ec5807490a7277324bee132606a5d8ea8 mainnet")
		log.Println("")
		log.Println("使用默认 txid 进行测试...")
		txid = "eb40327709215a77cf4c671c2574c12ec5807490a7277324bee132606a5d8ea8"
	} else {
		txid = os.Args[1]
		if len(os.Args) >= 3 {
			network = os.Args[2]
		}
	}

	fmt.Printf("=== 交易验证程序 ===\n")
	fmt.Printf("网络: %s\n", network)
	fmt.Printf("交易ID: %s\n\n", txid)

	// 步骤1: 从链上获取交易（支持未确认交易）
	fmt.Println("步骤1: 从链上获取交易...")
	tx, err := bt.FetchTXRaw(txid, network)
	if err != nil {
		log.Fatalf("❌ 获取交易失败: %v", err)
	}
	fmt.Printf("✅ 成功获取交易\n")
	fmt.Printf("   交易版本: %d\n", tx.Version)
	fmt.Printf("   输入数量: %d\n", tx.InputCount())
	fmt.Printf("   输出数量: %d\n", tx.OutputCount())
	fmt.Printf("   锁定时间: %d\n", tx.LockTime)
	fmt.Printf("   原始交易十六进制长度: %d 字节\n\n", len(tx.String())/2)

	// 检查交易是否在链上（已确认）
	isOnChain, err := bt.IsTxOnChain(txid, network)
	if err != nil {
		log.Printf("⚠️  检查交易状态失败: %v", err)
	} else if isOnChain {
		fmt.Printf("ℹ️  交易状态: 已在链上（已确认）\n\n")
	} else {
		fmt.Printf("ℹ️  交易状态: 未确认（可能在内存池中）\n\n")
	}

	if tx.InputCount() == 0 {
		log.Fatal("❌ 交易没有任何输入，无法执行脚本验证")
	}

	// 步骤2: 重新组装交易并验证所有输入
	fmt.Println("步骤2: 重新组装交易并验证所有输入...")
	fmt.Printf("开始验证 %d 个输入...\n\n", tx.InputCount())

	successCount := 0
	skipCount := 0
	failCount := 0

	// 遍历所有输入进行验证
	for i := 0; i < tx.InputCount(); i++ {
		input := tx.InputIdx(i)
		
		// 跳过 Coinbase 输入
		if isCoinbaseInput(input) {
			fmt.Printf("输入 %d: ⏭️  检测到Coinbase输入，跳过验证\n", i)
			skipCount++
			continue
		}

		// 获取前序交易ID和输出索引
		prevTxID := input.PreviousTxIDStr()
		vout := input.PreviousTxOutIndex

		fmt.Printf("输入 %d:\n", i)
		fmt.Printf("  前序交易ID: %s\n", prevTxID)
		fmt.Printf("  输出索引: %d\n", vout)

		// 获取前序交易
		prevTx, err := bt.FetchTXRaw(prevTxID, network)
		if err != nil {
			fmt.Printf("  ❌ 获取前序交易失败: %v\n\n", err)
			failCount++
			continue
		}

		// 从前序交易中获取对应的输出
		prevOutput := prevTx.OutputIdx(int(vout))
		if prevOutput == nil {
			fmt.Printf("  ❌ 前序交易在索引 %d 处没有输出\n\n", vout)
			failCount++
			continue
		}

		fmt.Printf("  前序输出金额: %d satoshis\n", prevOutput.Satoshis)
		if prevOutput.LockingScript != nil {
			fmt.Printf("  锁定脚本: %s\n", hex.EncodeToString(prevOutput.LockingScript.Bytes()))
		}

		// 使用解释器验证
		if err := verifyWithInterpreter(tx, i, prevOutput); err != nil {
			fmt.Printf("  ❌ 解释器验证失败: %v\n\n", err)
			failCount++
		} else {
			fmt.Printf("  ✅ 解释器验证通过\n\n")
			successCount++
		}
	}

	// 步骤3: 输出验证结果摘要
	fmt.Println("=== 验证结果摘要 ===")
	fmt.Printf("总输入数: %d\n", tx.InputCount())
	fmt.Printf("✅ 验证成功: %d\n", successCount)
	fmt.Printf("⏭️  跳过 (Coinbase): %d\n", skipCount)
	fmt.Printf("❌ 验证失败: %d\n", failCount)

	if failCount > 0 {
		log.Fatalf("\n❌ 部分输入验证失败，请检查交易数据")
	} else if successCount == 0 && skipCount == tx.InputCount() {
		fmt.Println("\n⚠️  所有输入都是Coinbase输入，无法使用标准脚本验证")
	} else {
		fmt.Println("\n✅ 所有非Coinbase输入验证通过！")
	}
}

// isCoinbaseInput 检测输入是否为Coinbase输入
// Coinbase输入的特征：PreviousTxID全0，PreviousTxOutIndex为最大值(0xFFFFFFFF)
func isCoinbaseInput(input *bt.Input) bool {
	prevTxID := input.PreviousTxID()
	vout := input.PreviousTxOutIndex
	
	// 检查PreviousTxID是否全0
	allZero := true
	for _, b := range prevTxID {
		if b != 0 {
			allZero = false
			break
		}
	}
	
	// Coinbase: PreviousTxID全0 且 PreviousTxOutIndex为最大值
	return allZero && vout == 0xFFFFFFFF
}

// verifyWithInterpreter 使用 go-bt 的脚本解释器，对构建好的交易执行脚本验证。
// tx: 要验证的交易
// inputIdx: 要验证的输入索引
// prevOutput: 前序交易的输出（包含金额和锁定脚本）
func verifyWithInterpreter(tx *bt.Tx, inputIdx int, prevOutput *bt.Output) error {
	if tx == nil {
		return fmt.Errorf("tx 为空")
	}

	if prevOutput == nil {
		return fmt.Errorf("prevOutput 为空")
	}

	if inputIdx < 0 || inputIdx >= tx.InputCount() {
		return fmt.Errorf("输入索引 %d 超出范围（交易共有 %d 个输入）", inputIdx, tx.InputCount())
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

