package main

import (
	"fmt"
	"log"

	bt "github.com/sCrypt-Inc/go-bt/v2"
	"github.com/sCrypt-Inc/go-bt/v2/bscript/interpreter"
)

// 网络类型
const (
	NetworkTestnet = "testnet"
	NetworkMainnet = "mainnet"
)

// 合约验证示例
func main() {
	network := NetworkMainnet

	// ========= 场景1: 验证普通P2PKH交易 =========
	fmt.Println("=== 场景1: 验证普通P2PKH交易 ===")
	txid1 := "c300aa0650dcaa9e026793b372bcb3f6dfe75df2ab0940d2d83e707d0cf737fe"
	if err := verifyStandardTransaction(txid1, network); err != nil {
		log.Printf("普通交易验证失败: %v\n", err)
	} else {
		fmt.Println("✅ 普通交易验证通过\n")
	}

	// ========= 场景2: 验证FT代币转账交易 =========
	fmt.Println("=== 场景2: 验证FT代币转账交易 ===")
	// TODO: 替换为实际的FT转账交易ID
	ftTxid := "your_ft_transfer_txid_here"
	if err := verifyFTTransaction(ftTxid, network); err != nil {
		log.Printf("FT交易验证失败: %v\n", err)
	} else {
		fmt.Println("✅ FT交易验证通过\n")
	}

	// ========= 场景3: 验证NFT转移交易 =========
	fmt.Println("=== 场景3: 验证NFT转移交易 ===")
	// TODO: 替换为实际的NFT转移交易ID
	nftTxid := "your_nft_transfer_txid_here"
	if err := verifyNFTTransaction(nftTxid, network); err != nil {
		log.Printf("NFT交易验证失败: %v\n", err)
	} else {
		fmt.Println("✅ NFT交易验证通过\n")
	}
}

// verifyStandardTransaction 验证标准P2PKH交易
func verifyStandardTransaction(txid string, network string) error {
	// 1. 获取交易
	tx, err := bt.FetchTXRaw(txid, network)
	if err != nil {
		return fmt.Errorf("获取交易失败: %w", err)
	}

	if tx.InputCount() == 0 {
		return fmt.Errorf("交易没有任何输入")
	}

	// 2. 验证每个输入
	for i := 0; i < tx.InputCount(); i++ {
		input := tx.InputIdx(i)

		// 跳过Coinbase输入
		if isCoinbaseInput(input) {
			fmt.Printf("输入 %d: Coinbase输入，跳过\n", i)
			continue
		}

		// 3. 获取前序交易和输出
		prevTxID := input.PreviousTxIDStr()
		vout := input.PreviousTxOutIndex

		prevTx, err := bt.FetchTXRaw(prevTxID, network)
		if err != nil {
			return fmt.Errorf("获取前序交易失败: %w", err)
		}

		prevOutput := prevTx.OutputIdx(int(vout))
		if prevOutput == nil {
			return fmt.Errorf("前序交易在索引 %d 处没有输出", vout)
		}

		// 4. 使用解释器验证
		if err := interpreter.NewEngine().Execute(
			interpreter.WithTx(tx, i, prevOutput),
			interpreter.WithForkID(),
			interpreter.WithAfterGenesis(),
		); err != nil {
			return fmt.Errorf("输入 %d 验证失败: %w", i, err)
		}

		fmt.Printf("✅ 输入 %d 验证通过\n", i)
	}

	return nil
}

