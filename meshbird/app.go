package meshbird

import (
	"log"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/matthewgao/meshbird/config"
	"github.com/matthewgao/meshbird/iface"
	"github.com/matthewgao/meshbird/protocol"
	"github.com/matthewgao/meshbird/transport"
)

type App struct {
	config config.Config
	// peers  map[string]*Peer
	client *Peer
	routes map[string]Route
	mutex  sync.RWMutex
	server *transport.Server
	iface  *iface.Iface
}

func NewApp(config config.Config) *App {
	return &App{
		config: config,
		// peers:  make(map[string]*Peer),
		routes: make(map[string]Route),
	}
}

func (a *App) Run() error {
	a.server = transport.NewServer(a.config.LocalAddr, a.config.LocalPrivateAddr, a, a.config.Key)
	a.server.SetConfig(a.config)

	a.routes[a.config.RemoteLocalAddrs] = Route{
		LocalAddr:        a.config.LocalAddr,
		LocalPrivateAddr: a.config.LocalPrivateAddr,
		IP:               a.config.Ip,
		DC:               a.config.Dc,
	}

	err := a.bootstrap()
	if err != nil {
		return err
	}
	go a.server.Start()
	return a.runIface()
}

func (a *App) runIface() error {
	a.iface = iface.New("", a.config.Ip, a.config.Mtu)
	err := a.iface.Start()
	if err != nil {
		return err
	}
	pkt := iface.NewPacketIP(a.config.Mtu)
	if a.config.Verbose == 1 {
		log.Printf("interface name: %s", a.iface.Name())
	}
	for {
		n, err := a.iface.Read(pkt)
		if err != nil {
			return err
		}
		src := pkt.GetSourceIP().String()
		dst := pkt.GetDestinationIP().String()
		if a.config.Verbose == 1 {
			log.Printf("tun packet: src=%s dst=%s len=%d", src, dst, n)
		}
		// a.mutex.RLock()
		// peer, ok := a.peers[a.routes[dst].LocalAddr]
		if a.config.ServerMode == 1 {
			log.Printf("receiver tun packet dst address is %s", dst)
			conn, ok := a.server.Conns[a.routes[dst].LocalAddr]
			if !ok {
				if a.config.Verbose == 1 {
					log.Printf("unknown destination, packet dropped")
				}
			} else {
				conn.SendPacket(pkt)
			}
		} else {
			//client send packet
			a.client.SendPacket(pkt)
		}
		// a.mutex.RUnlock()

	}
}

func (a *App) bootstrap() error {
	if a.config.ServerMode == 1 {
		log.Printf("running in server mode, skip make connection to client")
		return nil
	}
	// seedAddrs := strings.Split(a.config.SeedAddrs, ",")
	// for _, seedAddr := range seedAddrs {
	// 	parts := strings.Split(seedAddr, "/")
	// 	if len(parts) < 2 {
	// 		continue
	// 	}
	// 	seedDC := parts[0]
	// 	seedAddr = parts[1]
	// 	if seedAddr == a.config.LocalAddr {
	// 		log.Printf("skip seed addr %s because it's local addr", seedAddr)
	// 		continue
	// 	}

	//For server no need to make connection to client -gs
	// if a.config.ServerMode == 0 {
	peer := NewPeer("server", a.config.RemoteAddrs, a.config, a.getRoutes)
	peer.Start()

	// a.mutex.Lock()
	// a.peers[seedAddr] = peer
	// a.mutex.Unlock()
	// }
	// }
	return nil
}

func (a *App) getRoutes() []Route {
	a.mutex.Lock()
	routes := make([]Route, len(a.routes))
	i := 0
	for _, route := range a.routes {
		routes[i] = route
		i++
	}
	a.mutex.Unlock()
	return routes
}

func (a *App) OnData(buf []byte) {
	ep := protocol.Envelope{}
	err := proto.Unmarshal(buf, &ep)
	if err != nil {
		log.Printf("proto unmarshal err: %s", err)
		return
	}
	switch ep.Type.(type) {
	case *protocol.Envelope_Ping:
		ping := ep.GetPing()
		//log.Printf("received ping: %s", ping.String())
		a.mutex.Lock()
		a.routes[ping.GetIP()] = Route{
			LocalAddr:        ping.GetLocalAddr(),
			LocalPrivateAddr: ping.GetLocalPrivateAddr(),
			IP:               ping.GetIP(),
			DC:               ping.GetDC(),
		}
		// if _, ok := a.peers[ping.GetLocalAddr()]; !ok {
		// 	var peer *Peer
		// 	if a.config.Dc == ping.GetDC() {
		// 		peer = NewPeer(ping.GetDC(), ping.GetLocalPrivateAddr(),
		// 			a.config, a.getRoutes)
		// 	} else {
		// 		peer = NewPeer(ping.GetDC(), ping.GetLocalAddr(),
		// 			a.config, a.getRoutes)
		// 	}
		// 	peer.Start()
		// 	a.peers[ping.GetLocalAddr()] = peer
		// 	if a.config.Verbose == 1 {
		// 		log.Printf("new peer %s", ping)
		// 	}
		// }
		if a.config.Verbose == 1 {
			log.Printf("routes %s", a.routes)
		}
		a.mutex.Unlock()
	case *protocol.Envelope_Packet:
		pkt := iface.PacketIP(ep.GetPacket().GetPayload())
		if a.config.Verbose == 1 {
			log.Printf("received packet: src=%s dst=%s len=%d",
				pkt.GetSourceIP(), pkt.GetDestinationIP(), len(pkt))
		}
		a.iface.Write(pkt)
	}
}
