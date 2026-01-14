package main

import (
	"fmt"
	"net"
	"os"
	"text/template"

	vdpa "github.com/k8snetworkplumbingwg/govdpa/pkg/kvdpa"
	cli "github.com/urfave/cli/v2"
)

const deviceTemplate = ` - Name: {{ .Name }}
   Management Device: {{ .MgmtDev.Name }}
   Driver: {{ .Driver }}
{{- if eq .Driver "virtio_vdpa" }}
   Virtio Net Device:
      Name: {{ .VirtioNet.Name }}
      NetDev: {{ .VirtioNet.NetDev }}
{{ else if eq .Driver "vhost_vdpa" }}
   Vhost Vdpa Device:
      Name: {{ .VhostVdpa.Name }}
      Path: {{ .VhostVdpa.Path }}
{{ end }}`

func listAction(c *cli.Context) error {
	var devs []vdpa.VdpaDevice
	var err error
	var mgmtDev = c.String("mgmtdev")
	if mgmtDev != "" {
		var busName, devName string
		busName, devName, err = vdpa.ExtractBusAndMgmtDevice(mgmtDev)
		if err != nil {
			return err
		}
		devs, err = vdpa.GetVdpaDevicesByMgmtDev(busName, devName)
	} else {
		devs, err = vdpa.ListVdpaDevices()
	}
	if err != nil {
		return err
	}
	tmpl := template.Must(template.New("device").Parse(deviceTemplate))

	for _, dev := range devs {
		if err := tmpl.Execute(os.Stdout, dev); err != nil {
			panic(err)
		}
	}
	return nil
}

func getAction(c *cli.Context) error {
	tmpl := template.Must(template.New("device").Parse(deviceTemplate))
	for i := 0; i < c.Args().Len(); i++ {
		name := c.Args().Get(i)
		dev, err := vdpa.GetVdpaDevice(name)
		if err != nil {
			return err
		}
		if err := tmpl.Execute(os.Stdout, dev); err != nil {
			panic(err)
		}
	}
	return nil
}

func addAction(c *cli.Context) error {
	if c.Args().Len() != 2 {
		err := cli.ShowAppHelp(c)
		return err
	}

	mgmtDevName := c.Args().Get(0)
	devName := c.Args().Get(1)

	return vdpa.AddVdpaDevice(mgmtDevName, devName)
}

func deleteAction(c *cli.Context) error {
	if c.Args().Len() != 1 {
		err := cli.ShowAppHelp(c)
		return err
	}

	devName := c.Args().Get(0)

	return vdpa.DeleteVdpaDevice(devName)
}

func setMacAction(c *cli.Context) error {
	if c.Args().Len() != 2 {
		err := cli.ShowAppHelp(c)
		return err
	}

	devName := c.Args().Get(0)
	macAddr, err := net.ParseMAC(c.Args().Get(1))

	if err != nil {
		return err
	}

	return vdpa.SetVdpaDeviceMac(devName, macAddr)
}

func main() {
	app := &cli.App{
		Name:  "kvdpa-cli",
		Usage: "Interact with Kernel vDPA devices",
		Commands: []*cli.Command{
			{Name: "list",
				Usage:  "List vdpa devices",
				Action: listAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "mgmtdev",
						Usage: "Name of the management device: [busName/]devName",
					},
				},
			},
			{Name: "get",
				Usage:     "Get a specific vdpa device",
				Action:    getAction,
				ArgsUsage: "[name]",
			},
			{Name: "add",
				Usage:     "Add a new vdpa device",
				Action:    addAction,
				ArgsUsage: "[mgmtdev] [dev]",
			},
			{Name: "del",
				Usage:     "Delete a vdpa device",
				Action:    deleteAction,
				ArgsUsage: "[dev]",
			},
			{Name: "setmac",
				Usage:     "Set mac to an existing vdpa device",
				Action:    setMacAction,
				ArgsUsage: "[dev] [mac]",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