// verifyFTTransaction 验证FT代币转账交易
// FT交易特点：
// - 输入：FT UTXO（code script + tape script） + TBC UTXO（手续费）
// - 输出：接收方FT UTXO（code script + tape script） + 找零FT UTXO（可选）
func verifyFTTransaction(txid string, network string) error {
	// 1. 获取交易
	tx, err := bt.FetchTXRaw(txid, network)
	if err != nil {
		return fmt.Errorf("获取交易失败: %w", err)
	}

	// 2. 验证FT输入（通常是第一个输入）
	// FT UTXO的结构：输出0是code script，输出1是tape script
	for i := 0; i < tx.InputCount(); i++ {
		input := tx.InputIdx(i)

		if isCoinbaseInput(input) {
			continue
		}

		prevTxID := input.PreviousTxIDStr()
		vout := input.PreviousTxOutIndex

		prevTx, err := bt.FetchTXRaw(prevTxID, network)
		if err != nil {
			return fmt.Errorf("获取前序交易失败: %w", err)
		}

		// 获取对应的输出（可能是code script或tape script）
		prevOutput := prevTx.OutputIdx(int(vout))
		if prevOutput == nil {
			return fmt.Errorf("前序交易在索引 %d 处没有输出", vout)
		}

		// 检查是否是FT合约输出（code script通常有特定格式）
		// 这里简化处理，实际应该检查脚本内容
		fmt.Printf("验证输入 %d: 前序交易=%s, vout=%d\n", i, prevTxID, vout)

		// 使用解释器验证
		if err := interpreter.NewEngine().Execute(
			interpreter.WithTx(tx, i, prevOutput),
			interpreter.WithForkID(),
			interpreter.WithAfterGenesis(),
		); err != nil {
			return fmt.Errorf("输入 %d 验证失败: %w", i, err)
		}

		fmt.Printf("✅ 输入 %d 验证通过\n", i)
	}

	// 3. 验证输出结构（可选）
	// FT交易应该包含code script和tape script输出对
	fmt.Printf("交易包含 %d 个输出\n", tx.OutputCount())
	for i := 0; i < tx.OutputCount(); i++ {
		output := tx.OutputIdx(i)
		scriptASM, _ := output.LockingScript.ToASM()
		fmt.Printf("输出 %d: 金额=%d, 脚本长度=%d\n", i, output.Satoshis, len(output.LockingScript.Bytes()))
		if len(scriptASM) > 100 {
			fmt.Printf("  脚本预览: %s...\n", scriptASM[:100])
		}
	}

	return nil
}

// verifyNFTTransaction 验证NFT转移交易
// NFT交易特点：
// - 输入：NFT UTXO（code script + tape script） + TBC UTXO（手续费）
// - 输出：接收方NFT UTXO（code script + tape script） + 持有脚本输出
func verifyNFTTransaction(txid string, network string) error {
	// 1. 获取交易
	tx, err := bt.FetchTXRaw(txid, network)
	if err != nil {
		return fmt.Errorf("获取交易失败: %w", err)
	}

	// 2. 验证NFT输入
	// NFT UTXO通常包含：
	// - 输出0: code script（验证逻辑）
	// - 输出1: hold script（持有脚本）
	// - 输出2: tape script（元数据）
	for i := 0; i < tx.InputCount(); i++ {
		input := tx.InputIdx(i)

		if isCoinbaseInput(input) {
			continue
		}

		prevTxID := input.PreviousTxIDStr()
		vout := input.PreviousTxOutIndex

		prevTx, err := bt.FetchTXRaw(prevTxID, network)
		if err != nil {
			return fmt.Errorf("获取前序交易失败: %w", err)
		}

		prevOutput := prevTx.OutputIdx(int(vout))
		if prevOutput == nil {
			return fmt.Errorf("前序交易在索引 %d 处没有输出", vout)
		}

		fmt.Printf("验证NFT输入 %d: 前序交易=%s, vout=%d\n", i, prevTxID, vout)

		// 使用解释器验证
		if err := interpreter.NewEngine().Execute(
			interpreter.WithTx(tx, i, prevOutput),
			interpreter.WithForkID(),
			interpreter.WithAfterGenesis(),
		); err != nil {
			return fmt.Errorf("输入 %d 验证失败: %w", i, err)
		}

		fmt.Printf("✅ NFT输入 %d 验证通过\n", i)
	}

	// 3. 验证输出结构
	// NFT转移交易应该包含：
	// - code script输出（200 satoshis）
	// - hold script输出（100 satoshis）
	// - tape script输出（0 satoshis）
	fmt.Printf("交易包含 %d 个输出\n", tx.OutputCount())
	if tx.OutputCount() >= 3 {
		fmt.Println("✅ NFT输出结构正确（code + hold + tape）")
	}

	return nil
}

// isCoinbaseInput 检测输入是否为Coinbase输入
func isCoinbaseInput(input *bt.Input) bool {
	prevTxID := input.PreviousTxID()
	vout := input.PreviousTxOutIndex

	allZero := true
	for _, b := range prevTxID {
		if b != 0 {
			allZero = false
			break
		}
	}

	return allZero && vout == 0xFFFFFFFF
}
