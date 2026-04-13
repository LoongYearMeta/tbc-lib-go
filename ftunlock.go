package bt

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bk/crypto"
	"github.com/sCrypt-Inc/go-bt/v2/util/partialsha256"
)

const (
	ftVersion   = 10
	ftHashLen   = 32
	ftAmountLen = 8

	// 与 tbc-contract/lib/util/ftunlock.ts 一致
	ftV1Length        = 1564
	ftV1PartialOffset = 1536
	ftV2Length        = 1884
	ftV2PartialOffset = 1856
	coinLength        = 2012
	coinPartialOffset = 1984
)

// partialOffsetGetPreTx 对应 getPreTxdata（preTxdata 路径）。
// 与 tbc-contract/lib/util/ftunlock.ts 一致：
// - coin 精确匹配 → coin_partial_offset
// - ft v2 区间 [ft_v2_length, coin_length) → ft_v2_partial_offset
// - 否则默认 ft_v1_partial_offset
func partialOffsetGetPreTx(scriptLen int) int {
	if scriptLen == coinLength {
		return coinPartialOffset
	}
	if scriptLen >= ftV2Length && scriptLen < coinLength {
		return ftV2PartialOffset
	}
	return ftV1PartialOffset
}

// partialOffsetGetPrePre 对应 getPrePreTxdata / getCurrentTxdata 的 FT code+tape 成对分支。
// 与 ftunlock.ts 一致：v1/coin 精确匹配；v2 为区间 [ft_v2_length, coin_length)；否则 off=0 走通用 split。
func partialOffsetGetPrePre(scriptLen int) int {
	switch scriptLen {
	case ftV1Length:
		return ftV1PartialOffset
	case coinLength:
		return coinPartialOffset
	default:
		if scriptLen >= ftV2Length && scriptLen < coinLength {
			return ftV2PartialOffset
		}
		return 0
	}
}

// getLengthHex 返回变长整数编码 (OP_PUSHDATA1/2)
func getLengthHex(length int) []byte {
	if length < 76 {
		return []byte{byte(length)}
	}
	if length < 256 {
		return []byte{0x4c, byte(length)}
	}
	b := make([]byte, 3)
	b[0] = 0x4d
	binary.LittleEndian.PutUint16(b[1:], uint16(length))
	return b
}

// getSize 返回脚本长度的 little-endian 编码
func getSize(length int) []byte {
	if length < 256 {
		return []byte{byte(length)}
	}
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(length))
	return b
}

// getPrePreOutputsData 获取 grandparent 的 outputs1/outputs2
// 与 JS 一致：vout==0 时 outputs1=0x00 无 length；outputs2 空时为 0x00 无 length
func getPrePreOutputsData(tx *Tx, vout int) (outputs1, outputs1len, outputs2, outputs2len []byte) {
	if vout > 0 {
		var buf1 []byte
		for i := 0; i < vout; i++ {
			sat := make([]byte, 8)
			binary.LittleEndian.PutUint64(sat, tx.Outputs[i].Satoshis)
			buf1 = append(buf1, sat...)
			buf1 = append(buf1, crypto.Sha256(tx.Outputs[i].LockingScript.Bytes())...)
		}
		outputs1 = buf1
		outputs1len = getLengthHex(len(buf1))
	} else {
		outputs1 = []byte{0x00}
	}
	var buf2 []byte
	for i := vout + 1; i < len(tx.Outputs); i++ {
		sat := make([]byte, 8)
		binary.LittleEndian.PutUint64(sat, tx.Outputs[i].Satoshis)
		buf2 = append(buf2, sat...)
		buf2 = append(buf2, crypto.Sha256(tx.Outputs[i].LockingScript.Bytes())...)
	}
	if len(buf2) > 0 {
		outputs2 = buf2
		outputs2len = getLengthHex(len(buf2))
	} else {
		outputs2 = []byte{0x00}
	}
	return
}

