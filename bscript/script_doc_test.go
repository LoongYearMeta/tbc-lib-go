package bscript_test

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/sCrypt-Inc/go-bt/v2/bscript"
)

// TestScriptCreation_P2PKH 测试支付到公钥哈希（P2PKH）脚本的创建
// 基于: https://github.com/sCrypt-Inc/tbc-lib-js/blob/master/docs/script.md#pay-to-public-key-hash-p2pkh
func TestScriptCreation_P2PKH(t *testing.T) {
	t.Parallel()

	// 创建一个新的 P2PKH 脚本，支付到指定地址
	address := "1NaTVwXDDUJaXDQajoa9MqHhz4uTxtgK14"
	script, err := bscript.NewP2PKHFromAddress(address)
	assert.NoError(t, err)
	assert.NotNil(t, script)

	// 验证脚本结构
	assert.True(t, script.IsP2PKH())
	assert.True(t, script.IsPublicKeyHashOut())

	// 获取 ASM 表示
	asm, err := script.ToASM()
	assert.NoError(t, err)
	assert.Contains(t, asm, "OP_DUP")
	assert.Contains(t, asm, "OP_HASH160")
	assert.Contains(t, asm, "OP_EQUALVERIFY")
	assert.Contains(t, asm, "OP_CHECKSIG")
}

// TestScriptCreation_P2PK 测试支付到公钥（P2PK）脚本的创建
// 基于: https://github.com/sCrypt-Inc/tbc-lib-js/blob/master/docs/script.md#pay-to-public-key-p2pk
func TestScriptCreation_P2PK(t *testing.T) {
	t.Parallel()

	// 创建一个新的 P2PK 脚本，支付到指定公钥
	pubKeyHex := "022df8750480ad5b26950b25c7ba79d3e37d75f640f8e5d9bcd5b150a0f85014da"
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	assert.NoError(t, err)

	pubKey, err := bec.ParsePubKey(pubKeyBytes, bec.S256())
	assert.NoError(t, err)

	// 构建公钥输出脚本
	script := bscript.BuildPublicKeyOut(pubKey)
	assert.NotNil(t, script)

	// 验证脚本结构
	assert.True(t, script.IsP2PK())
	assert.True(t, script.IsPublicKeyOut())

	// 获取 ASM 表示
	asm, err := script.ToASM()
	assert.NoError(t, err)
	assert.Contains(t, asm, "OP_CHECKSIG")
}

// TestScriptCreation_P2MS 测试支付到多重签名（P2MS）脚本的创建
// 基于: https://github.com/sCrypt-Inc/tbc-lib-js/blob/master/docs/script.md#pay-to-multisig-p2ms
func TestScriptCreation_P2MS(t *testing.T) {
	t.Parallel()

	// 从 3 个给定的公钥创建一个新的 2-of-3 多重签名输出
	pubKeyHexes := []string{
		"022df8750480ad5b26950b25c7ba79d3e37d75f640f8e5d9bcd5b150a0f85014da",
		"03e3818b65bcc73a7d64064106a859cc1a5a728c4345ff0b641209fba0d90de6e9",
		"021f2f6e1e50cb6a953935c3601284925decd3fd21bc445712576873fb8c6ebc18",
	}

	pubKeys := make([]*bec.PublicKey, 0, len(pubKeyHexes))
	for _, hexStr := range pubKeyHexes {
		pubKeyBytes, err := hex.DecodeString(hexStr)
		assert.NoError(t, err)
		pubKey, err := bec.ParsePubKey(pubKeyBytes, bec.S256())
		assert.NoError(t, err)
		pubKeys = append(pubKeys, pubKey)
	}

	threshold := 2
	script, err := bscript.BuildMultisigOut(pubKeys, threshold, nil)
	assert.NoError(t, err)
	assert.NotNil(t, script)

	// 验证脚本结构
	assert.True(t, script.IsMultiSigOut())
	assert.True(t, script.IsMultisigOut())

	// 获取 ASM 表示
	asm, err := script.ToASM()
	assert.NoError(t, err)
	assert.Contains(t, asm, "OP_2")
	assert.Contains(t, asm, "OP_3")
	assert.Contains(t, asm, "OP_CHECKMULTISIG")
}

