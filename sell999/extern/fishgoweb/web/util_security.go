package web

import (
	"errors"
	. "github.com/milkbobo/fishgoweb/language"
	. "github.com/milkbobo/fishgoweb/util"
	"runtime"
)

type Security interface {
}

type SecurityConfig struct {
	IpWhite []string
}

func NewSecurity(config SecurityConfig) (Security, error) {
	var netConfig string
	if len(config.IpWhite) == 0 {
		return nil, nil
	}
	if runtime.GOOS == "darwin" {
		netConfig = "en0"
	} else {
		netConfig = "eth0"
	}
	ip, err := NewIfconfig().GetIP(netConfig)
	if err != nil {
		return nil, err
	}

	ipStr := ip.IP.String()
	if len(config.IpWhite) != 0 && ArrayIn(config.IpWhite, ipStr) == -1 {
		return nil, errors.New("当前IP: " + ipStr + "不在IP白名单中: " + Implode(config.IpWhite, ","))
	}

	return nil, nil
}

func NewSecurityFromConfig() (Security, error) {
	ipwhite := globalBasic.Config.Get().SecurityIpWhite
	ipwhiteList := Explode(ipwhite, ",")
	return NewSecurity(SecurityConfig{IpWhite: ipwhiteList})
}
