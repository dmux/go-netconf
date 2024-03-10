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
	"strings"
)

type NetworkConfig struct {
	Version  int                `yaml:"version"`
	Networks map[string]Network `yaml:"network"`
}

type Network struct {
	Version   int                 `yaml:"version"`
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

func readYAML(filePath string) ([]byte, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func showNetworkConfig(app *tview.Application) {
	filePath := "/etc/netplan/01-netcfg.yaml"
	configData, err := readYAML(filePath)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	configText := string(configData)
	modal := tview.NewModal().
		SetText(configText).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "OK" {
				app.Stop()
			}
		})
	if err := app.SetRoot(modal, false).SetFocus(modal).Run(); err != nil {
		panic(err)
	}
}

func main() {
	runConfigurationUI()
}

func trimStrings(slice []string) []string {
	trimmed := make([]string, len(slice))
	for i, s := range slice {
		trimmed[i] = strings.TrimSpace(s)
	}
	return trimmed
}

func runConfigurationUI() {
	app := tview.NewApplication()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		return event
	})

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
	form.AddInputField("IPv4 Address (cidr format)", ipv4Address, 0, nil, func(text string) {
		ipv4Address = text
	})

	gateway := ""
	form.AddInputField("Gateway", gateway, 0, nil, func(text string) {
		gateway = text
	})

	dnsServers := ""
	form.AddInputField("DNS Servers (comma separated)", dnsServers, 0, nil, func(text string) {
		dnsServers = text
	})

	form.AddButton("OK", func() {
		selectedOptionIndex, _ := interfaceField.GetCurrentOption()
		interfaceName := interfaceNames[selectedOptionIndex]
		config := NetworkConfig{
			Networks: map[string]Network{
				"network": {
					Version: 2,
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
					Addresses: trimStrings(strings.Split(dnsServers, ",")),
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

		runConfigurationUI()
	}).
		AddButton("Cancel", func() {
			app.Stop()
		})

	form.AddButton("Show Network Config", func() {
		showNetworkConfig(app)
	})

	form.AddFormItem(interfaceField)

	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}
}
