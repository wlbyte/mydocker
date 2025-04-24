package network

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/vishvananda/netlink"
	"github.com/wlbyte/mydocker/consts"
)

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

func setInterfaceUP(name string) error {
	errFormat := "bridge.setInterfaceUP: %w"
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
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

func (b *BridgeNetworkDriver) DelConnect(network *Network, endpoint *Endpoint) error {
	errFormat := "bridge.DelConnect: %w"
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
	if err := netlink.LinkDel(&endpoint.Device); err != nil {
		fmt.Println("[warn] bridge.DelConnect: ", err)
	}
	return nil
}

func ConfigBridge(driverStr, bridgeName, subnetStr string) error {
	errFormat := "configNetwork: %w"
	_, err := net.InterfaceByName(bridgeName)
	if err == nil {
		return nil
	}
	if !strings.Contains(err.Error(), "no such network interface") {
		return err
	}
	driver, err := NewNetworkDriver(driverStr)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if subnetStr == "" {
		return fmt.Errorf(errFormat, errors.New("subnet is required"))
	}
	ip, ipRange, _ := net.ParseCIDR(subnetStr)
	ipRange.IP = ip
	_, sub, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	ipam := NewIPAM()
	gateway, err := ipam.Allocate(sub)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	err = driver.Create(subnetStr, bridgeName)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}

	n := &Network{
		Name:    bridgeName,
		Subnet:  subnetStr,
		Driver:  driverStr,
		Gateway: gateway.String(),
	}
	if err := n.Dump(); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	return nil
}
