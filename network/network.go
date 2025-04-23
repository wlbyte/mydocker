package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"github.com/wlbyte/mydocker/consts"
	"github.com/wlbyte/mydocker/container"
)

/*
1. docker run 运行容器
2. 如果没有执行网络， 则默认使用 bridge 网络
3. 检查默认bridge有没有创建，没有则创建
4. 创建容器网络端点， 并把本地端点加入到 bridge
5. 把 peer加入到容器
6. up容器 lo 口， 配置容器接口地址和路由

*/

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

func Connect(c *container.Container) error {
	errFormat := "network.Connect: %w"
	// 初始化默认bridge网络
	if c.Network == "host" {
		return fmt.Errorf(errFormat, errors.New("unsupport host"))
	} else if c.Network == "mydocker0" || c.Network == "" {
		if err := ConfigBridge("", c.Network, "172.18.0.0/24"); err != nil {
			return fmt.Errorf(errFormat, err)
		}
	} else {
		if _, err := net.InterfaceByName(c.Network); err != nil {
			return fmt.Errorf(errFormat, err)
		}
	}
	// create veth
	bridge := &BridgeNetworkDriver{}

	n := Network{
		Name: c.Network,
	}
	if err := n.Load(); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	// 获取 IP
	_, sub, err := net.ParseCIDR(n.Subnet)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	ip, err := NewIPAM().Allocate(sub)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}

	endpoint := &Endpoint{
		ID:          c.Id + "-" + n.Name,
		IPAddress:   ip,
		Network:     &n,
		PortMapping: c.PortMapping,
	}

	if err := bridge.Connect(&n, endpoint); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	if err := configEndpointIpAddressAndRoute(endpoint, c); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	if err := configPortMapping(endpoint, c); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	return nil
}

func Disconnect(c *container.Container) error {
	return nil
}

func configEndpointIpAddressAndRoute(e *Endpoint, c *container.Container) error {
	errFormat := "configEndpointIpAddressAndRoute: %w"
	l, err := netlink.LinkByName(e.Device.PeerName)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	defer enterContainerNetns(&l, c)()

	_, ipNet, err := net.ParseCIDR(e.Network.Subnet)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	ipNet.IP = e.IPAddress
	if err := setInterfaceIP(e.Device.PeerName, string(ipNet.String())); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	if err := setInterfaceUP(e.Device.PeerName); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := setInterfaceUP("lo"); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	_, subnet, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	defaultRoute := netlink.Route{
		LinkIndex: l.Attrs().Index,
		Dst:       subnet,
		Gw:        net.ParseIP(e.Network.Gateway),
	}
	if err := netlink.RouteAdd(&defaultRoute); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	return nil
}

func configPortMapping(ep *Endpoint, cinfo *container.Container) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			logrus.Errorf("port mapping format error, %v", pm)
			continue
		}
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		//err := cmd.Run()
		output, err := cmd.Output()
		if err != nil {
			logrus.Errorf("iptables Output, %v", output)
			continue
		}
	}
	return nil
}

func enterContainerNetns(enLink *netlink.Link, cinfo *container.Container) func() {
	errFormat := "enterContainer: %s"
	f, err := os.OpenFile(fmt.Sprintf("/proc/%d/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		fmt.Printf(errFormat, err)
		f.Close()
		return nil
	}

	nsFD := f.Fd()
	runtime.LockOSThread()
	origns, err := netns.Get()
	if err != nil {
		fmt.Printf(errFormat, err)
		runtime.UnlockOSThread()
		origns.Close()
		return nil
	}

	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		fmt.Printf(errFormat, err)
		runtime.UnlockOSThread()
		origns.Close()
		return nil
	}
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		fmt.Printf(errFormat, err)
		runtime.UnlockOSThread()
		origns.Close()
		return nil
	}

	return func() {
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}
