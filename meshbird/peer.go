package meshbird

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/matthewgao/meshbird/config"
	"github.com/matthewgao/meshbird/iface"
	"github.com/matthewgao/meshbird/protocol"

	"github.com/golang/protobuf/proto"
	"github.com/matthewgao/meshbird/transport"
	"github.com/matthewgao/meshbird/utils"
)

type Peer struct {
	remoteDC   string
	remoteAddr string
	config     config.Config
	client     *transport.Client
}

func NewPeer(remoteDC, remoteAddr string, cfg config.Config, handler transport.ServerHandler, getRoutes func() []Route) *Peer {
	peer := &Peer{
		remoteDC:   remoteDC,
		remoteAddr: remoteAddr,
		config:     cfg,
	}
	// if remoteDC == cfg.Dc {
	// 	peer.client = transport.NewClient(remoteAddr, "", cfg.TransportThreads)
	// } else {
	peer.client = transport.NewClient(remoteAddr, cfg.Key, cfg.TransportThreads)
	// }
	peer.client.SetHandler(handler)
	return peer
}

func (p *Peer) Start() {
	p.client.Start()
	go p.process()
}

func (p *Peer) process() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("peer process panic: %s", err)
		}
	}()
	tickerPing := time.NewTicker(time.Second)
	defer tickerPing.Stop()
	for range tickerPing.C {
		p.SendPing()
	}
}

func (p *Peer) GetRealLocalAddrWithPort() string {
	return fmt.Sprintf("%s:%s", p.config.Ip, p.GetRealLocalAddrWithPort())
}

func (p *Peer) SendPing() {
	ip, _, err := net.ParseCIDR(p.config.Ip)
	utils.POE(err)
	env := &protocol.Envelope{
		Type: &protocol.Envelope_Ping{
			Ping: &protocol.MessagePing{
				Timestamp:        time.Now().UnixNano(),
				LocalAddr:        p.GetRealLocalAddrWithPort(), //唯一的表示一个CLINET端
				LocalPrivateAddr: p.config.LocalPrivateAddr,
				DC:               "client",
				IP:               ip.String(),
			},
		},
	}
	data, err := proto.Marshal(env)
	utils.POE(err)
	p.client.Write(data)
}

func (p *Peer) SendPacket(pkt iface.PacketIP) {
	data, _ := proto.Marshal(&protocol.Envelope{
		Type: &protocol.Envelope_Packet{
			Packet: &protocol.MessagePacket{Payload: pkt},
		},
	})
	p.client.Write(data)
}
