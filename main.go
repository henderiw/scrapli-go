package main

import (
	"fmt"
	"os"

	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/platform"
)

func main() {
	host := os.Args[1]

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
		return
	}
	d, err := p.GetNetworkDriver()
	if err != nil {
		fmt.Printf("failed to fetch network driver from the platform; error: %+v\n", err)

		return
	}

	err = d.Open()
	if err != nil {
		fmt.Printf("failed to open driver; error: %+v\n", err)

		return
	}

	defer d.Close()

	r, err := d.SendCommand("show version")
	if err != nil {
		fmt.Printf("failed to send command; error: %+v\n", err)
		return
	}

	fmt.Printf(
		"sent command '%s', output received (SendCommand):\n %s\n\n\n",
		r.Input,
		r.Result,
	)

}
