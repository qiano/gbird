package upload

import (
	"fmt"
	"gbird/util"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"time"
)

//SaveFile 保存上传的文件
func SaveFile(file *multipart.FileHeader, flag string) (path string, err error) {
	basedir := util.GetCurrDir()
	savedir := "/assets/" + flag + "/"
	if !util.IsExistFileOrDir(basedir + savedir) {
		os.MkdirAll(basedir+savedir, 0777) //创建文件夹
	}
	tnow := fmt.Sprintf("%d", time.Now().UnixNano())
	path = basedir + savedir + tnow + "_" + flag + "_" + file.Filename
	src, err := file.Open()
	if err != nil {
		return
	}
	defer src.Close()

	out, err := os.Create(path)
	if err != nil {
		return
	}
	defer out.Close()
	io.Copy(out, src)
	return
}

//SaveFileDirect 直接保存文件在assets中
func SaveFileDirect(file *multipart.FileHeader, path, filename string) (string, error) {
	basedir := util.GetCurrDir()
	savedir :=  path + "/"
	// savedir := "/assets/" + path + "/"
	if !util.IsExistFileOrDir(basedir + savedir) {
		os.MkdirAll(basedir+savedir, 0777) //创建文件夹
	}
	path = basedir + savedir + filename + "."+strings.Split(file.Filename, ".")[1]
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()
	io.Copy(out, src)
	return path, nil
}