// GetPrePreTxdata 获取 grandparent 交易的 txdata，用于 FT 解锁
// 对应 JS ftunlock.getPrePreTxdata
func GetPrePreTxdata(tx *Tx, vout int) (string, error) {
	var buf []byte
	// vliolength: 10 (version + nLockTime + inputCount + outputCount 共 16 bytes)
	buf = append(buf, 0x10)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, ftVersion)
	buf = append(buf, b...)
	binary.LittleEndian.PutUint32(b, tx.LockTime)
	buf = append(buf, b...)
	binary.LittleEndian.PutUint32(b, uint32(len(tx.Inputs)))
	buf = append(buf, b...)
	binary.LittleEndian.PutUint32(b, uint32(len(tx.Outputs)))
	buf = append(buf, b...)

	var inputBuf1, inputBuf2 []byte
	for _, in := range tx.Inputs {
		prevID := ReverseBytes(in.PreviousTxID())
		inputBuf1 = append(inputBuf1, prevID...)
		oi := make([]byte, 4)
		binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
		inputBuf1 = append(inputBuf1, oi...)
		binary.LittleEndian.PutUint32(oi, in.SequenceNumber)
		inputBuf1 = append(inputBuf1, oi...)
		scriptHash := crypto.Sha256(in.UnlockingScript.Bytes())
		inputBuf2 = append(inputBuf2, scriptHash...)
	}
	hash1 := crypto.Sha256(inputBuf1)
	hash2 := crypto.Sha256(inputBuf2)
	buf = append(buf, 0x20)
	buf = append(buf, hash1...)
	buf = append(buf, 0x20)
	buf = append(buf, hash2...)

	o1, o1len, o2, o2len := getPrePreOutputsData(tx, vout)
	buf = append(buf, o1len...)
	buf = append(buf, o1...)

	lockScript := tx.Outputs[vout].LockingScript.Bytes()
	scriptLen := len(lockScript)
	sat := make([]byte, 8)
	binary.LittleEndian.PutUint64(sat, tx.Outputs[vout].Satoshis)
	buf = append(buf, 0x08)
	buf = append(buf, sat...)

	if off := partialOffsetGetPrePre(scriptLen); off > 0 {
		suffix := lockScript[off:]
		partialHash := partialsha256.CalculatePartialHash(lockScript[:off])
		ph, _ := hex.DecodeString(partialHash)
		buf = append(buf, getLengthHex(len(suffix))...)
		buf = append(buf, suffix...)
		buf = append(buf, 0x20)
		buf = append(buf, ph...)
		buf = append(buf, getLengthHex(len(getSize(scriptLen)))...)
		buf = append(buf, getSize(scriptLen)...)
	} else {
		var suffix, ph []byte
		if scriptLen < 64 {
			suffix = lockScript
			ph = []byte{0x00}
		} else {
			n := scriptLen / 64
			partialLen := 64 * n
			phStr := partialsha256.CalculatePartialHash(lockScript[:partialLen])
			ph, _ = hex.DecodeString(phStr)
			suffix = lockScript[partialLen:]
		}
		buf = append(buf, getLengthHex(len(suffix))...)
		buf = append(buf, suffix...)
		if len(ph) == 1 {
			buf = append(buf, ph...)
		} else {
			buf = append(buf, 0x20)
			buf = append(buf, ph...)
		}
		buf = append(buf, getLengthHex(len(getSize(scriptLen)))...)
		buf = append(buf, getSize(scriptLen)...)
	}

	buf = append(buf, o2len...)
	buf = append(buf, o2...)

	return hex.EncodeToString(buf) + "52", nil
}

