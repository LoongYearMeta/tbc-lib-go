package bt

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/libsv/go-bk/crypto"
	"github.com/sCrypt-Inc/go-bt/v2/util/partialsha256"
)

const (
	ftVersion   = 10
	ftHashLen   = 32
	ftAmountLen = 8
)

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
	buf = append(buf, 0x10, 0x00)
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
	buf = append(buf, 0x20, 0x00)
	buf = append(buf, hash1...)
	buf = append(buf, 0x20, 0x00)
	buf = append(buf, hash2...)

	o1, o1len, o2, o2len := getPrePreOutputsData(tx, vout)
	buf = append(buf, o1len...)
	buf = append(buf, o1...)

	lockScript := tx.Outputs[vout].LockingScript.Bytes()
	scriptLen := len(lockScript)
	sat := make([]byte, 8)
	binary.LittleEndian.PutUint64(sat, tx.Outputs[vout].Satoshis)
	buf = append(buf, 0x08, 0x00)
	buf = append(buf, sat...)

	if scriptLen == 1564 {
		suffix := lockScript[1536:]
		partialHash := partialsha256.CalculatePartialHash(lockScript[:1536])
		ph, _ := hex.DecodeString(partialHash)
		buf = append(buf, getLengthHex(len(suffix))...)
		buf = append(buf, suffix...)
		buf = append(buf, 0x20, 0x00)
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
			buf = append(buf, 0x20, 0x00)
			buf = append(buf, ph...)
		}
		buf = append(buf, getLengthHex(len(getSize(scriptLen)))...)
		buf = append(buf, getSize(scriptLen)...)
	}

	buf = append(buf, o2len...)
	buf = append(buf, o2...)

	return hex.EncodeToString(buf) + "52", nil
}
