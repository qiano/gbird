package config

import (
	"encoding/json"
	"flag"
	// "fmt"
	"io/ioutil"
	"os"
	// "path"
	"path/filepath"
)

// 项目根目录
var ROOT string
var DebugEnv bool
var Config map[string]string
var Option map[string]string
var TradeOption map[string]string

var SimAccount map[string]string

func init() {
	binDir, err := ExecutableDir()
	if err != nil {
		return
	}
	// ROOT = path.Dir(binDir)
	ROOT = binDir
	return
}

func init() {
	LoadConfig("/config.json")
	var pDebugEnv *bool
	pDebugEnv = flag.Bool("d", false, "enable debug for dev")
	flag.Parse()
	DebugEnv = *pDebugEnv
}

func get_config_path(file string) (filepath string) {
	return ROOT + file
}

func load_config(file string) (config map[string]string, err error) {
	// Load 全局配置文件
	configFile := get_config_path(file)
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	config = make(map[string]string)
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func SaveContent(file string, object interface{}) error {
	content, err := json.Marshal(&object)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(ROOT+file, content, 666)
}

func LoadFile(file string) (content []byte, err error) {
	file = get_config_path(file)
	content, err = ioutil.ReadFile(file)
	return
}

func LoadContent(file string, object interface{}) error {
	content, err := LoadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, &object)
}

func save_config(file string, config map[string]string) (err error) {
	return SaveContent(file, config)
}

//LoadConfig 加载配置文件
func LoadConfig(filepath string) error {
	_Config, err := load_config(filepath)
	// _Config, err := load_config("/conf/config.json")
	if err != nil {
		return err
	}
	Config = make(map[string]string)
	Config = _Config
	return nil
}

// filesExists returns whether or not the named file or directory exists.
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// 获得可执行程序所在目录
func ExecutableDir() (string, error) {
	pathAbs, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "", err
	}
	return filepath.Dir(pathAbs), nil
}
