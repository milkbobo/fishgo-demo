package test

import (
	. "github.com/milkbobo/fishgoweb/web"
	"github.com/milkbobo/fishgoweb/web/test/models/number"
)

type InnerTest struct {
	Test
	NumberAoTest number.NumberAoTest
}

func (this *InnerTest) TestBasic() {
	this.NumberAoTest.TestBasic()
}

func init() {
	InitTest(&InnerTest{})
}
