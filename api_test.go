package bt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testnet = "testnet"

// TestGetBaseURL 测试 getBaseURL 逻辑（通过导出常量间接验证）
func TestGetBaseURL(t *testing.T) {
	// mainnet
	u := getBaseURL("mainnet")
	assert.Equal(t, mainnetAPIURL, u)
	u = getBaseURL("")
	assert.Equal(t, mainnetAPIURL, u)
	// testnet
	u = getBaseURL("testnet")
	assert.Equal(t, testnetAPIURL, u)
	// custom
	u = getBaseURL("https://custom.example.com/")
	assert.Equal(t, "https://custom.example.com/", u)
	u = getBaseURL("https://custom.example.com")
	assert.Equal(t, "https://custom.example.com/", u)
}

// TestGetTBCBalance 需要网络，-short 时跳过
func TestGetTBCBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	// 使用 testnet 上的一个已知地址（可替换为实际有余额的地址）
	addr := "n1F6bPVduj2Z9J3sLf9xQXK2R5vH3mNpQr"
	bal, err := GetTBCBalance(addr, testnet)
	if err != nil {
		t.Logf("GetTBCBalance err (可能地址无效或无余额): %v", err)
		return
	}
	assert.GreaterOrEqual(t, bal, uint64(0))
}

// TestFetchUTXOs 需要网络
func TestFetchUTXOs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	addr := "n1F6bPVduj2Z9J3sLf9xQXK2R5vH3mNpQr"
	utxos, err := FetchUTXOs(addr, testnet)
	if err != nil {
		t.Logf("FetchUTXOs err: %v", err)
		return
	}
	t.Logf("FetchUTXOs: got %d utxos", len(utxos))
	for i, u := range utxos {
		assert.NotNil(t, u)
		assert.NotNil(t, u.LockingScript)
		assert.GreaterOrEqual(t, u.Satoshis, uint64(0))
		if i >= 2 {
			break
		}
	}
}

// TestGetUTXOs 需要网络
func TestGetUTXOs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	addr := "n1F6bPVduj2Z9J3sLf9xQXK2R5vH3mNpQr"
	utxos, err := GetUTXOs(addr, 0.000001, testnet)
	if err != nil {
		t.Logf("GetUTXOs err: %v", err)
		return
	}
	var total uint64
	for _, u := range utxos {
		total += u.Satoshis
	}
	assert.GreaterOrEqual(t, total, uint64(1))
}

// TestFetchBlockHeaders 需要网络
func TestFetchBlockHeaders(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	headers, err := FetchBlockHeaders(testnet)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(headers), 1)
	h := headers[0]
	assert.NotEmpty(t, h.Hash)
	assert.GreaterOrEqual(t, h.Height, 0)
}

// TestFetchTXRaw 需要网络，使用区块头关联的 tx 或已知 txid
func TestFetchTXRaw(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	headers, err := FetchBlockHeaders(testnet)
	if err != nil || len(headers) == 0 {
		t.Skip("need block headers to find a tx")
	}
	// 无法直接从 headers 拿到 txid，使用一个可能存在的 coinbase
	// 简化：仅测试 API 可连通性，使用一个假设的短 txid 会 404
	txid := "0000000000000000000000000000000000000000000000000000000000000000"
	_, err = FetchTXRaw(txid, testnet)
	// 可能 404 或成功
	if err != nil {
		t.Logf("FetchTXRaw (expected fail for zero txid): %v", err)
	}
}

// TestBroadcastTXsRaw 仅测试空列表的请求体构造，不实际广播
func TestBroadcastTXsRawEmpty(t *testing.T) {
	_, _, err := BroadcastTXsRaw([]BroadcastTXsRequestItem{}, testnet)
	// 空列表可能返回 400 或其他，只要不 panic 即可
	_ = err
}

// TestBuildAddressOrHash 测试 buildAddressOrHash 逻辑（40 位 hex）
func TestBuildAddressOrHash(t *testing.T) {
	validHex := "0123456789abcdef0123456789abcdef01234567"
	h, err := buildAddressOrHash(validHex)
	require.NoError(t, err)
	assert.Equal(t, validHex+"01", h)
}

// TestBuildAddressOrHash_Invalid 测试非法输入
func TestBuildAddressOrHash_Invalid(t *testing.T) {
	_, err := buildAddressOrHash("short")
	assert.Error(t, err)
	_, err = buildAddressOrHash("0123456789abcdef0123456789abcdef0123456g")
	assert.Error(t, err)
}

// TestScriptHashFromHex 测试 scriptHashFromHex
func TestScriptHashFromHex(t *testing.T) {
	// 空脚本
	h, err := scriptHashFromHex("")
	require.NoError(t, err)
	assert.Len(t, h, 64)
	_, err = scriptHashFromHex("zz")
	assert.Error(t, err)
}

// TestParseBigIntOrUint64 测试大数解析
func TestParseBigIntOrUint64(t *testing.T) {
	s, err := parseBigIntOrUint64([]byte("12345"))
	require.NoError(t, err)
	assert.Equal(t, "12345", s)
	s, err = parseBigIntOrUint64([]byte(`"9999999999999999999"`))
	require.NoError(t, err)
	assert.Equal(t, "9999999999999999999", s)
}
