package blacklist

import (
	"errors"
	"net"
	"sync"
)

var (
	ErrSubnetExists    = errors.New("subnet is in the list")
	ErrSubnetNotExists = errors.New("subnet is not in the list")
)

type Subnets struct {
	subnets map[string]*net.IPNet
	mtx     sync.RWMutex
}

func NewSubnets() *Subnets {
	return &Subnets{
		subnets: map[string]*net.IPNet{},
	}
}

func (s *Subnets) Add(cidr *net.IPNet) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, exists := s.subnets[cidr.String()]; exists {
		return ErrSubnetExists
	}

	s.subnets[cidr.String()] = cidr
	return nil
}

func (s *Subnets) Delete(cidr *net.IPNet) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, exists := s.subnets[cidr.String()]; !exists {
		return ErrSubnetNotExists
	}

	delete(s.subnets, cidr.String())

	return nil
}

func (s *Subnets) CheckIp(ip net.IP) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	for _, cidr := range s.subnets {
		if cidr.Contains(ip) {
			return true
		}
	}

	return false
}
