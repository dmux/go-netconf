package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v2"
	"log"
	"net"
	"os"
	"os/exec"
)

type NetworkConfig struct {
	Networks map[string]Network `yaml:"network"`
}

type Network struct {
	Ethernets map[string]Ethernet `yaml:"ethernets,omitempty"`
}

type Ethernet struct {
	Dhcp4       bool   `yaml:"dhcp4,omitempty"`
	Dhcp6       bool   `yaml:"dhcp6,omitempty"`
	Addresses   string `yaml:"addresses,omitempty"`
	Gateway4    string `yaml:"gateway4,omitempty"`
	Nameservers Names  `yaml:"nameservers,omitempty"`
}

type Names struct {
	Addresses []string `yaml:"addresses,omitempty"`
}

func main() {
	app := tview.NewApplication()

	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Configure Networking").SetTitleAlign(tview.AlignLeft)
	form.SetFieldBackgroundColor(tcell.ColorDefault)

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	interfaceNames := make([]string, 0, len(interfaces))
	for _, iface := range interfaces {
		interfaceNames = append(interfaceNames, iface.Name)
	}

	interfaceField := tview.NewDropDown()
	interfaceField.SetLabel("Interface Name")
	interfaceField.SetOptions(interfaceNames, nil)

	dhcp := false
	form.AddCheckbox("DHCP", dhcp, nil)

	ipv4Address := ""
	form.AddInputField("IPv4 Address", ipv4Address, 0, nil, func(text string) {
		ipv4Address = text
	})

	gateway := ""
	form.AddInputField("Gateway", gateway, 0, nil, func(text string) {
		gateway = text
	})

	dnsServer := ""
	form.AddInputField("DNS Server", dnsServer, 0, nil, func(text string) {
		dnsServer = text
	})

	form.AddButton("OK", func() {
		selectedOptionIndex, _ := interfaceField.GetCurrentOption()
		interfaceName := interfaceNames[selectedOptionIndex]
		config := NetworkConfig{
			Networks: map[string]Network{
				"version": {
					Ethernets: map[string]Ethernet{
						interfaceName: {
							Dhcp4: dhcp,
						},
					},
				},
			},
		}

		if !dhcp {
			config.Networks["version"].Ethernets[interfaceName] = Ethernet{
				Addresses: ipv4Address,
				Gateway4:  gateway,
				Nameservers: Names{
					Addresses: []string{dnsServer},
				},
			}
		}

		yamlBytes, err := yaml.Marshal(config)
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile("/etc/netplan/01-netcfg.yaml", yamlBytes, 0644)
		if err != nil {
			log.Fatal(err)
		}

		cmd := exec.Command("netplan", "apply")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Error applying netplan: %s", out)
		}

		fmt.Println("Network configuration applied successfully.")

		app.Stop()
	}).
		AddButton("Cancel", func() {
			app.Stop()
		})

	form.AddFormItem(interfaceField)

	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}
}
