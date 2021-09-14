package common

import (
	"encoding/json"
	. "github.com/milkbobo/fishgoweb/language"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"strings"
	"time"
)

type CommonFunc struct {
	BaseModel
}

var LocationTime = time.FixedZone("CST", 8*3600)

func (this *CommonFunc) BsonMarshal(data interface{}) ([]byte, error) {
	return bson.Marshal(data)
}

func (this *CommonFunc) BsonUnmarshal(bsonByteData []byte) (bson.D, error) {
	result := bson.D{}
	err := bson.Unmarshal(bsonByteData, &result)

	return result, err
}

func (this *CommonFunc) StructToBsonD(data interface{}) (bson.D, error) {

	result := bson.D{}
	insertByte, err := bson.Marshal(data)
	if err != nil {
		return result, err
	}

	err = bson.Unmarshal(insertByte, &result)
	return result, err
}

func (this *CommonFunc) GetPaging() *options.FindOptions {
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
	currentInt, err := strconv.ParseInt(current, 10, 64)
	if err != nil {
		if strings.Index(err.Error(), "invalid syntax") > -1 {
			Throw(1, "PageIndex必须为数字")
		}
		panic(err)
	}
	pageSizeInt, err := strconv.ParseInt(pageSize, 10, 64)
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

	opts := options.Find()
	//opts.Limit = &pageSizeInt
	//opts.Skip = &pageIndex
	opts.SetLimit(pageSizeInt)
	opts.SetSkip(pageIndex)
	return opts
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
	return (timestamp - 1598306400)/30
}

func (this *CommonFunc) TimestampToTimeString(timestamp int64) string {
	return time.Unix(timestamp, 0).In(LocationTime).Format("2006-01-02 15:04:05")
}

func init() {
}
