package network

import (
	"fmt"
	"net"
	"os"

	"github.com/hsheth2/water/waterutil"
)

const DEFAULT_MTU = 1500

var MTU int

func Init() {
	MTU = 0
}

type Interface struct {
	name string
	file *os.File
}

func (i Interface) Name() string {
	return i.name
}

func (i *Interface) Write(data []byte) (n int, err error) {
	return i.file.Write(data)
}

func (i *Interface) Read(data []byte) (n int, err error) {
	return i.file.Read(data)
}

func CreateTunInterface(ifceName string) (*Interface, error) {
	ifce, err := interfaceOpen("tun", ifceName)
	if err != nil {
		return nil, fmt.Errorf("create new tun interface %s err: %s", ifce, err)
	}
	err = UpInterface(ifce.Name())
	if err != nil {
		return nil, fmt.Errorf("tun interface %s up err: %s", ifce.Name(), err)
	}
	return ifce, nil
}

func CreateTunInterfaceWithIp(iface string, IpAddr string) (*Interface, error) {
	ifce, err := CreateTunInterface(iface)
	if err != nil {
		return nil, err
	}
	err = AssignIpAddress(ifce.Name(), IpAddr)
	return ifce, err
}

func NextNetworkPacket(iface *Interface) ([]byte, error) {
	if MTU == 0 {
		MTU = DEFAULT_MTU
	}

	raw_data := make([]byte, MTU)

	_, err := iface.Read(raw_data)
	return raw_data, err
}

func IPv4Destination(packet []byte) net.IP {
	return waterutil.IPv4Destination(packet)

}
