// Package classification User API.
//
// The purpose of this service is to provide an application
// that is using plain go code to define an API
//
//      Host: localhost
//      Version: 0.0.1
//
// swagger:meta
package types

import (
	"net"
	"strconv"
)

const (
	DefaultAPIPort = 9500
)

func GetAPIServerAddressFromIP(ip string) string {
	return net.JoinHostPort(ip, strconv.Itoa(DefaultAPIPort))
}