// getPreOutputsData 获取 parent 的 outputs1/outputs2（outputs2 从 vout+2 开始，跳过 code 和 tape）
func getPreOutputsData(tx *Tx, vout int) (outputs1, outputs1len, outputs2, outputs2len []byte) {
	if vout > 0 {
		var buf1 []byte
		for i := 0; i < vout; i++ {
			sat := make([]byte, 8)
			binary.LittleEndian.PutUint64(sat, tx.Outputs[i].Satoshis)
			buf1 = append(buf1, sat...)
			buf1 = append(buf1, crypto.Sha256(tx.Outputs[i].LockingScript.Bytes())...)
		}
		outputs1 = buf1
		outputs1len = getLengthHex(len(buf1))
	} else {
		outputs1 = []byte{0x00}
	}
	var buf2 []byte
	for i := vout + 2; i < len(tx.Outputs); i++ {
		sat := make([]byte, 8)
		binary.LittleEndian.PutUint64(sat, tx.Outputs[i].Satoshis)
		buf2 = append(buf2, sat...)
		buf2 = append(buf2, crypto.Sha256(tx.Outputs[i].LockingScript.Bytes())...)
	}
	if len(buf2) > 0 {
		outputs2 = buf2
		outputs2len = getLengthHex(len(buf2))
	} else {
		outputs2 = []byte{0x00}
		outputs2len = []byte{}
	}
	return
}

// GetPreTxdata 获取 parent 交易的 txdata，用于 FT 解锁
// 对应 JS ftunlock.getPreTxdata
func GetPreTxdata(tx *Tx, vout int) (string, error) {
	if vout+1 >= len(tx.Outputs) {
		return "", fmt.Errorf("vout+1 out of range")
	}
	var buf []byte
	buf = append(buf, 0x10)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, ftVersion)
	buf = append(buf, b...)
	binary.LittleEndian.PutUint32(b, tx.LockTime)
	buf = append(buf, b...)
	binary.LittleEndian.PutUint32(b, uint32(len(tx.Inputs)))
	buf = append(buf, b...)
	binary.LittleEndian.PutUint32(b, uint32(len(tx.Outputs)))
	buf = append(buf, b...)

	var inputBuf1, inputBuf2 []byte
	for _, in := range tx.Inputs {
		prevID := ReverseBytes(in.PreviousTxID())
		inputBuf1 = append(inputBuf1, prevID...)
		oi := make([]byte, 4)
		binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
		inputBuf1 = append(inputBuf1, oi...)
		binary.LittleEndian.PutUint32(oi, in.SequenceNumber)
		inputBuf1 = append(inputBuf1, oi...)
		scriptHash := crypto.Sha256(in.UnlockingScript.Bytes())
		inputBuf2 = append(inputBuf2, scriptHash...)
	}
	buf = append(buf, getLengthHex(len(inputBuf1))...)
	buf = append(buf, inputBuf1...)
	buf = append(buf, 0x20)
	buf = append(buf, crypto.Sha256(inputBuf2)...)

	o1, o1len, o2, o2len := getPreOutputsData(tx, vout)
	buf = append(buf, o1len...)
	buf = append(buf, o1...)

	lockScript := tx.Outputs[vout].LockingScript.Bytes()
	scriptLen := len(lockScript)
	off := partialOffsetGetPreTx(scriptLen)
	if scriptLen < off {
		return "", fmt.Errorf("lock script too short for FT")
	}
	suffix := lockScript[off:]
	partialHash := partialsha256.CalculatePartialHash(lockScript[:off])
	ph, _ := hex.DecodeString(partialHash)

	sat := make([]byte, 8)
	binary.LittleEndian.PutUint64(sat, tx.Outputs[vout].Satoshis)
	buf = append(buf, 0x08)
	buf = append(buf, sat...)
	buf = append(buf, getLengthHex(len(suffix))...)
	buf = append(buf, suffix...)
	buf = append(buf, 0x20)
	buf = append(buf, ph...)
	buf = append(buf, getLengthHex(len(getSize(scriptLen)))...)
	buf = append(buf, getSize(scriptLen)...)

	binary.LittleEndian.PutUint64(sat, tx.Outputs[vout+1].Satoshis)
	buf = append(buf, 0x08)
	buf = append(buf, sat...)
	buf = append(buf, getLengthHex(len(tx.Outputs[vout+1].LockingScript.Bytes()))...)
	buf = append(buf, tx.Outputs[vout+1].LockingScript.Bytes()...)

	buf = append(buf, o2len...)
	buf = append(buf, o2...)

	return hex.EncodeToString(buf), nil
}

