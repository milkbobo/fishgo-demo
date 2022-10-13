package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	. "github.com/milkbobo/fishgoweb/language"
)

var (
	didInherit = os.Getenv("LISTEN_FDS") != ""
	ppid       = os.Getppid()
)

type Grace interface {
	ListenAndServe(port int, handler http.Handler) error
}

type GraceConfig struct {
	Driver  string
	Stop    []string
	Restart []string
}

type graceImplement struct {
	runGrace      bool
	stopSignal    map[os.Signal]bool
	restartSignal map[os.Signal]bool
}

func NewGrace(config GraceConfig) (Grace, error) {
	var goGrace bool
	if config.Driver != "" {
		goGrace = true
	} else {
		goGrace = false
	}
	stopSignal := map[os.Signal]bool{}
	for _, singleStop := range config.Stop {
		signal, err := stringToSignal(singleStop)
		if err != nil {
			return nil, err
		}
		stopSignal[signal] = true
	}
	restartSignal := map[os.Signal]bool{}
	for _, singleRestart := range config.Restart {
		signal, err := stringToSignal(singleRestart)
		if err != nil {
			return nil, err
		}
		restartSignal[signal] = true
	}
	return &graceImplement{
		runGrace:      goGrace,
		stopSignal:    stopSignal,
		restartSignal: restartSignal,
	}, nil
}

func NewGraceFromConfig() (Grace, error) {
	gracedirver := globalBasic.Config.Get().Grace.Driver
	gracestopStr := globalBasic.Config.Get().Grace.Stop
	gracerestartStr := globalBasic.Config.Get().Grace.Start

	gracestop := Explode(gracestopStr, ",")
	gracerestart := Explode(gracerestartStr, ",")
	config := GraceConfig{
		Driver:  gracedirver,
		Stop:    gracestop,
		Restart: gracerestart,
	}
	return NewGrace(config)
}

func stringToSignal(signalStr string) (os.Signal, error) {
	result := map[string]os.Signal{
		"TERM": syscall.SIGTERM,
		"INT":  syscall.SIGINT,
		"HUP":  syscall.SIGHUP,
		//"USR1": syscall.SIGUSR1,
		//"USR2": syscall.SIGUSR2,
	}
	target, isExist := result[signalStr]
	if isExist == false {
		return nil, fmt.Errorf("invalid signal %v", signalStr)
	} else {
		return target, nil
	}
}

func (this *graceImplement) listenAndServeGrace(httpPort string, handler http.Handler) error {
	server := &http.Server{
		Addr:    httpPort,
		Handler: handler,
	}

	//等待服务器结束
	errorEvent := make(chan error)
	waitEvent := make(chan bool)

	go func() {
		defer close(waitEvent)

		// 当前的 Goroutine 等待信号量
		quit := make(chan os.Signal)
		// 监控信号
		signalArray := []os.Signal{}
		for v, _ := range this.stopSignal {
			signalArray = append(signalArray, v)
		}
		signal.Notify(quit, signalArray...)
		// 这里会阻塞当前 Goroutine 等待信号
		<-quit
		//调用Server.Shutdown graceful结束
		globalBasic.Log.Debug("%+v", "开始优雅关闭")
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := server.Shutdown(timeoutCtx)
		if err != nil {
			fmt.Errorf("Server Shutdown: %+v", err)
		}
	}()

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Errorf("ListenAndServe: %+v", err)
		}
	}()

	select {
	case err := <-errorEvent:
		if err == nil {
			panic("unexpected nil error")
		}
		return err
	case <-waitEvent:
		globalBasic.Log.Debug("Exiting pid %v.", os.Getpid())
		return nil
	}
}

func (this *graceImplement) ListenAndServe(httpPort int, handler http.Handler) error {
	if this.runGrace == false {
		return http.ListenAndServe(":"+strconv.Itoa(httpPort), handler)
	} else {
		return this.listenAndServeGrace(":"+strconv.Itoa(httpPort), handler)
	}
}
