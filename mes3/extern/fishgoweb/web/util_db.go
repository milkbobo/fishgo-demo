package web

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/core"
	"xorm.io/xorm"
)

type DatabaseCommon interface {
	SQL(querystring string, args ...interface{}) DatabaseSession
	NoAutoTime() DatabaseSession
	NoAutoCondition(no ...bool) DatabaseSession
	Cascade(trueOrFalse ...bool) DatabaseSession
	Where(querystring string, args ...interface{}) DatabaseSession
	ID(id interface{}) DatabaseSession
	Distinct(columns ...string) DatabaseSession
	Select(str string) DatabaseSession
	Cols(columns ...string) DatabaseSession
	AllCols() DatabaseSession
	MustCols(columns ...string) DatabaseSession
	UseBool(columns ...string) DatabaseSession
	Omit(columns ...string) DatabaseSession
	Nullable(columns ...string) DatabaseSession
	In(column string, args ...interface{}) DatabaseSession
	Incr(column string, arg ...interface{}) DatabaseSession
	Decr(column string, arg ...interface{}) DatabaseSession
	SetExpr(column string, expression string) DatabaseSession
	Table(tableNameOrBean interface{}) DatabaseSession
	Alias(alias string) DatabaseSession
	Limit(limit int, start ...int) DatabaseSession
	Desc(colNames ...string) DatabaseSession
	Asc(colNames ...string) DatabaseSession
	OrderBy(order string) DatabaseSession
	Join(join_operator string, tablename interface{}, condition string, args ...interface{}) DatabaseSession
	GroupBy(keys string) DatabaseSession
	Having(conditions string) DatabaseSession
	Exec(args ...interface{}) (sql.Result, error)
	Query(...interface{}) (resultsSlice []map[string][]byte, err error)
	Insert(beans ...interface{}) (int64, error)
	InsertOne(bean interface{}) (int64, error)
	Update(bean interface{}, condiBeans ...interface{}) (int64, error)
	Delete(bean ...interface{}) (int64, error)
	Get(bean ...interface{}) (bool, error)
	Find(beans interface{}, condiBeans ...interface{}) error
	Count(bean ...interface{}) (int64, error)
}

type DatabaseSession interface {
	DatabaseCommon
	Close() error
	And(querystring string, args ...interface{}) DatabaseSession
	Or(querystring string, args ...interface{}) DatabaseSession
	ForUpdate() DatabaseSession
	Begin() error
	Commit() error
	LastSQL() (string, []interface{})
}

type Database interface {
	DatabaseCommon
	Close() error
	NewSession() DatabaseSession
	GetStats() sql.DBStats
}

type DatabaseConfig struct {
	Driver            string
	Host              string
	Port              int
	User              string
	Password          string
	Charset           string
	Collation         string
	Database          string
	Debug             bool
	MaxConnection     int
	MaxIdleConnection int
}

type databaseImplement struct {
	*xorm.Engine
	config DatabaseConfig
}

type databaseSessionImplement struct {
	*xorm.Session
}

func NewDatabase(config DatabaseConfig) (Database, error) {
	if config.Driver == "" {
		return nil, nil
	}
	if config.Charset == "" {
		config.Charset = "utf8"
	}
	if config.Collation == "" {
		config.Collation = "utf8_general_ci"
	}
	dblink := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%v&collation=%v&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
		config.Collation,
	)
	tempDb, err := xorm.NewEngine(config.Driver, dblink)
	if err != nil {
		return nil, err
	}

	tempDb.SetTableMapper(&tableMapper{})
	tempDb.SetColumnMapper(&columnMapper{})
	if config.Debug {
		tempDb.ShowSQL(true)
	}
	if config.MaxConnection > 0 {
		tempDb.SetMaxOpenConns(config.MaxConnection)
	}
	if config.MaxIdleConnection > 0 {
		tempDb.SetMaxIdleConns(config.MaxIdleConnection)
		tempDb.DB().SetConnMaxLifetime(time.Hour * 3)
	}
	tempDb.Ping()
	return &databaseImplement{
		Engine: tempDb,
		config: config,
	}, nil
}