// GetCurrentTxdata 获取当前交易的 txdata，用于 FT 解锁
// 对应 JS ftunlock.getCurrentTxdata
func GetCurrentTxdata(tx *Tx, inputIndex int) (string, error) {
	inputIndexMap := map[int]byte{0: 0x00, 1: 0x51, 2: 0x52, 3: 0x53, 4: 0x54, 5: 0x55}
	endTag := byte(0x51)
	var buf []byte
	buf = append(buf, endTag)

	for i := 0; i < len(tx.Outputs); i++ {
		lockScript := tx.Outputs[i].LockingScript.Bytes()
		scriptLen := len(lockScript)

		sat := make([]byte, 8)
		binary.LittleEndian.PutUint64(sat, tx.Outputs[i].Satoshis)
		buf = append(buf, 0x08)
		buf = append(buf, sat...)

		if off := partialOffsetGetPrePre(scriptLen); off > 0 {
			suffix := lockScript[off:]
			partialHash := partialsha256.CalculatePartialHash(lockScript[:off])
			ph, _ := hex.DecodeString(partialHash)
			buf = append(buf, getLengthHex(len(suffix))...)
			buf = append(buf, suffix...)
			buf = append(buf, 0x20)
			buf = append(buf, ph...)
			size := getSize(scriptLen)
			buf = append(buf, getLengthHex(len(size))...)
			buf = append(buf, size...)

			i++
			binary.LittleEndian.PutUint64(sat, tx.Outputs[i].Satoshis)
			buf = append(buf, 0x08)
			buf = append(buf, sat...)
			buf = append(buf, getLengthHex(len(tx.Outputs[i].LockingScript.Bytes()))...)
			buf = append(buf, tx.Outputs[i].LockingScript.Bytes()...)
		} else {
			// 与 JS ftunlock.getCurrentTxdata 对齐：
			// - 当 off==0 时，如果 scriptLen < 64：suffixdata=全量，partialhash='00'
			// - 当 scriptLen >= 64：suffixdata=后半部分(按 64*n 切)，partialhash=partial_sha256(前 64*n)
			// 否则直接整段 push 会导致 stack 对比值错位，触发 OP_EQUALVERIFY。
			var suffixdata []byte
			var suffixPartialHash []byte
			if scriptLen < 64 {
				suffixdata = lockScript
				suffixPartialHash = []byte{0x00} // 与 TS partialhash='00' 对齐
			} else {
				n := scriptLen / 64
				partialLength := 64 * n
				partialHashHex := partialsha256.CalculatePartialHash(lockScript[:partialLength])
				ph, _ := hex.DecodeString(partialHashHex)
				suffixdata = lockScript[partialLength:]
				suffixPartialHash = ph
			}

			buf = append(buf, getLengthHex(len(suffixdata))...)
			buf = append(buf, suffixdata...)

			if len(suffixPartialHash) == 1 && suffixPartialHash[0] == 0x00 {
				// JS：partialhash.length === 2 时不写 0x20，只写 1 字节 0x00
				buf = append(buf, 0x00)
			} else {
				// JS：写 hashlength(0x20) + 32字节 partialhash
				buf = append(buf, 0x20)
				buf = append(buf, suffixPartialHash...)
			}

			size := getSize(scriptLen)
			buf = append(buf, getLengthHex(len(size))...)
			buf = append(buf, size...)
		}
		buf = append(buf, 0x52)
	}

	if idx, ok := inputIndexMap[inputIndex]; ok {
		buf = append(buf, idx)
	}
	return hex.EncodeToString(buf), nil
}

