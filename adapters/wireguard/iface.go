package wireguard

import (
	"fmt"
	"net"
)

type Interface struct {
	name   string
	config []byte
}

func NewInterface(name string) *Interface {
	return &Interface{name: name}
}

func (i *Interface) Create(config []byte) error {
	i.config = config
	return nil
}

func (i *Interface) Delete() error {
	return nil
}

func (i *Interface) Up() error {
	return nil
}

func (i *Interface) Down() error {
	return nil
}

func (i *Interface) Address() (net.IP, error) {
	return nil, fmt.Errorf("not implemented")
}
