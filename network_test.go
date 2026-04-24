package tbc

import (
	"testing"
)

func TestNetwork_Basic(t *testing.T) {
	// 测试预定义网络
	if Livenet == nil {
		t.Fatal("Livenet is nil")
	}
	if Testnet == nil {
		t.Fatal("Testnet is nil")
	}
	if Regtest == nil {
		t.Fatal("Regtest is nil")
	}
	if STN == nil {
		t.Fatal("STN is nil")
	}

	// 测试网络名称
	if Livenet.Name != "livenet" {
		t.Errorf("Livenet.Name = %q, expected %q", Livenet.Name, "livenet")
	}
	if Testnet.Name != "testnet" {
		t.Errorf("Testnet.Name = %q, expected %q", Testnet.Name, "testnet")
	}
}

func TestGetNetwork(t *testing.T) {
	// 按名称获取
	net, ok := GetNetwork("livenet")
	if !ok {
		t.Fatal("GetNetwork(\"livenet\") failed")
	}
	if net != Livenet {
		t.Error("GetNetwork returned wrong network")
	}

	// 按别名获取
	net, ok = GetNetwork("mainnet")
	if !ok {
		t.Fatal("GetNetwork(\"mainnet\") failed")
	}
	if net != Livenet {
		t.Error("GetNetwork returned wrong network")
	}

	// 按魔数获取
	magic := Livenet.NetworkMagicUint32()
	net, ok = GetNetwork(magic)
	if !ok {
		t.Fatal("GetNetwork by magic failed")
	}
	if net != Livenet {
		t.Error("GetNetwork by magic returned wrong network")
	}
}

func TestNetwork_IsMainnet(t *testing.T) {
	if !Livenet.IsMainnet() {
		t.Error("Livenet.IsMainnet() should return true")
	}
	if Testnet.IsMainnet() {
		t.Error("Testnet.IsMainnet() should return false")
	}
	if Regtest.IsMainnet() {
		t.Error("Regtest.IsMainnet() should return false")
	}
}

func TestAddNetwork(t *testing.T) {
	customNet := &Network{
		Name:         "custom",
		Alias:        "customnet",
		PubKeyHash:   0x42,
		PrivateKey:   0xc2,
		ScriptHash:   0x05,
		XPubKey:      0x0488b21e,
		XPrivKey:     0x0488ade4,
		NetworkMagic: [4]byte{0x01, 0x02, 0x03, 0x04},
		Port:         9999,
		DNSServers:   []string{"custom.example.com"},
		CashAddrPref: "custom",
	}

	err := AddNetwork(customNet)
	if err != nil {
		t.Fatalf("AddNetwork failed: %v", err)
	}

	// 验证可以获取
	net, ok := GetNetwork("custom")
	if !ok {
		t.Fatal("GetNetwork(\"custom\") failed after AddNetwork")
	}
	if net != customNet {
		t.Error("GetNetwork returned wrong network")
	}

	// 清理
	_ = RemoveNetwork(customNet)
}
