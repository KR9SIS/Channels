package main

import (
	"log"

	"github.com/pion/webrtc/v4"
)

type WebRTCService struct{}

func (m *WebRTCService) SetupPeer() error {
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	log.Print(peerConnectionConfig)
	return nil
}