// GetCurrentInputsdata 获取当前交易的 inputs data
// 对应 JS ftunlock.getCurrentInputsdata
func GetCurrentInputsdata(tx *Tx) string {
	var inputBuf []byte
	for _, in := range tx.Inputs {
		prevID := ReverseBytes(in.PreviousTxID())
		inputBuf = append(inputBuf, prevID...)
		oi := make([]byte, 4)
		binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
		inputBuf = append(inputBuf, oi...)
		binary.LittleEndian.PutUint32(oi, in.SequenceNumber)
		inputBuf = append(inputBuf, oi...)
	}
	buf := append(getLengthHex(len(inputBuf)), inputBuf...)
	return hex.EncodeToString(buf)
}

// GetContractTxdata 获取合约交易数据
// 对应 JS ftunlock.getContractTxdata（含 vout<0 时与 FT v2 swap 一致的全输出序列化分支）
func GetContractTxdata(tx *Tx, vout int) (string, error) {
	var buf []byte
	buf = append(buf, 0x10)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, ftVersion)
	buf = append(buf, b...)
	binary.LittleEndian.PutUint32(b, tx.LockTime)
	buf = append(buf, b...)
	binary.LittleEndian.PutUint32(b, uint32(len(tx.Inputs)))
	buf = append(buf, b...)
	binary.LittleEndian.PutUint32(b, uint32(len(tx.Outputs)))
	buf = append(buf, b...)

	var inputBuf1, inputBuf2 []byte
	for _, in := range tx.Inputs {
		prevID := ReverseBytes(in.PreviousTxID())
		inputBuf1 = append(inputBuf1, prevID...)
		oi := make([]byte, 4)
		binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
		inputBuf1 = append(inputBuf1, oi...)
		binary.LittleEndian.PutUint32(oi, in.SequenceNumber)
		inputBuf1 = append(inputBuf1, oi...)
		scriptHash := crypto.Sha256(in.UnlockingScript.Bytes())
		inputBuf2 = append(inputBuf2, scriptHash...)
	}
	buf = append(buf, 0x20)
	buf = append(buf, crypto.Sha256(inputBuf1)...)
	buf = append(buf, 0x20)
	buf = append(buf, crypto.Sha256(inputBuf2)...)

	if vout < 0 {
		for i := 0; i < len(tx.Outputs); i++ {
			buf = append(buf, 0x08)
			sat := make([]byte, 8)
			binary.LittleEndian.PutUint64(sat, tx.Outputs[i].Satoshis)
			buf = append(buf, sat...)
			buf = append(buf, 0x20)
			buf = append(buf, crypto.Sha256(tx.Outputs[i].LockingScript.Bytes())...)
		}
		for i := len(tx.Outputs); i < 15; i++ {
			buf = append(buf, 0x00)
			buf = append(buf, 0x00)
		}
		return hex.EncodeToString(buf), nil
	}

	if vout >= len(tx.Outputs) {
		return "", fmt.Errorf("vout out of range")
	}

	o1, o1len, o2, o2len := getPrePreOutputsData(tx, vout)
	buf = append(buf, o1len...)
	buf = append(buf, o1...)

	buf = append(buf, 0x08)
	sat := make([]byte, 8)
	binary.LittleEndian.PutUint64(sat, tx.Outputs[vout].Satoshis)
	buf = append(buf, sat...)
	buf = append(buf, 0x20)
	buf = append(buf, crypto.Sha256(tx.Outputs[vout].LockingScript.Bytes())...)
	buf = append(buf, o2len...)
	buf = append(buf, o2...)

	return hex.EncodeToString(buf), nil
}

// GetSizeHex 返回脚本长度的 hex，用于 FT mint code 等
// 对应 JS ftunlock.getSize，导出供 contract 使用
func GetSizeHex(length int) string {
	s := getSize(length)
	return hex.EncodeToString(s)
}
