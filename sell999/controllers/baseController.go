package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"reflect"
	"strconv"
	"time"

	. "github.com/milkbobo/fishgoweb/encoding"
	. "github.com/milkbobo/fishgoweb/language"
	. "github.com/milkbobo/fishgoweb/web"
)

type BaseController struct {
	Controller
}

type baseControllerResult struct {
	Code int
	Data interface{}
	Msg  string
}

func (this *BaseController) initCors() {
	//缓存设置
	this.Ctx.WriteHeader("Cache-Control", "private, no-store, no-cache, must-revalidate, max-age=0")
	this.Ctx.WriteHeader("Cache-Control", "post-check=0, pre-check=0")
	this.Ctx.WriteHeader("Pragma", "no-cache")

	//Access-Control-Allow-Origin设置为*只是用来解决post请求跨域的问题的，cookie是没法跨域的，或者说只能跨2级域名，每个cookie都会有对应的域，写的时候可以指定为允许同一个1级域名下的所有2级域名访问
	origin := this.Ctx.GetHeader("Origin")
	if origin == "" {
		origin = this.Ctx.GetSite()
	}

	this.Ctx.WriteHeader("Access-Control-Allow-Origin", origin)
	this.Ctx.WriteHeader("Access-Control-Allow-Credentials", "true")
	// this.Ctx.WriteHeader("Access-Control-Allow-Headers", "Content-Type")
}

/*
 * 文件信息
 */
type FileInfo struct {
	Data       []byte
	Name       string
	ModifyTime time.Time
	Etag       string
	Encoding   string
}

func (this *BaseController) fileRender(result baseControllerResult) {
	//判断result的FileInfo是否为空
	fileInfo, ok := result.Data.(FileInfo)
	if ok == false || len(fileInfo.Data) == 0 {
		//为空，返回404
		this.Ctx.WriteStatus(404)
		this.Ctx.Write([]byte("File Not Found"))
	} else {
		//不为空，调用http.ServeContent(xxxx)
		this.Ctx.WriteHeader("Etag", fileInfo.Etag)
		if fileInfo.Encoding != "" {
			this.Ctx.WriteHeader("Content-Encoding", fileInfo.Encoding)
		} else {
			this.Ctx.WriteHeader("Content-Length", strconv.Itoa(len(fileInfo.Data)))
		}
		http.ServeContent(
			this.Ctx.GetRawResponseWriter().(http.ResponseWriter),
			this.Ctx.GetRawRequest().(*http.Request),
			fileInfo.Name,
			fileInfo.ModifyTime,
			bytes.NewReader(fileInfo.Data),
		)
	}
}

func (this *BaseController) rawRender(result baseControllerResult) {
	if result.Code != 0 {
		panic(result.Msg)
	}
	this.Write(result.Data.([]byte))
}

func (this *BaseController) jsonRender(result baseControllerResult) {
	if result.Data == nil {
		result.Data = ""
	}
	this.Ctx.WriteHeader("Content-Type", "text/javascript;charset=utf-8")
	resultString, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	this.Ctx.Write(resultString)
}

// websocket
func (this *BaseController) websocketRender(result baseControllerResult) {
	// TODO
}

func (this *BaseController) redirectRender(result baseControllerResult) {
	//FIXME 没有做更多的容错尝试
	if result.Code == 0 {
		if result.Data == nil {
			panic(errors.New("未指定跳转链接!"))
		}

		//使用页面跳转，解决跨域问题
		var outputResult struct {
			Url string
		}
		outputResult.Url = result.Data.(string)
		t, err := template.ParseFiles("views/redirect.html")
		if err != nil {
			panic(err)
		}

		dataBuffer := bytes.NewBuffer(nil)
		err = t.Execute(dataBuffer, outputResult)
		if err != nil {
			panic(err)
		}
		this.Ctx.Write(dataBuffer.Bytes())
	} else {
		this.Ctx.Write([]byte("跳转不成功 " + result.Msg))
	}
}

func (this *BaseController) excelRender(result baseControllerResult) {
	//获取excel的导出配置
	var excelArgs struct {
		ViewTitle  string `validate:"_viewTitle"`
		ViewFormat string `validate:"_viewFormat"`
	}
	this.Check(&excelArgs)

	excelTitle := excelArgs.ViewTitle
	excelFormat, err := DecodeUrl(excelArgs.ViewFormat)
	if err != nil {
		panic(err)
	}
	jsonFormat := map[string]string{}
	err = json.Unmarshal([]byte(excelFormat), &jsonFormat)
	if err != nil {
		panic(err)
	}
	jsonData := reflect.ValueOf(result.Data).FieldByName("Data").Interface()
	tableData := ArrayColumnTable(jsonFormat, jsonData)

	//写入数据
	resultByte, err := EncodeXlsx(tableData)
	if err != nil {
		panic(err)
	}
	this.Ctx.WriteMimeHeader("xlsx", excelTitle)
	this.Ctx.Write(resultByte)
}

func (this *BaseController) downloadRender(result baseControllerResult) {
	title := reflect.ValueOf(result.Data).FieldByName("Title").Interface().(string)
	data := reflect.ValueOf(result.Data).FieldByName("Data").Interface().([]byte)
	this.Ctx.WriteMimeHeader("xlsx", title)
	this.Ctx.Write(data)
}

func (this *BaseController) AutoRender(returnValue interface{}, renderName string) {
	if this.Ctx.GetMethod() == "OPTIONS" {
		this.initCors()
		return
	}

	result := baseControllerResult{}
	resultError, ok := returnValue.(Exception)

	if ok {
		//带错误码的error
		result.Code = resultError.GetCode()
		result.Msg = resultError.GetMessage()
		result.Data = nil
	} else {
		//正常返回
		result.Code = 0
		result.Data = returnValue
		result.Msg = ""
	}
	this.initCors()

	var inputViewName struct {
		View string `validate:"_view"`
	}
	this.Check(&inputViewName)

	if inputViewName.View == "excel" || renderName == "excel" {
		this.excelRender(result)
	} else if renderName == "raw" {
		this.rawRender(result)
	} else if renderName == "download" {
		this.downloadRender(result)
	} else if renderName == "file" {
		this.fileRender(result)
	} else if renderName == "json" {
		this.jsonRender(result)
	} else if renderName == "websocket" {
		this.websocketRender(result)
	} else if renderName == "redirect" {
		this.redirectRender(result)
	} else {
		panic("不合法的renderName " + renderName)
	}

}
