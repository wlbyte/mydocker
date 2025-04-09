package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/vishvananda/netlink"
	"github.com/wlbyte/mydocker/consts"
)

var (
	drivers = map[string]Driver{}
)

type Network struct {
	Name    string `json:"name"`
	Subnet  string `json:"subnet"`
	Gateway string `json:"gateway"`
	Driver  string `json:"driver"`
}

func (n *Network) Dump() error {
	errFormat := "network.Dump: %w"
	bs, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	filePath := filepath.Join(consts.PATH_NETWORK_NETWORK, n.Name+".json")
	if err := os.WriteFile(filePath, bs, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}
func (n *Network) Load() error {
	errFormat := "network.Load: %w"
	filePath := filepath.Join(consts.PATH_NETWORK_NETWORK, n.Name+".json")
	bs, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := json.Unmarshal(bs, n); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

type Endpoint struct {
	ID          string `json:"id"`
	Device      netlink.Veth
	IPAddress   net.IP
	MacAddress  net.HardwareAddr
	Network     *Network
	PortMapping []string
}

type IPAMer interface {
	Allocate(subnet *net.IPNet) (ip net.IP, err error)
	Release(subnet *net.IPNet, ipaddr *net.IP) error
	ReleaseSubnet(subnet string) error
}

// IPAM 实现了 IPAMer 接口
type IPAM struct {
	wg                  sync.Mutex
	SubnetAllocatorPath string
	Subnets             map[string]*string
}

var ipAllocator = &IPAM{
	wg:                  sync.Mutex{},
	SubnetAllocatorPath: consts.PATH_IPAM_JSON,
}

func NewIPAM() IPAMer {
	return ipAllocator
}

func (i *IPAM) load() error {
	errFormat := "ipam.Load: %w"
	bs, err := os.ReadFile(i.SubnetAllocatorPath)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}

	if err := json.Unmarshal(bs, &i.Subnets); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	return nil
}

func (i *IPAM) dump() error {
	errFormat := "ipam.Dump: %w"
	bs, err := json.Marshal(i.Subnets)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return os.WriteFile(consts.PATH_IPAM_JSON, bs, consts.MODE_0755)
}

func (i *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	errFormat := "ipam.Allocate: %w"
	i.wg.Lock()
	defer i.wg.Unlock()
	i.Subnets = map[string]*string{}
	i.load()
	ones, total := subnet.Mask.Size()
	if _, exist := i.Subnets[subnet.String()]; !exist {
		newS := strings.Repeat("0", 1<<(total-ones))
		i.Subnets[subnet.String()] = &newS
	}
	n, err := GetChar(i.Subnets[subnet.String()], '0')
	if err != nil {
		return nil, fmt.Errorf(errFormat, err)
	}
	if err := SetChar(n, i.Subnets[subnet.String()], '1'); err != nil {
		return nil, fmt.Errorf(errFormat, err)
	}
	ip = net.IP{0, 0, 0, 0}
	ipTmp := Uint2IPv4(uint(n))
	ip[0] = subnet.IP[0] | ipTmp[0]
	ip[1] = subnet.IP[1] | ipTmp[1]
	ip[2] = subnet.IP[2] | ipTmp[2]
	ip[3] = subnet.IP[3] | ipTmp[3]
	if err := i.dump(); err != nil {
		return nil, fmt.Errorf(errFormat, err)
	}
	return ip, nil
}

func (i *IPAM) Release(subnet *net.IPNet, ip *net.IP) error {
	errFormat := "ipam.Release: %w"
	i.wg.Lock()
	defer i.wg.Unlock()
	subN := IPv42Uint(subnet.IP)
	ipN := IPv42Uint(*ip)
	n := ipN - subN
	if i.Subnets == nil {
		i.Subnets = map[string]*string{}
		i.load()
	}
	if err := SetChar(n, i.Subnets[subnet.String()], '0'); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := i.dump(); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func (i *IPAM) ReleaseSubnet(subnet string) error {
	errFormat := "ipam.ReleaseSubnet: %w"
	i.wg.Lock()
	defer i.wg.Unlock()
	if err := i.load(); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	delete(i.Subnets, subnet)
	if err := i.dump(); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

type Driver interface {
	Name() string
	Create(subnet string, name string) error
	Delete(name string) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network *Network, endpoint *Endpoint) error
}

// BridgeNetworkDriver 实现 Driver 接口
type BridgeNetworkDriver struct {
}

func newDefaultNetworkDriver() Driver {
	return &BridgeNetworkDriver{}
}

func NewNetworkDriver(driver string) (Driver, error) {
	errFormat := "newNetworkDriver: %w"
	var d Driver
	if driver == "" || driver == consts.DEFAULT_DRIVER {
		d = newDefaultNetworkDriver()
	} else {
		return nil, fmt.Errorf(errFormat, errors.New("unsupported driver"))
	}
	return d, nil
}

func (b *BridgeNetworkDriver) Name() string {
	return consts.DEFAULT_DRIVER
}
func (b *BridgeNetworkDriver) Create(subnet string, name string) error {
	errFormat := "bridge.Create: %w"
	_, sub, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := b.initBridge(subnet, name); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := setupIPTables(name, sub, "add"); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func (b *BridgeNetworkDriver) initBridge(subnet, name string) error {
	errFormat := "bridge.initBridge: %w"
	if err := createBridgeInterface(name); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	address, err := ParseFirstIP(subnet)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := setInterfaceIP(name, address); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	return nil
}

func createBridgeInterface(name string) error {
	errFormat := "bridge.createBridgeInterface: %w"
	_, err := net.InterfaceByName(name)
	if err == nil {
		return fmt.Errorf(errFormat, errors.New("bridge already exists"))
	}
	if !strings.Contains(err.Error(), "no such network interface") {
		return err
	}
	la := netlink.NewLinkAttrs()
	la.Name = name
	br := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := netlink.LinkSetUp(br); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func setInterfaceIP(name string, address string) error {
	errFormat := "bridge.setInterfaceIP: %w"
	addr, err := netlink.ParseAddr(address)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := netlink.AddrReplace(link, addr); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func setupIPTables(bridgeName string, subnet *net.IPNet, action string) error {
	errFormat := "bridge.setupIPTables: %w"
	var act string
	if action == "del" {
		act = "-D"
	} else {
		act = "-A"
	}
	cmdStr := fmt.Sprintf("-t nat %s POSTROUTING -s %s ! -o %s -j MASQUERADE", act, subnet.String(), bridgeName)
	if output, err := exec.Command("iptables", strings.Split(cmdStr, " ")...).CombinedOutput(); err != nil {
		return fmt.Errorf(errFormat, errors.New(string(output)))
	}
	return nil
}
func (b *BridgeNetworkDriver) Delete(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("network.Delete: %w", err)
	}
	return nil
}
func (b *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	errFormat := "bridge.Connect: %w"
	br, err := netlink.LinkByName(network.Name)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	la.MasterIndex = br.Attrs().Index
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}
	if err := netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}
func (b *BridgeNetworkDriver) Disconnect(network *Network, endpoint *Endpoint) error {
	errFormat := "bridge.Disconnect: %w"
	vethName := endpoint.ID[:5]
	veth, err := netlink.LinkByName(vethName)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := netlink.LinkSetNoMaster(veth); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	return nil
}
