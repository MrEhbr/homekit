package main

import (
	"context"
	"net/http"

	"github.com/MrEhbr/homekit"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/log"
)

func main() {
	// create an accessory
	info := accessory.Info{Name: "Lamp"}
	ac := accessory.NewSwitch(info)

	t, err := homekit.NewTransport(12345, "config", "00102003", ac.Accessory)
	if err != nil {
		log.Info.Panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	hc.OnTermination(func() {
		cancel()
	})

	if err := t.Run(ctx); err != nil && err != http.ErrServerClosed {
		log.Info.Printf("transport Run error: %s", err)
	}
}