func NewDatabaseFromConfig(configName string) (Database, error) {
	config := DatabaseConfig{}
	switch configName {
	case "db":
		config = DatabaseConfig(globalBasic.Config.Get().DB)
	case "db2":
		config = DatabaseConfig(globalBasic.Config.Get().DB2)
	case "db3":
		config = DatabaseConfig(globalBasic.Config.Get().DB3)
	case "db4":
		config = DatabaseConfig(globalBasic.Config.Get().DB4)
	case "db5":
		config = DatabaseConfig(globalBasic.Config.Get().DB5)
	}

	return NewDatabase(config)
}

type zeroable interface {
	IsZero() bool
}

func (this *databaseImplement) GetStats() sql.DBStats {
	return this.Engine.DB().Stats()
}

func (this *databaseImplement) rValue(bean interface{}) reflect.Value {
	return reflect.Indirect(reflect.ValueOf(bean))
}

func (this *databaseImplement) isZero(k interface{}) bool {
	switch k.(type) {
	case int:
		return k.(int) == 0
	case int8:
		return k.(int8) == 0
	case int16:
		return k.(int16) == 0
	case int32:
		return k.(int32) == 0
	case int64:
		return k.(int64) == 0
	case uint:
		return k.(uint) == 0
	case uint8:
		return k.(uint8) == 0
	case uint16:
		return k.(uint16) == 0
	case uint32:
		return k.(uint32) == 0
	case uint64:
		return k.(uint64) == 0
	case float32:
		return k.(float32) == 0
	case float64:
		return k.(float64) == 0
	case bool:
		return k.(bool) == false
	case string:
		return k.(string) == ""
	case zeroable:
		return k.(zeroable).IsZero()
	}
	return false
}

func (this *databaseImplement) value2Interface(fieldValue reflect.Value) (interface{}, error) {
	fieldType := fieldValue.Type()
	fieldTypeKind := fieldType.Kind()
	switch fieldTypeKind {
	case reflect.Bool:
		return fieldValue.Bool(), nil
	case reflect.String:
		return fieldValue.String(), nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return fieldValue.Int(), nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return fieldValue.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return fieldValue.Float(), nil
	case reflect.Struct:
		if fieldType == reflect.TypeOf(time.Time{}) {
			t := fieldValue.Interface().(time.Time)
			tf := t.Format("2006-01-02 15:04:05")
			return tf, nil
		} else {
			return nil, fmt.Errorf("Unsupported type %v", fieldType)
		}
	default:
		return nil, fmt.Errorf("Unsupported type %v", fieldType)
	}
}

type tableName interface {
	TableName() string
}

func (this *databaseImplement) autoMapType(v reflect.Value) *core.Table {
	t := v.Type()
	table := core.NewEmptyTable()
	if tb, ok := v.Interface().(tableName); ok {
		table.Name = tb.TableName()
	} else {
		if v.CanAddr() {
			if tb, ok = v.Addr().Interface().(tableName); ok {
				table.Name = tb.TableName()
			}
		}
		if table.Name == "" {
			table.Name = this.GetTableMapper().Obj2Table(t.Name())
		}
	}
	table.Type = t
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag
		ormTagStr := tag.Get("xorm")
		if ormTagStr == "-" || ormTagStr == "<-" {
			continue
		}
		col := &core.Column{FieldName: t.Field(i).Name, Nullable: true, IsPrimaryKey: false,
			IsAutoIncrement: false, MapType: core.TWOSIDES, Indexes: make(map[string]int)}
		col.Name = this.GetTableMapper().Obj2Table(t.Field(i).Name)
		table.AddColumn(col)
	}
	return table
}

func newDatabaseSession(sess *xorm.Session) DatabaseSession {
	return &databaseSessionImplement{Session: sess}
}

func (this *databaseImplement) NewSession() DatabaseSession {
	return newDatabaseSession(this.Engine.NewSession())
}

