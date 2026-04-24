package tbc

import (
	"bytes"
	"io"
	"testing"
)

func TestReadVarBytes(t *testing.T) {
	// 准备测试数据：varint(5) + "hello"
	data := []byte{0x05, 'h', 'e', 'l', 'l', 'o'}
	r := bytes.NewReader(data)

	result, err := ReadVarBytes(r, 100)
	if err != nil {
		t.Fatalf("ReadVarBytes failed: %v", err)
	}

	expected := []byte("hello")
	if !bytes.Equal(result, expected) {
		t.Errorf("ReadVarBytes = %q, expected %q", result, expected)
	}
}

func TestWriteVarBytes(t *testing.T) {
	data := []byte("hello")
	var buf bytes.Buffer

	err := WriteVarBytes(&buf, data)
	if err != nil {
		t.Fatalf("WriteVarBytes failed: %v", err)
	}

	// 验证写入的内容：varint(5) + "hello"
	expected := []byte{0x05, 'h', 'e', 'l', 'l', 'o'}
	result := buf.Bytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("WriteVarBytes = %v, expected %v", result, expected)
	}
}

func TestReadWriteVarBytes_RoundTrip(t *testing.T) {
	testCases := [][]byte{
		[]byte("hello"),
		[]byte(""),
		make([]byte, 100),
		make([]byte, 255),
		make([]byte, 256),
	}

	for _, tc := range testCases {
		var buf bytes.Buffer

		// 写入
		err := WriteVarBytes(&buf, tc)
		if err != nil {
			t.Fatalf("WriteVarBytes failed for %d bytes: %v", len(tc), err)
		}

		// 读取
		r := bytes.NewReader(buf.Bytes())
		result, err := ReadVarBytes(r, 10000)
		if err != nil {
			t.Fatalf("ReadVarBytes failed for %d bytes: %v", len(tc), err)
		}

		// 验证
		if !bytes.Equal(result, tc) {
			t.Errorf("round trip failed: expected %v, got %v", tc, result)
		}

		// 验证读取完毕
		remaining, _ := io.ReadAll(r)
		if len(remaining) != 0 {
			t.Errorf("extra data after ReadVarBytes: %v", remaining)
		}
	}
}

func TestIsHexString(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"00", true},
		{"ff", true},
		{"FF", true},
		{"aBcD", true},
		{"0123456789abcdef", true},
		{"0", false},      // 奇数长度
		{"abc", false},    // 奇数长度
		{"gh", false},    // 无效字符
		{"0g", false},    // 无效字符
		{"0x00", false},  // 包含 'x'
	}

	for _, tc := range testCases {
		result := IsHexString(tc.input)
		if result != tc.expected {
			t.Errorf("IsHexString(%q) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}
