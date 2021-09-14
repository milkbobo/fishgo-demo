package web

import (
	"errors"
	"flag"
	"os"
	"path"
	"reflect"

	"github.com/BurntSushi/toml"
)

type AppConfigBase struct {
	Appname    string `toml:"appname"`
	HttpPort   int    `toml:"httpport"`
	RunMode    string `toml:"runmode"`
	Accesslogs bool   `toml:"accesslogs"`
}

type AppConfigInfo struct {
	SecurityIpWhite string `toml:"securityipwhite"`
	Log             struct {
		Driver          string `toml:"driver"`
		File            string `toml:"file"`
		Maxline         int    `toml:"maxline"`
		Maxsize         int    `toml:"maxsize"`
		Daily           bool   `toml:"daily"`
		Maxday          int    `toml:"maxday"`
		Rotate          bool   `toml:"rotate"`
		Level           string `toml:"level"`
		Queuedriver     string `toml:"queuedriver"`
		Queuepoolsize   int    `toml:"queuepoolsize"`
		Cachedriver     string `toml:"cachedriver"`
		Cachesaveprefix string `toml:"cachesaveprefix"`
		PrettyPrint     bool   `toml:"prettyPrint"`
		Async           bool   `toml:"async"`
	} `toml:"log"`
	Grace struct {
		Driver string `toml:"driver"`
		Stop   string `toml:"stop"`
		Start  string `toml:"start"`
	} `toml:"grace"`
	Session struct {
		Driver          string `toml:"driver"`
		CookieName      string `toml:"cookieName"`
		EnableSetCookie bool   `toml:"enableSetCookie,omitempty"`
		GcLifeTime      int    `toml:"gclifetime"`
		Secure          bool   `toml:"secure"`
		CookieLifeTime  int    `toml:"cookieLifeTime"`
		ProviderConfig  string `toml:"providerConfig"`
		Domain          string `toml:"domain"`
		SessionIdLength int    `toml:"sessionIdLength"`
	} `toml:"session"`
	Mdb     AppConfigInfoMongoDB `toml:"mdb"`
	Mdb2    AppConfigInfoMongoDB `toml:"mdb2"`
	Mdb3    AppConfigInfoMongoDB `toml:"mdb3"`
	Mdb4    AppConfigInfoMongoDB `toml:"mdb4"`
	Mdb5    AppConfigInfoMongoDB `toml:"mdb5"`
	Esdb    AppConfigInfoEsDB    `toml:"Esdb"`
	Esdb2   AppConfigInfoEsDB    `toml:"Esdb2"`
	Esdb3   AppConfigInfoEsDB    `toml:"Esdb3"`
	DB      AppConfigInfoDB      `toml:"db"`
	DB2     AppConfigInfoDB      `toml:"db2"`
	DB3     AppConfigInfoDB      `toml:"db3"`
	DB4     AppConfigInfoDB      `toml:"db4"`
	DB5     AppConfigInfoDB      `toml:"db5"`
	Monitor struct {
		Driver        string `toml:"driver"`
		AppId         string `toml:"appId"`
		ErrorCount    string `toml:"errorCount"`
		CriticalCount string `toml:"criticalCount"`
	} `toml:"monitor"`
	Queue struct {
		Driver     string `toml:"driver"`
		SavePath   string `toml:"savepath"`
		SavePrefix string `toml:"saveprefix"`
		PoolSize   int    `toml:"poolsize"`
		Debug      bool   `toml:"debug"`
	} `toml:"queue"`
	Cache struct {
		Driver     string `toml:"driver"`
		SavePrefix string `toml:"saveprefix"`
		SavePath   string `toml:"savePath"`
		GcInterval int    `toml:"gcInterval"`
	} `toml:"cache"`
}

type AppConfigInfoMongoDB struct {
	Port     int    `toml:"port"`
	Host     string `toml:"host"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}

type AppConfigInfoEsDB struct {
	Port     int    `toml:"port"`
	Host     string `toml:"host"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type AppConfigInfoDB struct {
	Driver            string `toml:"driver"`
	Host              string `toml:"host"`
	Port              int    `toml:"port"`
	User              string `toml:"user"`
	Password          string `toml:"password"`
	Charset           string `toml:"charset"`
	Collation         string `toml:"collation"`
	Database          string `toml:"database"`
	Debug             bool   `toml:"debug"`
	MaxConnection     int    `toml:"maxConnection"`
	MaxIdleConnection int    `toml:"maxIdleConnection"`
}

type CheckAppConfig struct {
	AppConfigBase
	Prod AppConfigInfo `toml:"prod"`
	Dev  AppConfigInfo `toml:"dev"`
	Test AppConfigInfo `toml:"test"`
}

type AppConfig struct {
	AppConfigBase
	AppConfigInfo
}

type Configure interface {
	Get() AppConfig
}

type configureImplement struct {
	runMode  string
	configer AppConfig
}

func checkFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	} else {
		return true
	}
}

func findAppConfPath(file string) (string, bool, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", false, err
	}
	appPath := workingDir + "/" + file
	if checkFileExist(appPath) {
		return appPath, true, nil
	}

	for workingDir != "/" {
		workingDir = path.Dir(workingDir)
		appPath := workingDir + "/" + file
		if checkFileExist(appPath) {
			return appPath, false, nil
		}
	}
	return "", false, errors.New("can't not find conf")
}

func getAppConfigInfo(runmode string, checkAppConfigData CheckAppConfig) AppConfigInfo {
	if runmode == "prod" {
		return checkAppConfigData.Prod
	} else if runmode == "dev" {
		return checkAppConfigData.Dev
	} else if runmode == "test" {
		return checkAppConfigData.Test
	}
	return AppConfigInfo{}
}

func NewConfig(file string) (Configure, error) {
	appConfigPath, isCurrentDir, err := findAppConfPath(file)
	if err != nil {
		return nil, err
	}

	checkAppConfigData := CheckAppConfig{}
	if _, err := toml.DecodeFile(appConfigPath, &checkAppConfigData); err != nil {
		panic(err)
	}

	ConfigData := AppConfig{}
	ConfigData.AppConfigBase = checkAppConfigData.AppConfigBase

	var runMode string
	if isCurrentDir == false {
		runMode = "test"
		ConfigData.AppConfigInfo = checkAppConfigData.Test
	} else if flagRunMode := flag.String("runmode", "", "是什么环境运行环境"); *flagRunMode != "" {
		runMode = *flagRunMode
		ConfigData.AppConfigInfo = getAppConfigInfo(runMode, checkAppConfigData)
	} else if envRunMode := os.Getenv("BEEGO_RUNMODE"); envRunMode != "" {
		runMode = envRunMode
		ConfigData.AppConfigInfo = getAppConfigInfo(runMode, checkAppConfigData)
	} else if configRunMode := ConfigData.RunMode; configRunMode != "" {
		runMode = configRunMode
		ConfigData.AppConfigInfo = getAppConfigInfo(runMode, checkAppConfigData)
	} else {
		runMode = "dev"
		ConfigData.AppConfigInfo = checkAppConfigData.Dev
	}

	if reflect.DeepEqual(ConfigData, AppConfig{}) {
		panic("runMode不能为空")
	}

	return &configureImplement{
		runMode:  runMode,
		configer: ConfigData,
	}, nil
}

func (this *configureImplement) Get() AppConfig {
	return this.configer
}
