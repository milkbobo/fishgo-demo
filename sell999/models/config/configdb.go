package config

import (
	. "sell999/models/common"

	. "github.com/milkbobo/fishgoweb/language"
	. "github.com/milkbobo/fishgoweb/web"
)

type ConfigDbModel struct {
	BaseModel
}

func (this *ConfigDbModel) Search(where Config, limit CommonPage) Configs {
	result := Configs{}

	if limit.PageSize == 0 {
		return result
	}

	db := this.DB.NewSession()
	defer db.Close()

	db = this.searchWhere(db, where)
	count, err := db.Count(&Config{})
	if err != nil {
		panic(err)
	}
	result.Count = int(count)

	data := []Config{}
	db = this.searchWhere(db, where)
	if limit.PageSize > 0 {
		db = db.Limit(limit.PageSize, limit.PageIndex)
	}
	err = db.OrderBy("configId desc").Find(&data)
	if err != nil {
		panic(err)
	}
	result.Data = data

	return result
}

func (this *ConfigDbModel) searchWhere(db DatabaseSession, where Config) DatabaseSession {
	if where.Name != "" {
		db = db.And("name like ?", "%"+where.Name+"%")
	}
	return db
}

func (this *ConfigDbModel) Get(configId int) Config {
	result := []Config{}
	err := this.DB.Where("configId = ?", configId).Find(&result)
	if err != nil {
		panic(err)
	}
	if len(result) == 0 {
		Throw(1, "找不到此配置")
	}
	return result[0]
}

func (this *ConfigDbModel) Add(data Config) {
	_, err := this.DB.Insert(&data)
	if err != nil {
		panic(err)
	}
}

func (this *ConfigDbModel) Mod(configId int, data Config) {
	_, err := this.DB.Where("configId = ?", configId).Update(&data)
	if err != nil {
		panic(err)
	}
}

func (this *ConfigDbModel) GetByName(name string) []Config {
	var configs []Config
	err := this.DB.Where("name=?", name).Find(&configs)
	if err != nil {
		panic(err)
	}
	return configs
}