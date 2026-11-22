package main

import (
	"log"

	"tailscale.com/tsnet"
)

type TailscaleService struct{}

func (m *TailscaleService) GetTailnetPeers() error {
	srv := new(tsnet.Server)
	if err := srv.Start(); err != nil {
		log.Fatalf("can't start tsnet server: %v", err)
		return err
	}
	defer srv.Close()
	return nil
}
