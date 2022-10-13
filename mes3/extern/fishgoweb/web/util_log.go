package web

import (
	"encoding/json"
	"fmt"
	"github.com/beego/beego/logs"
	"github.com/k0kubun/pp"
	. "github.com/milkbobo/fishgoweb/language"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

type Log interface {
	WithContextAndMonitor(ctx Context, monitor Monitor) Log
	Emergency(format string, v ...interface{})
	Alert(format string, v ...interface{})
	Critical(format string, v ...interface{})
	Error(format string, v ...interface{})
	Warning(format string, v ...interface{})
	Notice(format string, v ...interface{})
	Informational(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Close()
}

type LogConfig struct {
	Driver      string `json:driver`
	Filename    string `json:"filename"`
	Maxlines    int    `json:"maxlines"`
	Maxsize     int    `json:"maxsize"`
	Daily       bool   `json:"daily"`
	Maxdays     int    `json:"maxdays"`
	Rotate      bool   `json:"rotate"`
	Level       int    `json:"level"`
	PrettyPrint bool   `json:"prettyprint"`
	Async       bool   `json:"async"`
}

type logImplement struct {
	*logs.BeeLogger
	monitor     Monitor
	logPrefix   string
	prettyPrint bool
}

func getLevel(in string) int {
	levelString := map[string]int{
		"Emergency":     logs.LevelEmergency,
		"Alert":         logs.LevelAlert,
		"Critical":      logs.LevelCritical,
		"Error":         logs.LevelError,
		"Warning":       logs.LevelWarning,
		"Notice":        logs.LevelNotice,
		"Informational": logs.LevelInformational,
		"Debug":         logs.LevelDebug,
	}
	for key, value := range levelString {
		if strings.ToLower(in) == strings.ToLower(key) {
			return value
		}
	}
	return logs.LevelDebug
}

func NewLog(config LogConfig) (Log, error) {
	if config.Driver == "" {
		config.Driver = "console"
	}
	if config.Level == 0 {
		config.Level = logs.LevelDebug
	}
	logConfigString, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	Log := logs.NewLogger(10000)
	err = Log.SetLogger(config.Driver, string(logConfigString))
	if err != nil {
		return nil, err
	}
	if config.Async {
		Log = Log.Async()
	}

	return &logImplement{
		BeeLogger:   Log,
		prettyPrint: config.PrettyPrint,
	}, nil
}

func NewLogFromConfig() (Log, error) {
	logConfig := LogConfig{}
	logConfig.Driver = globalBasic.Config.Get().Log.Driver
	logConfig.Filename = globalBasic.Config.Get().Log.File
	logConfig.Maxlines = globalBasic.Config.Get().Log.Maxline
	logConfig.Maxsize = globalBasic.Config.Get().Log.Maxsize
	logConfig.Daily = globalBasic.Config.Get().Log.Daily
	logConfig.Maxdays = globalBasic.Config.Get().Log.Maxday
	logConfig.Rotate = globalBasic.Config.Get().Log.Rotate
	logConfig.Level = getLevel(globalBasic.Config.Get().Log.Level)
	logConfig.PrettyPrint = globalBasic.Config.Get().Log.PrettyPrint
	logConfig.Async = globalBasic.Config.Get().Log.Async

	return NewLog(logConfig)
}

func (this *logImplement) WithContextAndMonitor(ctx Context, monitor Monitor) Log {
	logPrefix := ctx.GetRemoteAddr()
	newLogManager := *this
	newLogManager.logPrefix = logPrefix
	newLogManager.monitor = monitor
	return &newLogManager
}

func (this *logImplement) getTraceLineNumber(traceNumber int) string {
	_, filename, line, ok := runtime.Caller(traceNumber + 1)
	if !ok {
		return "???.go:???"
	} else {
		return path.Base(filename) + ":" + strconv.Itoa(line)
	}
}

func (this *logImplement) getLogFormat(format string, v []interface{}) string {
	if this.prettyPrint {
		format = strings.Replace(format, "%+v", "%v", -1)
		format = strings.Replace(format, "%#v", "%v", -1)
		for singleIndex, singleV := range v {
			singleVType := reflect.TypeOf(singleV)
			singleVTypeKind := GetTypeKind(singleVType)
			if singleVTypeKind == TypeKind.BOOL ||
				singleVTypeKind == TypeKind.INT ||
				singleVTypeKind == TypeKind.UINT ||
				singleVTypeKind == TypeKind.FLOAT ||
				singleVTypeKind == TypeKind.STRING {
				v[singleIndex] = singleV
			} else {
				v[singleIndex] = pp.Sprint(singleV)
			}
		}
	}
	v = append([]interface{}{this.logPrefix, this.getTraceLineNumber(2)}, v...)
	return fmt.Sprintf("%s %s "+format, v...)
}

func (this *logImplement) Emergency(format string, v ...interface{}) {
	this.BeeLogger.Emergency("%s", this.getLogFormat(format, v))
}

func (this *logImplement) Alert(format string, v ...interface{}) {
	this.BeeLogger.Alert("%s", this.getLogFormat(format, v))
}

func (this *logImplement) Critical(format string, v ...interface{}) {
	if this.monitor != nil {
		this.monitor.AscCriticalCount()
	}
	this.BeeLogger.Critical("%s", this.getLogFormat(format, v))
}

func (this *logImplement) Error(format string, v ...interface{}) {
	if this.monitor != nil {
		this.monitor.AscErrorCount()
	}
	this.BeeLogger.Error("%s", this.getLogFormat(format, v))
}

func (this *logImplement) Warning(format string, v ...interface{}) {
	this.BeeLogger.Warning("%s", this.getLogFormat(format, v))
}

func (this *logImplement) Notice(format string, v ...interface{}) {
	this.BeeLogger.Notice("%s", this.getLogFormat(format, v))
}

func (this *logImplement) Informational(format string, v ...interface{}) {
	this.BeeLogger.Informational("%s", this.getLogFormat(format, v))
}

func (this *logImplement) Debug(format string, v ...interface{}) {
	this.BeeLogger.Debug("%s", this.getLogFormat(format, v))
}

func (this *logImplement) Warn(format string, v ...interface{}) {
	this.BeeLogger.Warn("%s", this.getLogFormat(format, v))
}

func (this *logImplement) Info(format string, v ...interface{}) {
	this.BeeLogger.Info("%s", this.getLogFormat(format, v))
}

func (this *logImplement) Trace(format string, v ...interface{}) {
	this.BeeLogger.Trace(this.getLogFormat(format, v))
}

func (this *logImplement) Close() {
	this.BeeLogger.Close()
}
