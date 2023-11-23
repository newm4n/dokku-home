package configuration

import (
	_ "embed"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"strings"
)

var (
	defCfg      map[string]string
	initialized = false
)

// initialize this configuration
func initialize() {
	viper.SetEnvPrefix("service")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	defCfg = make(map[string]string)

	defCfg["server.host"] = "0.0.0.0"
	defCfg["server.port"] = "8080"
	defCfg["server.log.level"] = "warn" // valid values are trace, debug, info, warn, error, fatal

	defCfg["server.timeout.write"] = "10 seconds"
	defCfg["server.timeout.read"] = "15 seconds"
	defCfg["server.timeout.idle"] = "60 seconds"

	defCfg["server.timeout.graceshut"] = "15 seconds"
	defCfg["config.path"] = "/path/to/config.json"

	for k := range defCfg {
		err := viper.BindEnv(k)
		if err != nil {
			log.Errorf("Failed to bind env \"%s\" into configuration. Got %s", k, err)
		}
	}

	initialized = true
}

// SetConfig put configuration key value
func SetConfig(key, value string) {
	viper.Set(key, value)
}

// Get fetch configuration as string value
func Get(key string) string {
	if !initialized {
		initialize()
	}
	ret := viper.GetString(key)
	if len(ret) == 0 {
		if ret, ok := defCfg[key]; ok {
			return ret
		}
		log.Debugf("%s config key not found", key)
	}
	return ret
}

// GetBoolean fetch configuration as boolean value
func GetBoolean(key string) bool {
	if len(Get(key)) == 0 {
		return false
	}
	b, err := strconv.ParseBool(Get(key))
	if err != nil {
		panic(err)
	}
	return b
}

// GetInt fetch configuration as integer value
func GetInt(key string) int {
	if len(Get(key)) == 0 {
		return 0
	}
	i, err := strconv.ParseInt(Get(key), 10, 64)
	if err != nil {
		panic(err)
	}
	return int(i)
}

// GetFloat fetch configuration as float value
func GetFloat(key string) float64 {
	if len(Get(key)) == 0 {
		return 0
	}
	f, err := strconv.ParseFloat(Get(key), 64)
	if err != nil {
		panic(err)
	}
	return f
}

// Set configuration key value
func Set(key, value string) {
	defCfg[key] = value
}

var (
	backendCache *ConfigurationFile
)

func GetEndPoint() []*BackendEnd {
	return loadEndPoint().EncPoints
}

var (
	//go:embed configpath/EndPoint.json
	EndPointStr []byte
)

func loadEndPoint() *ConfigurationFile {
	configPth := Get("config.path")
	fInfo, err := os.Stat(configPth)
	if err != nil || fInfo.IsDir() {
		log.Warnf("loading from default config")
		defaultEp := &ConfigurationFile{}
		err := json.Unmarshal(EndPointStr, defaultEp)
		if err != nil {
			log.Errorf("got error %s while loading defaylt config", err)
			return &ConfigurationFile{}
		}
		backendCache = defaultEp
		return defaultEp
	}
	fileBytes, err := os.ReadFile(configPth)
	if err != nil {
		log.Warnf("loading from default config because of %s", err.Error())
		defaultEp := &ConfigurationFile{}
		err := json.Unmarshal(EndPointStr, defaultEp)
		if err != nil {
			log.Errorf("got error %s while loading defaylt config", err)
			return &ConfigurationFile{}
		}
		backendCache = defaultEp
		return defaultEp
	}
	log.Infof("loading config from %s", configPth)
	fileEP := &ConfigurationFile{}
	err = json.Unmarshal(fileBytes, fileEP)
	if err != nil {
		log.Errorf("got error %s while loading config from %s", err, configPth)
		return &ConfigurationFile{}
	}
	if fileEP.Version != "1.0" {
		log.Errorf("configuration %s is not version 1.0", configPth)
		return &ConfigurationFile{}
	}
	backendCache = fileEP
	return fileEP
}