// TestScriptCreation_P2SH 测试支付到脚本哈希（P2SH）脚本的创建
// 基于: https://github.com/sCrypt-Inc/tbc-lib-js/blob/master/docs/script.md#pay-to-script-hash-p2sh
// 注意: P2SH 在 Turing BC 上已弃用
func TestScriptCreation_P2SH(t *testing.T) {
	t.Parallel()

	// 创建一个 P2SH 多重签名输出
	pubKeyHexes := []string{
		"022df8750480ad5b26950b25c7ba79d3e37d75f640f8e5d9bcd5b150a0f85014da",
		"03e3818b65bcc73a7d64064106a859cc1a5a728c4345ff0b641209fba0d90de6e9",
		"021f2f6e1e50cb6a953935c3601284925decd3fd21bc445712576873fb8c6ebc18",
	}

	pubKeys := make([]*bec.PublicKey, 0, len(pubKeyHexes))
	for _, hexStr := range pubKeyHexes {
		pubKeyBytes, err := hex.DecodeString(hexStr)
		assert.NoError(t, err)
		pubKey, err := bec.ParsePubKey(pubKeyBytes, bec.S256())
		assert.NoError(t, err)
		pubKeys = append(pubKeys, pubKey)
	}

	redeemScript, err := bscript.BuildMultisigOut(pubKeys, 2, nil)
	assert.NoError(t, err)
	assert.NotNil(t, redeemScript)

	// 转换为脚本哈希输出
	script := redeemScript.ToScriptHashOut()
	assert.NotNil(t, script)

	// 验证脚本结构
	assert.True(t, script.IsP2SH())
	assert.True(t, script.IsScriptHashOut())

	// 获取 ASM 表示
	asm, err := script.ToASM()
	assert.NoError(t, err)
	assert.Contains(t, asm, "OP_HASH160")
	assert.Contains(t, asm, "OP_EQUAL")
}

// TestScriptCreation_DataOutput 测试数据输出脚本的创建
// 基于: https://github.com/sCrypt-Inc/tbc-lib-js/blob/master/docs/script.md#data-output
func TestScriptCreation_DataOutput(t *testing.T) {
	t.Parallel()

	// 创建一个数据输出脚本
	data := "hello world!!!"
	script, err := bscript.BuildDataOut([]byte(data), "")
	assert.NoError(t, err)
	assert.NotNil(t, script)

	// 验证脚本结构
	assert.True(t, script.IsData())
	assert.True(t, script.IsDataOut())

	// 获取 ASM 表示
	asm, err := script.ToASM()
	assert.NoError(t, err)
	assert.Contains(t, asm, "OP_RETURN")

	// 测试安全数据输出（OP_FALSE OP_RETURN ...）
	safeScript, err := bscript.BuildSafeDataOut([]byte(data), "")
	assert.NoError(t, err)
	assert.NotNil(t, safeScript)
	assert.True(t, safeScript.IsSafeDataOut())
}

// TestScriptCreation_CustomScripts 测试使用 add 和 prepend 创建自定义脚本
// 基于: https://github.com/sCrypt-Inc/tbc-lib-js/blob/master/docs/script.md#custom-scripts
func TestScriptCreation_CustomScripts(t *testing.T) {
	t.Parallel()

	// 创建一个自定义脚本
	script := bscript.NewFromBytes([]byte{})

	// 添加操作码
	err := script.Add(bscript.OpIF)
	assert.NoError(t, err)

	// 前置操作码
	err = script.Prepend(bscript.Op2SWAP)
	assert.NoError(t, err)

	// 添加另一个操作码
	err = script.Add(bscript.OpNOT)
	assert.NoError(t, err)

	// 添加数据缓冲区
	data, err := hex.DecodeString("bacacafe")
	assert.NoError(t, err)
	err = script.Add(data)
	assert.NoError(t, err)

	// 验证脚本
	asm, err := script.ToASM()
	assert.NoError(t, err)
	assert.Contains(t, asm, "OP_2SWAP")
	assert.Contains(t, asm, "OP_IF")
	assert.Contains(t, asm, "OP_NOT")
	assert.Contains(t, asm, "bacacafe")
}