func (this *databaseImplement) SQL(querystring string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Engine.SQL(querystring, args...))
}

func (this *databaseImplement) NoAutoTime() DatabaseSession {
	return newDatabaseSession(this.Engine.NoAutoTime())
}

func (this *databaseImplement) NoAutoCondition(no ...bool) DatabaseSession {
	return newDatabaseSession(this.Engine.NoAutoCondition(no...))
}

func (this *databaseImplement) Cascade(trueOrFalse ...bool) DatabaseSession {
	return newDatabaseSession(this.Engine.Cascade(trueOrFalse...))
}

func (this *databaseImplement) Where(querystring string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Engine.Where(querystring, args...))
}

func (this *databaseImplement) ID(id interface{}) DatabaseSession {
	return newDatabaseSession(this.Engine.ID(id))
}

func (this *databaseImplement) Distinct(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Engine.Distinct(columns...))
}

func (this *databaseImplement) Select(str string) DatabaseSession {
	return newDatabaseSession(this.Engine.Select(str))
}

func (this *databaseImplement) Cols(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Engine.Cols(columns...))
}

func (this *databaseImplement) AllCols() DatabaseSession {
	return newDatabaseSession(this.Engine.AllCols())
}

func (this *databaseImplement) MustCols(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Engine.MustCols(columns...))
}

func (this *databaseImplement) UseBool(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Engine.UseBool(columns...))
}

func (this *databaseImplement) Omit(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Engine.Omit(columns...))
}

func (this *databaseImplement) Nullable(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Engine.Nullable(columns...))
}

func (this *databaseImplement) In(column string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Engine.In(column, args...))
}

func (this *databaseImplement) Incr(column string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Engine.Incr(column, args...))
}

func (this *databaseImplement) Decr(column string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Engine.Decr(column, args...))
}

func (this *databaseImplement) SetExpr(column string, expression string) DatabaseSession {
	return newDatabaseSession(this.Engine.SetExpr(column, expression))
}

func (this *databaseImplement) Table(tableNameOrBean interface{}) DatabaseSession {
	return newDatabaseSession(this.Engine.Table(tableNameOrBean))
}

func (this *databaseImplement) Alias(alias string) DatabaseSession {
	return newDatabaseSession(this.Engine.Alias(alias))
}

func (this *databaseImplement) Limit(limit int, start ...int) DatabaseSession {
	//修复xorm的PageSize为0时，仍然不分页的问题
	if limit == 0 {
		start = []int{1}
	}
	return newDatabaseSession(this.Engine.Limit(limit, start...))
}

func (this *databaseImplement) Desc(colNames ...string) DatabaseSession {
	return newDatabaseSession(this.Engine.Desc(colNames...))
}

func (this *databaseImplement) Asc(colNames ...string) DatabaseSession {
	return newDatabaseSession(this.Engine.Asc(colNames...))
}

func (this *databaseImplement) OrderBy(order string) DatabaseSession {
	return newDatabaseSession(this.Engine.OrderBy(order))
}

func (this *databaseImplement) Join(join_operator string, tablename interface{}, condition string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Engine.Join(join_operator, tablename, condition, args...))
}

func (this *databaseImplement) GroupBy(keys string) DatabaseSession {
	return newDatabaseSession(this.Engine.GroupBy(keys))
}

func (this *databaseImplement) Having(conditions string) DatabaseSession {
	return newDatabaseSession(this.Engine.Having(conditions))
}

type tableMapper struct {
}

func (this *tableMapper) Obj2Table(name string) string {
	result := []rune{}
	result = append(result, 't')
	for _, chr := range name {
		if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
			result = append(result, '_')
			chr -= ('A' - 'a')
		}
		result = append(result, chr)
	}
	return string(result)
}

func (this *tableMapper) Table2Obj(in string) string {
	fmt.Println("Obj2Table2 " + in)
	return in
}

type columnMapper struct {
}

