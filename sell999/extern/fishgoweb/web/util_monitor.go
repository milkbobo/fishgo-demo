package web

import (
	"errors"
	. "github.com/milkbobo/fishgoweb/sdk"
)

type Monitor interface {
	AscErrorCount()
	AscCriticalCount()
}

type MonitorConfig struct {
	Driver        string
	AppId         string
	ErrorCount    string
	CriticalCount string
}

type monitorImplement struct {
	AliCloudMonitorSdk
	config MonitorConfig
}

func NewMonitor(config MonitorConfig) (Monitor, error) {
	if config.Driver == "" {
		return nil, nil
	} else if config.Driver == "aliyuncloudmonitor" {
		result := &monitorImplement{
			AliCloudMonitorSdk: AliCloudMonitorSdk{
				AppId: config.AppId,
			},
			config: config,
		}
		go result.AliCloudMonitorSdk.Sync()
		return result, nil
	} else {
		return nil, errors.New("invalid monitor config " + config.Driver)
	}
}

func NewMonitorFromConfig() (Monitor, error) {
	monitorConfig := MonitorConfig{}
	monitorConfig.Driver = globalBasic.Config.Get().Monitor.Driver
	monitorConfig.AppId = globalBasic.Config.Get().Monitor.AppId
	monitorConfig.ErrorCount = globalBasic.Config.Get().Monitor.ErrorCount
	monitorConfig.CriticalCount = globalBasic.Config.Get().Monitor.CriticalCount
	return NewMonitor(monitorConfig)
}

func (this *monitorImplement) AscErrorCount() {
	if this.config.ErrorCount != "" {
		this.Asc(this.config.ErrorCount, 1)
	}
}

func (this *monitorImplement) AscCriticalCount() {
	if this.config.CriticalCount != "" {
		this.Asc(this.config.CriticalCount, 1)
	}
}