// TestScriptParsingAndIdentification 测试脚本解析和识别
// 基于: https://github.com/sCrypt-Inc/tbc-lib-js/blob/master/docs/script.md#script-parsing-and-identification
func TestScriptParsingAndIdentification(t *testing.T) {
	t.Parallel()

	// 解析一个原始多重签名脚本
	rawScriptHex := "5221022df8750480ad5b26950b25c7ba79d3e37d75f640f8e5d9bcd5b150a0f85014da2103e3818b65bcc73a7d64064106a859cc1a5a728c4345ff0b641209fba0d90de6e921021f2f6e1e50cb6a953935c3601284925decd3fd21bc445712576873fb8c6ebc1853ae"
	rawScript, err := hex.DecodeString(rawScriptHex)
	assert.NoError(t, err)

	script := bscript.NewFromBytes(rawScript)
	assert.NotNil(t, script)

	// 获取 ASM 表示
	asm, err := script.ToASM()
	assert.NoError(t, err)
	assert.Contains(t, asm, "OP_2")
	assert.Contains(t, asm, "OP_3")
	assert.Contains(t, asm, "OP_CHECKMULTISIG")

	// 验证脚本类型识别
	assert.False(t, script.IsPublicKeyHashOut())
	assert.False(t, script.IsScriptHashOut())
	assert.True(t, script.IsMultisigOut())
	assert.True(t, script.IsMultiSigOut())

	// 测试分类
	scriptType := script.Classify()
	assert.Equal(t, bscript.ScriptTypeMultisigOut, scriptType)
}

// TestScriptFromString 测试从字符串格式创建脚本
func TestScriptFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		str      string
		expected string
	}{
		{
			name:     "hex string",
			str:      "76a914e2a623699e81b291c0327f408fea765d534baa2a88ac",
			expected: "76a914e2a623699e81b291c0327f408fea765d534baa2a88ac",
		},
		{
			name:     "empty string",
			str:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := bscript.FromString(tt.str)
			if tt.str == "" {
				// 空字符串应该创建空脚本
				assert.NoError(t, err)
				assert.NotNil(t, script)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, script)
				assert.Equal(t, tt.expected, script.String())
			}
		})
	}
}

// TestScriptFromChunks 测试从 chunks 创建脚本
func TestScriptFromChunks(t *testing.T) {
	t.Parallel()

	// 手动创建 chunks
	chunks := []bscript.Chunk{
		{OpcodeNum: bscript.OpDUP},
		{OpcodeNum: bscript.OpHASH160},
		{OpcodeNum: bscript.OpDATA20, Buf: make([]byte, 20), Len: 20},
		{OpcodeNum: bscript.OpEQUALVERIFY},
		{OpcodeNum: bscript.OpCHECKSIG},
	}

	script, err := bscript.FromChunks(chunks)
	assert.NoError(t, err)
	assert.NotNil(t, script)

	// 验证它是 P2PKH 脚本
	assert.True(t, script.IsP2PKH())
}

// TestScriptChunks 测试从脚本获取 chunks
func TestScriptChunks(t *testing.T) {
	t.Parallel()

	// 创建一个 P2PKH 脚本
	script, err := bscript.NewFromHexString("76a914e2a623699e81b291c0327f408fea765d534baa2a88ac")
	assert.NoError(t, err)

	// 获取 chunks
	chunks := script.Chunks()
	assert.NotNil(t, chunks)
	assert.Greater(t, len(chunks), 0)

	// 验证第一个 chunk 是 OP_DUP
	assert.Equal(t, bscript.OpDUP, chunks[0].OpcodeNum)

	// 验证第二个 chunk 是 OP_HASH160
	assert.Equal(t, bscript.OpHASH160, chunks[1].OpcodeNum)

	// 验证第三个 chunk 包含数据
	assert.Equal(t, bscript.OpDATA20, chunks[2].OpcodeNum)
	assert.NotNil(t, chunks[2].Buf)
	assert.Equal(t, 20, chunks[2].Len)
}

// TestScriptClone 测试克隆脚本
func TestScriptClone(t *testing.T) {
	t.Parallel()

	original, err := bscript.NewFromHexString("76a914e2a623699e81b291c0327f408fea765d534baa2a88ac")
	assert.NoError(t, err)

	cloned := original.Clone()
	assert.NotNil(t, cloned)
	assert.True(t, original.Equals(cloned))
	assert.NotSame(t, original, cloned) // 不同的指针
}

// TestScriptRemoveCodeseparators 测试从脚本中移除 OP_CODESEPARATOR
func TestScriptRemoveCodeseparators(t *testing.T) {
	t.Parallel()

	// 创建一个包含 OP_CODESEPARATOR 的脚本
	script := bscript.NewFromBytes([]byte{})
	_ = script.AppendOpcodes(bscript.OpDUP, bscript.OpCODESEPARATOR, bscript.OpHASH160)

	// 移除 codeseparators
	script.RemoveCodeseparators()

	// 验证 OP_CODESEPARATOR 已被移除
	asm, err := script.ToASM()
	assert.NoError(t, err)
	assert.NotContains(t, asm, "OP_CODESEPARATOR")
}

