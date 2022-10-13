package common

import (
	"encoding/json"
	. "github.com/milkbobo/fishgoweb/language"
	"strconv"
	"strings"
	"time"
)

type CommonFunc struct {
	BaseModel
}

var LocationTime = time.FixedZone("CST", 8*3600)

func (this *CommonFunc) GetPaging() CommonPage {
	current := this.Ctx.GetParam("Current")
	pageSize := this.Ctx.GetParam("PageSize")
	if current == "" {
		current = this.Ctx.GetParam("current")
		if current == "" {
			Throw(1, "Current不能为空")
		}
	}
	if pageSize == "" {
		pageSize = this.Ctx.GetParam("pageSize")
		if pageSize == "" {
			Throw(1, "PageSize不能为空")
		}
	}
	currentInt, err := strconv.Atoi(current)
	if err != nil {
		if strings.Index(err.Error(), "invalid syntax") > -1 {
			Throw(1, "PageIndex必须为数字")
		}
		panic(err)
	}
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		if strings.Index(err.Error(), "invalid syntax") > -1 {
			Throw(1, "PageSize必须为数字")
		}
		panic(err)
	}

	if pageSizeInt > 20 {
		Throw(1, "PageSize最多拿20个！！！")
	}

	pageIndex := (currentInt - 1) * pageSizeInt

	return CommonPage{
		PageSize:  pageSizeInt,
		PageIndex: pageIndex,
	}
}

func (this *CommonFunc) PostToStruct(data interface{}) {

	if this.Ctx.GetMethod() != "POST" {
		Throw(1, "请求Method不是POST方法: "+this.Ctx.GetMethod())
	}
	body, err := this.Ctx.GetBody()
	if err != nil {
		panic(err)
	}

	if len(body) == 0 {
		return
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
}

func (this *CommonFunc) HeightToTime(height int64) int64 {
	return 1598306400 + (height * 30)
}

func (this *CommonFunc) TimeToHeight(timestamp int64) int64 {
	return (timestamp - 1598306400) / 30
}

func (this *CommonFunc) TimestampToTimeString(timestamp int64) string {
	return time.Unix(timestamp, 0).In(LocationTime).Format("2006-01-02 15:04:05")
}

func init() {
}
