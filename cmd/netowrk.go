package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/consts"
	"github.com/wlbyte/mydocker/network"
)

func init() {
	NetworkCommand.Subcommands = []cli.Command{
		NetworkCreateCommand,
		NetworkListCommand,
		NetworkRemoveCommand,
	}
}

var NetworkCommand = cli.Command{
	Name:  "network",
	Usage: "network management",
	Action: func(context *cli.Context) error {
		if len(context.Args()) == 0 {
			return fmt.Errorf("missing network subcommand")
		}
		return nil
	},
}

// docker network create -driver "bridge" --subnet "172.18.0.0/16" mydocker0
var NetworkCreateCommand = cli.Command{
	Name:  "create",
	Usage: "create a container network",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "driver",
			Usage: "network driver",
		},
		cli.StringFlag{
			Name:  "subnet",
			Usage: "subnet cidr",
		},
	},
	Action: func(context *cli.Context) error {
		errFormat := "network.Create: %w"
		if len(context.Args()) < 1 {
			return fmt.Errorf(errFormat, fmt.Errorf("missing network name"))
		}
		networkName := context.Args().Get(0)
		driverStr := context.String("driver")
		driver, err := network.NewNetworkDriver(driverStr)
		if err != nil {
			return fmt.Errorf(errFormat, err)
		}
		subnetStr := context.String("subnet")
		if subnetStr == "" {
			return fmt.Errorf(errFormat, errors.New("subnet is required"))
		}
		ip, ipRange, _ := net.ParseCIDR(subnetStr)
		ipRange.IP = ip
		_, sub, err := net.ParseCIDR(subnetStr)
		if err != nil {
			return fmt.Errorf(errFormat, err)
		}
		ipam := network.NewIPAM()
		gateway, err := ipam.Allocate(sub)
		if err != nil {
			return fmt.Errorf(errFormat, err)
		}
		err = driver.Create(subnetStr, networkName)
		if err != nil {
			return fmt.Errorf(errFormat, err)
		}

		n := &network.Network{
			Name:    networkName,
			Subnet:  subnetStr,
			Driver:  driverStr,
			Gateway: gateway.String(),
		}
		if err := n.Dump(); err != nil {
			return fmt.Errorf(errFormat, err)
		}
		return nil
	},
}

var NetworkListCommand = cli.Command{
	Name:  "list",
	Usage: "list container network",
	Action: func(context *cli.Context) error {
		jsonFiles := findJsonFilePathAll(consts.PATH_NETWORK_NETWORK)
		for _, f := range jsonFiles {
			bs, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			fmt.Println(string(bs))
		}
		return nil
	},
}

var NetworkRemoveCommand = cli.Command{
	Name:  "remove",
	Usage: "remove container network",

	Action: func(context *cli.Context) error {
		errFormat := "network.Remove: %w"
		if len(context.Args()) < 1 {
			return fmt.Errorf(errFormat, fmt.Errorf("missing network name"))
		}
		networkName := context.Args().Get(0)
		if networkName == "mydocker0" {
			return fmt.Errorf(errFormat, fmt.Errorf("couldn't remove default network"))
		}
		n := &network.Network{
			Name: networkName,
		}
		jsonFile := findJsonFilePath(n.Name, consts.PATH_NETWORK_NETWORK)
		var driver network.Driver
		driver, err := network.NewNetworkDriver(n.Driver)
		if err != nil {
			return fmt.Errorf(errFormat, err)
		}
		driver.Delete(n.Name)
		bs, err := os.ReadFile(jsonFile)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf(errFormat, err)
		}
		if err := json.Unmarshal(bs, n); err != nil {
			return fmt.Errorf(errFormat, err)
		}
		ipam := network.NewIPAM()
		if err := ipam.ReleaseSubnet(n.Subnet); err != nil {
			return fmt.Errorf(errFormat, err)
		}
		if err := os.Remove(jsonFile); err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf(errFormat, err)
		}
		return nil
	},
}
