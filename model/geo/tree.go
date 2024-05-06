package geo

import (
	"net"
)

type node struct {
	left, right *node
	city        *City
}

type Tree struct {
	root *node
}

func NewTree() *Tree {
	return &Tree{
		root: &node{},
	}
}

func (t *Tree) Insert(cidr *net.IPNet, city *City) {
	count := 0
	maskOnes, _ := cidr.Mask.Size()

	root := t.root
	for _, b := range cidr.IP {
		for i := 7; i >= 0; i-- {
			if (b>>i)&1 > 0 {
				if root.right == nil {
					root.right = &node{}
				}
				root = root.right
			} else {
				if root.left == nil {
					root.left = &node{}
				}
				root = root.left
			}
			count++
			if count == maskOnes {
				root.city = city
				return
			}
		}
	}
}

func (t *Tree) Find(ip net.IP) *City {
	root := t.root
	ip = ip.To4()
	var res *City

	for _, b := range ip {
		for i := 7; i >= 0; i-- {
			if (b>>i)&1 > 0 {
				if root.right == nil {
					return res
				}
				root = root.right
			} else {
				if root.left == nil {
					return res
				}
				root = root.left
			}
			if root.city != nil {
				res = root.city
			}
		}
	}

	return res
}
