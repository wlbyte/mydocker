package network

import (
	"net"
	"testing"

	"github.com/wlbyte/mydocker/consts"
)

func TestAllocate(t *testing.T) {
	ipam := &IPAM{
		SubnetAllocatorPath: consts.PATH_IPAM_JSON,
	}
	_, subnet, err := net.ParseCIDR("172.18.0.0/16")
	if err != nil {
		t.Errorf("net.ParseCIDR %s", err)
	}
	ip, err := ipam.Allocate(subnet)
	if err != nil {
		t.Errorf("ipam.Allocate %s", err)
	}
	t.Logf("%s", ip.String())
}

func TestRelease(t *testing.T) {
	ipam := &IPAM{
		SubnetAllocatorPath: consts.PATH_IPAM_JSON,
	}
	_, subnet, err := net.ParseCIDR("172.18.0.0/16")
	if err != nil {
		t.Errorf("net.ParseCIDR %s", err)
	}
	ip := net.ParseIP("172.18.0.2")
	ip4 := ip[12:]
	if err := ipam.Release(subnet, &ip4); err != nil {
		t.Errorf("ipam.Release %s", err)
	}
}

var bridgeName = "mydocker0"

func TestBridgeCreate(t *testing.T) {
	bridge := &BridgeNetworkDriver{}
	err := bridge.Create("172.18.0.0/16", bridgeName)
	if err != nil {
		t.Errorf("bridge.Create %s", err)
	}
	t.Logf("bridge.Create %s", bridgeName)
}

func TestBridgeDelete(t *testing.T) {
	bridge := &BridgeNetworkDriver{}
	if err := bridge.Delete(bridgeName); err != nil {
		t.Errorf("bridge.Delete %s", err)
	}
	t.Logf("bridge.Delete %s", bridgeName)
}

func TestBridgeConnect(t *testing.T) {
	ep := &Endpoint{
		ID: "abcdefghijkl",
	}
	n := &Network{
		Name: bridgeName,
	}
	bridge := &BridgeNetworkDriver{}
	err := bridge.Connect(n, ep)
	if err != nil {
		t.Errorf("bridge.Connect %s", err)
	}
	t.Logf("bridge.Connect %s", ep.ID)
}

func TestBridgeDisConnect(t *testing.T) {
	ep := &Endpoint{
		ID: "abcdefghijkl",
	}
	n := &Network{
		Name: bridgeName,
	}
	bridge := &BridgeNetworkDriver{}
	err := bridge.Disconnect(n, ep)
	if err != nil {
		t.Errorf("bridge.Connect %s", err)
	}
	t.Logf("bridge.Connect %s", ep.ID)
}
