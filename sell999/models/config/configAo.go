package config

import (
	. "sell999/models/common"

	. "github.com/milkbobo/fishgoweb/language"
)

/*
* 系统，sdk配置接口
 */
type ConfigAoModel struct {
	BaseModel
	ConfigDb ConfigDbModel
}

/*
* 后台接口
 */
//搜索
func (this *ConfigAoModel) Search(where Config, limit CommonPage) Configs {
	return this.ConfigDb.Search(where, limit)
}

//名字取列表
func (this *ConfigAoModel) GetByNames(names []string) map[string]string {
	result := map[string]string{}
	for _, single := range names {
		result[single] = this.Get(single)
	}
	return result
}

//名字设列表
func (this *ConfigAoModel) SetByNames(names map[string]string) {
	for name, value := range names {
		this.Set(name, value)
	}
}

/*
* 前端接口
 */
//名字取
func (this *ConfigAoModel) Get(name string) string {
	result := ""

	configs := this.ConfigDb.GetByName(name)
	if len(configs) == 0 {
		result = ""
	} else {
		result = configs[0].Value
	}

	return result
}

//名字设
func (this *ConfigAoModel) Set(name string, value string) {
	if name == "" {
		Throw(1, "键名不能为空！")
	}

	configs := this.ConfigDb.GetByName(name)
	if len(configs) == 0 {
		this.ConfigDb.Add(Config{
			Name:  name,
			Value: value,
		})
	} else {
		this.ConfigDb.Mod(configs[0].ConfigId, Config{
			Value: value,
		})
	}
}