func (this *columnMapper) Obj2Table(name string) string {
	return strings.ToLower(name[0:1]) + name[1:]
}

func (this *columnMapper) Table2Obj(in string) string {
	fmt.Println("Obj2Table4 " + in)
	return in
}

func (this *databaseSessionImplement) SQL(querystring string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.SQL(querystring, args...))
}

func (this *databaseSessionImplement) NoAutoTime() DatabaseSession {
	return newDatabaseSession(this.Session.NoAutoTime())
}

func (this *databaseSessionImplement) NoAutoCondition(no ...bool) DatabaseSession {
	return newDatabaseSession(this.Session.NoAutoCondition(no...))
}

func (this *databaseSessionImplement) Cascade(trueOrFalse ...bool) DatabaseSession {
	return newDatabaseSession(this.Session.Cascade(trueOrFalse...))
}

func (this *databaseSessionImplement) Where(querystring string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.Where(querystring, args...))
}

func (this *databaseSessionImplement) ID(id interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.ID(id))
}

func (this *databaseSessionImplement) Distinct(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Session.Distinct(columns...))
}

func (this *databaseSessionImplement) Select(str string) DatabaseSession {
	return newDatabaseSession(this.Session.Select(str))
}

func (this *databaseSessionImplement) Cols(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Session.Cols(columns...))
}

func (this *databaseSessionImplement) AllCols() DatabaseSession {
	return newDatabaseSession(this.Session.AllCols())
}

func (this *databaseSessionImplement) MustCols(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Session.MustCols(columns...))
}

func (this *databaseSessionImplement) UseBool(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Session.UseBool(columns...))
}

func (this *databaseSessionImplement) Omit(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Session.Omit(columns...))
}

func (this *databaseSessionImplement) Nullable(columns ...string) DatabaseSession {
	return newDatabaseSession(this.Session.Nullable(columns...))
}

func (this *databaseSessionImplement) In(column string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.In(column, args...))
}

func (this *databaseSessionImplement) Incr(column string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.Incr(column, args...))
}

func (this *databaseSessionImplement) Decr(column string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.Decr(column, args...))
}

func (this *databaseSessionImplement) SetExpr(column string, expression string) DatabaseSession {
	return newDatabaseSession(this.Session.SetExpr(column, expression))
}

func (this *databaseSessionImplement) Table(tableNameOrBean interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.Table(tableNameOrBean))
}

func (this *databaseSessionImplement) Alias(alias string) DatabaseSession {
	return newDatabaseSession(this.Session.Alias(alias))
}

func (this *databaseSessionImplement) Limit(limit int, start ...int) DatabaseSession {
	//修复xorm的PageSize为0时，仍然不分页的问题
	if limit == 0 {
		start = []int{1}
	}
	return newDatabaseSession(this.Session.Limit(limit, start...))
}

func (this *databaseSessionImplement) Desc(colNames ...string) DatabaseSession {
	return newDatabaseSession(this.Session.Desc(colNames...))
}

func (this *databaseSessionImplement) Asc(colNames ...string) DatabaseSession {
	return newDatabaseSession(this.Session.Asc(colNames...))
}

func (this *databaseSessionImplement) OrderBy(order string) DatabaseSession {
	return newDatabaseSession(this.Session.OrderBy(order))
}

func (this *databaseSessionImplement) Join(join_operator string, tablename interface{}, condition string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.Join(join_operator, tablename, condition, args...))
}

func (this *databaseSessionImplement) GroupBy(keys string) DatabaseSession {
	return newDatabaseSession(this.Session.GroupBy(keys))
}

func (this *databaseSessionImplement) Having(conditions string) DatabaseSession {
	return newDatabaseSession(this.Session.Having(conditions))
}

func (this *databaseSessionImplement) And(querystring string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.And(querystring, args...))
}

func (this *databaseSessionImplement) Or(querystring string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.Or(querystring, args...))
}

func (this *databaseSessionImplement) ForUpdate() DatabaseSession {
	return newDatabaseSession(this.Session.ForUpdate())
}
