package config

import (
	"time"
)

type Config struct {
	ConfigId   int `xorm:"autoincr"`
	Name       string
	Value      string
	CreateTime time.Time `xorm:"created"`
	ModifyTime time.Time `xorm:"updated"`
}

type Configs struct {
	Data  []Config
	Count int
}
