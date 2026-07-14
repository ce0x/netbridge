package wireguard

import "fmt"

type RouteManager struct{}

func NewRouteManager() *RouteManager {
	return &RouteManager{}
}

func (r *RouteManager) AddRoute(cidr string, iface string) error {
	fmt.Printf("ip route add %s dev %s\n", cidr, iface)
	return nil
}

func (r *RouteManager) RemoveRoute(cidr string) error {
	fmt.Printf("ip route del %s\n", cidr)
	return nil
}

func (r *RouteManager) SetDefaultGateway(gw string, iface string) error {
	fmt.Printf("ip route add default via %s dev %s\n", gw, iface)
	return nil
}
