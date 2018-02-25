package gbird

import (
	"gbird/base"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

//Test 测试
type Test struct {
	ID            bson.ObjectId `bson:"_id"  collection:"scene" urlname:"scene" `
	SceneNo       string        `display:"场景编号" required:"true"`
	Name          string        `display:"场景名称" required:"true"`
	Group         string        `display:"场景分组"`
	BankNo        string        `display:"银行卡号"`
	EncTrueName   string        `display:"收款方用户名"`
	BankCode      string        `display:"收款方开户行"`
	Owner         bson.ObjectId `bson:",omitempty"` //场景拥有人
	RoyaltyMethod string        `display:"提成方式" enum:"P:百分比,O:按单"`
	RoyaltyValue  float64
	base.Base
}

func TestRegister(t *testing.T) {

}
