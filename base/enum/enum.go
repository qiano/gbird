package enum

import (
	"errors"
	"gbird/util/logger"
)

//EnumHub 枚举、状态集合
var enumHub map[string][]Enum

func init() {
	enumHub = make(map[string][]Enum)
}

//Enum 订单状态
type Enum struct {
	Code string      //代码
	Desc string      //描述
	Data interface{} //数据
}

//Register 注册枚举
func Register(catagory string, datas []Enum) {
	_, ok := enumHub[catagory]
	if !ok {
		enumHub[catagory] = datas
	} else {
		panic(errors.New("Enum注册 " + catagory + " 已存在"))
	}
	logger.Infoln("枚举" + catagory + "注册成功")
}

//GetArray 获取枚举集合
func GetArray(catagory string) []Enum {
	val, ok := enumHub[catagory]
	if !ok {
		panic(errors.New("Enum：" + catagory + " 未找到"))
	}
	return val
}

//GetVal  获取枚举值
func GetVal(catagory, code string) Enum {
	enums := GetArray(catagory)
	for i := 0; i < len(enums); i++ {
		if code == enums[i].Code {
			return enums[i]
		}
	}
	panic(errors.New("Enum：" + catagory + " " + code + " 未找到"))
}

//GetValByDesc 通过描述找枚举值
func GetValByDesc(catagory, desc string) Enum {
	enums := GetArray(catagory)
	for i := 0; i < len(enums); i++ {
		if desc == enums[i].Desc {
			return enums[i]
		}
	}
	panic(errors.New("Enum：" + catagory + " " + desc + " 未找到"))
}
