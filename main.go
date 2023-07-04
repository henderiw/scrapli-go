package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scrapli/scrapligo/driver/opoptions"
	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/platform"
)

const (
	keyStartMarker  = "-----BEGIN RSA PRIVATE KEY-----"
	keyEndMarker    = "-----END RSA PRIVATE KEY-----"
	certStartMarker = "-----BEGIN CERTIFICATE-----"
	certEndMarker   = "-----END CERTIFICATE-----"
	caStartMarker   = "-----BEGIN CERTIFICATE-----"
	caEndMarker     = "-----END CERTIFICATE-----"
)

func main() {
	host := os.Args[1]

	certData, err := getCertificateData()
	if err != nil {
		panic(err)
	}

	if err := sendConfig(host, certData); err != nil {
		panic(err)
	}
}

func sendConfig(host string, certData *certData) error {

	p, err := platform.NewPlatform(
		// cisco_iosxe refers to the included cisco iosxe platform definition
		"nokia_srl",
		host,
		options.WithAuthNoStrictKey(),
		options.WithAuthUsername("admin"),
		options.WithAuthPassword("NokiaSrl1!"),
	)
	if err != nil {
		fmt.Printf("failed to create platform; error: %+v\n", err)
		return err
	}
	d, err := p.GetNetworkDriver()
	if err != nil {
		fmt.Printf("failed to fetch network driver from the platform; error: %+v\n", err)

		return err
	}
	err = d.Open()
	if err != nil {
		fmt.Printf("failed to open driver; error: %+v\n", err)

		return err
	}
	defer d.Close()

	configs := []string{
		fmt.Sprintf("set / system tls server-profile %s", certData.ProfileName),
		fmt.Sprintf("set / system tls server-profile %s authenticate-client false", certData.ProfileName),
	}

	r, err := d.SendConfigs(configs)
	if err != nil {
		return err
	}
	fmt.Println("response:", r.Failed)
	for _, resp := range r.Responses {
		fmt.Println("response:", resp)
	}

	commands := []string{
		"enter candidate",
		fmt.Sprintf("set / system tls server-profile %s", certData.ProfileName),
		fmt.Sprintf("set / system tls server-profile %s authenticate-client false", certData.ProfileName),
		fmt.Sprintf("set / system tls server-profile %s key %s", certData.ProfileName, certData.Key),
		fmt.Sprintf("set / system tls server-profile %s certificate \"%s\"", certData.ProfileName, certData.Cert),
		fmt.Sprintf("set / system tls server-profile %s trust-anchor \"%s\"", certData.ProfileName, certData.CA),
		"commit save",
	}

	//for _, cmd := range commands {
	fmt.Printf("cmd %s\n", commands)
	r, err = d.SendCommands(commands, opoptions.WithEager())
	if err != nil {
		return err
	}
	fmt.Printf("cmd input %s, response: %v\n", r.Failed, r.Responses)

	//}

	return nil
}

type certData struct {
	ProfileName string
	CA          string
	Cert        string
	Key         string
}

func getCertificateData() (*certData, error) {
	certData := &certData{
		ProfileName: "k8s-profile",
	}
	files, err := os.ReadDir("data")
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if !f.IsDir() {
			b, err := os.ReadFile(filepath.Join("data", f.Name()))
			if err != nil {
				return nil, err
			}
			var found bool
			if f.Name() == "ca.crt" {
				certData.CA, found = getStringInBetween(string(b), caStartMarker, caEndMarker, true)
				if !found {
					return nil, fmt.Errorf("cannot get the ca string")
				}
			}
			if f.Name() == "tls.crt" {
				certData.Cert, found = getStringInBetween(string(b), certStartMarker, certEndMarker, true)
				if !found {
					return nil, fmt.Errorf("cannot get the cert string")
				}
			}
			if f.Name() == "tls.key" {
				certData.Key, found = getStringInBetween(string(b), keyStartMarker, keyEndMarker, false)
				if !found {
					return nil, fmt.Errorf("cannot get the key string")
				}
				certData.Key = strings.ReplaceAll(certData.Key, "\n", "")
			}
		}
	}
	return certData, nil
}

// GetStringInBetween returns a string between the start/end markers with markers either included or excluded
func getStringInBetween(str string, start, end string, include bool) (result string, found bool) {
	// start index
	sidx := strings.Index(str, start)
	if sidx == -1 {
		return "", false
	}

	// forward start index if we don't want to include the markers
	if !include {
		sidx += len(start)
	}

	newS := str[sidx:]

	// end index
	eidx := strings.Index(newS, end)
	if eidx == -1 {
		return "", false
	}
	// to include the end marker, increment the end index up till its length
	if include {
		eidx += len(end)
	}

	return newS[:eidx], true
}
