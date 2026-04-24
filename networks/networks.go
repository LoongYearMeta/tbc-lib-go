package networks

// Network 抽象层（参考 tbc-lib-js 的 lib/networks.js）
//
// 提供统一的网络配置管理，支持 livenet、testnet、regtest、stn 等网络类型

import (
	"encoding/binary"
	"fmt"
)

// Network 表示一个区块链网络配置
type Network struct {
	Name         string   // 网络名称，如 "livenet", "testnet"
	Alias        string   // 别名，如 "mainnet"
	PubKeyHash   byte     // P2PKH 地址版本字节
	PrivateKey   byte     // 私钥 WIF 版本字节
	ScriptHash   byte     // P2SH 地址版本字节
	XPubKey      uint32   // 扩展公钥魔数
	XPrivKey     uint32   // 扩展私钥魔数
	NetworkMagic [4]byte  // P2P 网络魔数（小端序）
	Port         int      // P2P 端口
	DNSServers   []string // DNS 种子节点
	CashAddrPref string   // CashAddr 前缀（如 "bitcoincash"）
}

var (
	// 网络魔数（小端序）
	networkMagicLivenet = [4]byte{0xe8, 0xf3, 0xe1, 0xe3} // 0xe3e1f3e8
	networkMagicTestnet = [4]byte{0xf4, 0xf3, 0xe5, 0xf4} // 0xf4e5f3f4
	networkMagicRegtest = [4]byte{0xfa, 0xbf, 0xb5, 0xda} // 0xdab5bffa
	networkMagicSTN     = [4]byte{0xf9, 0xc4, 0xce, 0xfb} // 0xfbcec4f9

	// DNS 种子
	dnsSeedsMainnet = []string{
		"seed.tbcdev.org",
		"seed.bitcoinunlimited.info",
	}
	dnsSeedsSTN = []string{
		"stn-seed.tbcdev.io",
	}

	// 预定义网络
	Livenet = &Network{
		Name:         "livenet",
		Alias:        "mainnet",
		PubKeyHash:   0x00,
		PrivateKey:   0x80,
		ScriptHash:   0x05,
		XPubKey:      0x0488b21e,
		XPrivKey:     0x0488ade4,
		NetworkMagic: networkMagicLivenet,
		Port:         8333,
		DNSServers:   dnsSeedsMainnet,
		CashAddrPref: "bitcoincash",
	}

	Testnet = &Network{
		Name:         "testnet",
		Alias:        "testnet",
		PubKeyHash:   0x6f,
		PrivateKey:   0xef,
		ScriptHash:   0xc4,
		XPubKey:      0x043587cf,
		XPrivKey:     0x04358394,
		NetworkMagic: networkMagicTestnet,
		Port:         18333,
		DNSServers:   dnsSeedsMainnet,
		CashAddrPref: "bchtest",
	}

	Regtest = &Network{
		Name:         "regtest",
		Alias:        "regtest",
		PubKeyHash:   0x6f,
		PrivateKey:   0xef,
		ScriptHash:   0xc4,
		XPubKey:      0x043587cf,
		XPrivKey:     0x04358394,
		NetworkMagic: networkMagicRegtest,
		Port:         18444,
		DNSServers:   []string{},
		CashAddrPref: "bchreg",
	}

	STN = &Network{
		Name:         "stn",
		Alias:        "stn",
		PubKeyHash:   0x6f,
		PrivateKey:   0xef,
		ScriptHash:   0xc4,
		XPubKey:      0x043587cf,
		XPrivKey:     0x04358394,
		NetworkMagic: networkMagicSTN,
		Port:         9333,
		DNSServers:   dnsSeedsSTN,
		CashAddrPref: "tbcstn",
	}

	// 网络映射表
	networkMap     = make(map[string]*Network)
	networkByMagic = make(map[uint32]*Network)
)

func init() {
	// 初始化网络映射
	AddNetwork(Livenet)
	AddNetwork(Testnet)
	AddNetwork(Regtest)
	AddNetwork(STN)
}

// AddNetwork 添加自定义网络（与 JS 的 Networks.add 对齐）
func AddNetwork(n *Network) error {
	if n == nil {
		return fmt.Errorf("network: nil network")
	}
	if n.Name == "" {
		return fmt.Errorf("network: empty name")
	}

	networkMap[n.Name] = n
	networkMap[n.Alias] = n

	// 按魔数索引
	magic := binary.LittleEndian.Uint32(n.NetworkMagic[:])
	networkByMagic[magic] = n

	return nil
}

// GetNetwork 根据名称或魔数获取网络（与 JS 的 Networks.get 对齐）
func GetNetwork(nameOrMagic interface{}) (*Network, bool) {
	switch v := nameOrMagic.(type) {
	case string:
		net, ok := networkMap[v]
		return net, ok
	case uint32:
		net, ok := networkByMagic[v]
		return net, ok
	case [4]byte:
		magic := binary.LittleEndian.Uint32(v[:])
		net, ok := networkByMagic[magic]
		return net, ok
	default:
		return nil, false
	}
}

// RemoveNetwork 移除网络（与 JS 的 Networks.remove 对齐）
func RemoveNetwork(n *Network) error {
	if n == nil {
		return fmt.Errorf("network: nil network")
	}

	delete(networkMap, n.Name)
	delete(networkMap, n.Alias)

	magic := binary.LittleEndian.Uint32(n.NetworkMagic[:])
	delete(networkByMagic, magic)

	return nil
}

// IsMainnet 判断网络是否为主网
func (n *Network) IsMainnet() bool {
	return n == Livenet || n.Name == "livenet" || n.Alias == "mainnet"
}

// String 返回网络名称
func (n *Network) String() string {
	if n == nil {
		return ""
	}
	return n.Name
}

// NetworkMagicUint32 返回网络魔数（uint32 格式）
func (n *Network) NetworkMagicUint32() uint32 {
	if n == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(n.NetworkMagic[:])
}
