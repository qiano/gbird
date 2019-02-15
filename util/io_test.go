package util

import (
	"fmt"
	"testing"
)

func TestCurrentPath(t *testing.T) {
	path := GetCurrDir()
	fmt.Println(path)
}

func TestCheckExist(t *testing.T) {
	fmt.Println("测试文件或文件夹是否存在")
	ret := IsExistFileOrDir("/users/qianshuai/go/src/qian/vitrodiag/uploadfile.html")
	if !ret {
		t.Fail()
	}
}