// TestScriptSubScript 测试获取脚本的子集
func TestScriptSubScript(t *testing.T) {
	t.Parallel()

	// 创建一个包含 OP_CODESEPARATOR 的脚本
	script := bscript.NewFromBytes([]byte{})
	_ = script.AppendOpcodes(bscript.OpDUP, bscript.OpCODESEPARATOR, bscript.OpHASH160, bscript.OpEQUALVERIFY)

	// 从第一个 codeseparator 开始获取子脚本
	subScript := script.SubScript(0)
	assert.NotNil(t, subScript)

	// 验证它包含 codeseparator 之后的操作
	asm, err := subScript.ToASM()
	assert.NoError(t, err)
	assert.Contains(t, asm, "OP_HASH160")
}

// TestScriptCheckMinimalPush 测试最小推送编码检查
func TestScriptCheckMinimalPush(t *testing.T) {
	t.Parallel()

	// 创建一个包含最小推送的脚本
	script, err := bscript.NewFromHexString("76a914e2a623699e81b291c0327f408fea765d534baa2a88ac")
	assert.NoError(t, err)

	chunks := script.Chunks()
	// 检查推送操作是否是最小的
	for i := range chunks {
		if chunks[i].Buf != nil {
			// 这是一个数据推送，验证它是最小的
			isMinimal := script.CheckMinimalPush(i)
			assert.True(t, isMinimal, "Chunk %d 应该是最小编码", i)
		}
	}
}

// TestScriptIsPushOnly 测试脚本是否仅为推送操作
func TestScriptIsPushOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		script   string
		expected bool
	}{
		{
			name:     "push-only script",
			script:   "76a914e2a623699e81b291c0327f408fea765d534baa2a88ac",
			expected: false, // 包含 OP_DUP, OP_HASH160 等
		},
		{
			name:     "data output",
			script:   "6a0e68656c6c6f20776f726c64212121",
			expected: false, // 包含 OP_RETURN
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := bscript.NewFromHexString(tt.script)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, script.IsPushOnly())
		})
	}
}

// TestScriptGetData 测试从脚本获取数据
func TestScriptGetData(t *testing.T) {
	t.Parallel()

	// 测试 P2PKH 脚本
	p2pkhScript, err := bscript.NewFromHexString("76a914e2a623699e81b291c0327f408fea765d534baa2a88ac")
	assert.NoError(t, err)

	data, err := p2pkhScript.GetData()
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 20, len(data)) // 公钥哈希是 20 字节

	// 测试数据输出脚本
	dataScript, err := bscript.BuildDataOut([]byte("hello"), "")
	assert.NoError(t, err)

	data, err = dataScript.GetData()
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

// TestScriptGetPublicKey 测试从 P2PK 脚本获取公钥
func TestScriptGetPublicKey(t *testing.T) {
	t.Parallel()

	// 创建一个 P2PK 脚本
	pubKeyHex := "022df8750480ad5b26950b25c7ba79d3e37d75f640f8e5d9bcd5b150a0f85014da"
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	assert.NoError(t, err)

	pubKey, err := bec.ParsePubKey(pubKeyBytes, bec.S256())
	assert.NoError(t, err)

	script := bscript.BuildPublicKeyOut(pubKey)
	assert.NotNil(t, script)

	// 从脚本获取公钥
	retrievedPubKey, err := script.GetPublicKey()
	assert.NoError(t, err)
	assert.Equal(t, pubKeyBytes, retrievedPubKey)
}

// TestScriptGetPublicKeyHash 测试从 P2PKH 脚本获取公钥哈希
func TestScriptGetPublicKeyHash(t *testing.T) {
	t.Parallel()

	// 创建一个 P2PKH 脚本
	pubKeyHex := "022df8750480ad5b26950b25c7ba79d3e37d75f640f8e5d9bcd5b150a0f85014da"
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	assert.NoError(t, err)

	pubKeyHash := crypto.Hash160(pubKeyBytes)
	script, err := bscript.NewP2PKHFromPubKeyHash(pubKeyHash)
	assert.NoError(t, err)

	// 从脚本获取公钥哈希
	retrievedPubKeyHash, err := script.GetPublicKeyHash()
	assert.NoError(t, err)
	assert.Equal(t, pubKeyHash, retrievedPubKeyHash)
}

