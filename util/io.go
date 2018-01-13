package util

import (
	// "fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//GetCurrDir 获取当前文件执行的路径
func GetCurrDir() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	path = strings.Replace(path, "\\", "/", -1)
	splitstring := strings.Split(path, "/")
	size := len(splitstring)
	return strings.Join(splitstring[0:size-1], "/")
}

//IsExistFileOrDir 文件夹或文件是否存在
func IsExistFileOrDir(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

//ExecutableDir 获得可执行程序所在目录
func ExecutableDir() (string, error) {
	pathAbs, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "", err
	}
	return filepath.Dir(pathAbs), nil
}

//CreateDir 创建目录
func CreateDir(path string, relative bool) error {
	paths := strings.Split(path, "/")
	dir := ""
	if relative {
		dir, _ = os.Getwd() //当前的目录
	}
	for i := 0; i <= len(paths); i++ {
		curPath := dir + "/" + strings.Join(paths[0:i], "/")
		if !IsExistFileOrDir(curPath) {
			err := os.Mkdir(curPath, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetFileContentAsStringLines(filePath string) ([]string, error) {
	result := []string{}
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return result, err
	}
	s := string(b)
	for _, lineStr := range strings.Split(s, "\n") {
		lineStr = strings.TrimSpace(lineStr)
		if lineStr == "" {
			continue
		}
		result = append(result, lineStr)
	}
	return result, nil
}

//SaveFile 保存文件
func SaveFile(savedir, filename string, data []byte) error {
	basedir := GetCurrDir()
	savedir = basedir + savedir
	if !IsExistFileOrDir(savedir) {
		os.MkdirAll(savedir, 0777) //创建文件夹
	}
	ioutil.WriteFile(savedir+filename, data, 0666)
	return nil
}
