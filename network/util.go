package network

import (
	"fmt"
	"net"
)

// Get finds the first available position (index) in the string that's set to 0
// Returns:
// - int: Index of first available position, exclude: first, second, last
// - error: nil if found, error if no positions available
// eg: s := "0000000"
func GetChar(s *string, c byte) (uint, error) {
	for i, v := range *s {
		if i == 0 || i == len(*s)-1 {
			continue
		}
		if v == rune(c) {
			return uint(i), nil
		}
	}
	return 0, fmt.Errorf("no valid positions")
}

// eg: s := "0000000"
func SetChar(n uint, s *string, c byte) error {
	if n >= uint(len(*s)) {
		return fmt.Errorf("bitStr.Set: index %d out of range [0-%d]", n, len(*s)-1)
	}
	bb := []byte(*s)
	if bb[n] == c {
		return nil
	}
	bb[n] = c
	*s = string(bb)

	return nil
}

func Uint2IPv4(n uint) net.IP {
	ip := []byte{0, 0, 0, 0}
	ip[0] = byte(n >> 24)
	if ip[0] > 0 {
		n = n - uint(ip[0])<<24
	}
	ip[1] = byte(n >> 16)
	if ip[1] > 0 {
		n = n - uint(ip[1])<<16
	}
	ip[2] = byte(n >> 8)
	if ip[2] > 0 {
		n = n - uint(ip[2])<<8
	}
	ip[3] = byte(n)
	return ip
}

func IPv42Uint(bs net.IP) uint {
	return uint(bs[0])<<24 + uint(bs[1])<<16 + uint(bs[2])<<8 + uint(bs[3])
}

func ParseFirstIP(subnet string) (string, error) {
	errFormat := "parseFirstIP: %w"
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", fmt.Errorf(errFormat, err)
	}
	ipnet.IP[3] |= 1

	return ipnet.String(), nil
}