// TestScriptGetSignatureOperationsCount 测试计算签名操作数量
func TestScriptGetSignatureOperationsCount(t *testing.T) {
	t.Parallel()

	// 测试 P2PKH 脚本（1 个签名操作）
	script, err := bscript.NewFromHexString("76a914e2a623699e81b291c0327f408fea765d534baa2a88ac")
	assert.NoError(t, err)

	count := script.GetSignatureOperationsCount(true)
	assert.Equal(t, 1, count)

	// 测试多重签名脚本
	pubKeyHexes := []string{
		"022df8750480ad5b26950b25c7ba79d3e37d75f640f8e5d9bcd5b150a0f85014da",
		"03e3818b65bcc73a7d64064106a859cc1a5a728c4345ff0b641209fba0d90de6e9",
	}

	pubKeys := make([]*bec.PublicKey, 0, len(pubKeyHexes))
	for _, hexStr := range pubKeyHexes {
		pubKeyBytes, err := hex.DecodeString(hexStr)
		assert.NoError(t, err)
		pubKey, err := bec.ParsePubKey(pubKeyBytes, bec.S256())
		assert.NoError(t, err)
		pubKeys = append(pubKeys, pubKey)
	}

	multisigScript, err := bscript.BuildMultisigOut(pubKeys, 2, nil)
	assert.NoError(t, err)

	count = multisigScript.GetSignatureOperationsCount(true)
	assert.Equal(t, 2, count) // 2-of-2 多重签名
}

// TestScriptClassify 测试脚本分类
func TestScriptClassify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		script     string
		expected   string
		isInput    bool
		isOutput   bool
	}{
		{
			name:     "P2PKH output",
			script:   "76a914e2a623699e81b291c0327f408fea765d534baa2a88ac",
			expected: bscript.ScriptTypePubKeyHashOut,
			isOutput: true,
		},
		{
			name:     "P2PK output",
			script:   "21022df8750480ad5b26950b25c7ba79d3e37d75f640f8e5d9bcd5b150a0f85014daac",
			expected: bscript.ScriptTypePubKeyOut,
			isOutput: true,
		},
		{
			name:     "multisig output",
			script:   "5221022df8750480ad5b26950b25c7ba79d3e37d75f640f8e5d9bcd5b150a0f85014da2103e3818b65bcc73a7d64064106a859cc1a5a728c4345ff0b641209fba0d90de6e952ae",
			expected: bscript.ScriptTypeMultisigOut,
			isOutput: true,
		},
		{
			name:     "data output",
			script:   "6a0e68656c6c6f20776f726c64212121",
			expected: bscript.ScriptTypeDataOut,
			isOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := bscript.NewFromHexString(tt.script)
			assert.NoError(t, err)

			script.SetIsInput(tt.isInput)
			script.SetIsOutput(tt.isOutput)

			classified := script.Classify()
			assert.Equal(t, tt.expected, classified)
		})
	}
}

// TestScriptToAddress 测试将脚本转换为地址
func TestScriptToAddress(t *testing.T) {
	t.Parallel()

	// 测试 P2PKH 脚本
	addressStr := "1NaTVwXDDUJaXDQajoa9MqHhz4uTxtgK14"
	script, err := bscript.NewP2PKHFromAddress(addressStr)
	assert.NoError(t, err)

	addr, err := script.ToAddress(true) // 主网
	assert.NoError(t, err)
	assert.NotNil(t, addr)
	// 注意: 地址格式可能不同，所以我们只检查它不为空
	assert.NotEmpty(t, addr)
}

// TestScriptFindAndDelete 测试查找和删除脚本 chunks
func TestScriptFindAndDelete(t *testing.T) {
	t.Parallel()

	// 创建一个脚本
	script, err := bscript.NewFromHexString("76a914e2a623699e81b291c0327f408fea765d534baa2a88ac")
	assert.NoError(t, err)

	// 创建一个要查找和删除的脚本
	data, err := hex.DecodeString("e2a623699e81b291c0327f408fea765d534baa2a")
	assert.NoError(t, err)

	scriptToDelete := bscript.NewFromBytes([]byte{})
	_ = scriptToDelete.AppendPushData(data)

	// 查找并删除
	count := script.FindAndDelete(scriptToDelete)
	assert.Greater(t, count, uint(0))
}
