package database

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/core"
	"xorm.io/xorm"
	_ "github.com/mattn/go-sqlite3"
)

type DatabaseCommon interface {
	Sql(querystring string, args ...interface{}) DatabaseSession
	NoAutoTime() DatabaseSession
	NoAutoCondition(no ...bool) DatabaseSession
	Cascade(trueOrFalse ...bool) DatabaseSession
	Where(querystring string, args ...interface{}) DatabaseSession
	Id(id interface{}) DatabaseSession
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

	Exec(sql string, args ...interface{}) (sql.Result, error)
	MustExec(sql string, args ...interface{}) sql.Result

	Query(sql string, paramStr ...interface{}) (resultsSlice []map[string][]byte, err error)
	MustQuery(sql string, paramStr ...interface{}) []map[string][]byte

	Insert(beans ...interface{}) (int64, error)
	MustInsert(beans ...interface{}) int64

	InsertOne(bean interface{}) (int64, error)
	MustInsertOne(bean interface{}) int64

	Update(bean interface{}, condiBeans ...interface{}) (int64, error)
	MustUpdate(bean interface{}, condiBeans ...interface{}) int64

	Delete(bean interface{}) (int64, error)
	MustDelete(bean interface{}) int64

	Get(bean interface{}) (bool, error)
	MustGet(bean interface{}) bool

	Find(beans interface{}, condiBeans ...interface{}) error
	MustFind(beans interface{}, condiBeans ...interface{})

	Count(bean interface{}) (int64, error)
	MustCount(bean interface{}) int64
}

type DatabaseSession interface {
	DatabaseCommon
	Close()
	And(querystring string, args ...interface{}) DatabaseSession
	Or(querystring string, args ...interface{}) DatabaseSession
	ForUpdate() DatabaseSession

	Begin() error
	MustBegin()

	Commit() error
	MustCommit()

	LastSQL() (string, []interface{})
}

type Database interface {
	DatabaseCommon
	Close() error
	NewSession() DatabaseSession
	UpdateBatch(rowsSlicePtr interface{}, indexColName string) (int64, error)
	MustUpdateBatch(rowSlicePtr interface{}, indexColName string) int64
}

type DatabaseConfig struct {
	Driver            string `config:"driver"`
	Host              string `config:"host"`
	Port              int    `config:"port"`
	User              string `config:"user"`
	Passowrd          string `config:"password"`
	Charset           string `config:"charset"`
	Collation         string `config:"collation"`
	Database          string `config:"database"`
	Debug             bool   `config:"debug"`
	MaxConnection     int    `config:"maxconnection"`
	MaxIdleConnection int    `config:"maxidleconnection"`
}

type databaseImplement struct {
	*xorm.Engine
	config DatabaseConfig
}

type databaseSessionImplement struct {
	*xorm.Session
}

func NewDatabaseTest() Database {
	db, err := NewDatabase(DatabaseConfig{
		Driver:   "sqlite3",
		Database: ":memory:",
	})
	if err != nil {
		panic(err)
	}
	return db
}

func NewDatabase(config DatabaseConfig) (Database, error) {
	var dblink string
	if config.Driver == "mysql" {
		if config.Charset == "" {
			config.Charset = "utf8mb4"
		}
		if config.Collation == "" {
			config.Collation = "utf8_general_ci"
		}
		dblink = fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=%v&collation=%v&loc=Local",
			config.User,
			config.Passowrd,
			config.Host,
			config.Port,
			config.Database,
			config.Charset,
			config.Collation,
		)
	} else if config.Driver == "sqlite3" {
		dblink = config.Database
	} else {
		return nil, errors.New("invalid database driver " + config.Driver)
	}
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
	if config.MaxIdleConnection <= 0 {
		config.MaxIdleConnection = 100
	}
	tempDb.SetMaxIdleConns(config.MaxIdleConnection)
	tempDb.DB().SetConnMaxLifetime(time.Hour * 7)
	tempDb.Ping()
	return &databaseImplement{
		Engine: tempDb,
		config: config,
	}, nil
}

type zeroable interface {
	IsZero() bool
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
			table.Name = this.TableMapper.Obj2Table(t.Name())
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
			IsAutoIncrement: false, MapType: core.TWOSIDES, Indexes: make(map[string]bool)}
		col.Name = this.ColumnMapper.Obj2Table(t.Field(i).Name)
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

func (this *databaseImplement) Sql(querystring string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Engine.Sql(querystring, args...))
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

func (this *databaseImplement) Id(id interface{}) DatabaseSession {
	return newDatabaseSession(this.Engine.Id(id))
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
	//??????xorm???PageSize???0??????????????????????????????
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

func (this *databaseImplement) MustExec(sql string, args ...interface{}) sql.Result {
	result, err := this.Exec(sql, args...)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseImplement) MustQuery(sql string, paramStr ...interface{}) []map[string][]byte {
	result, err := this.Query(sql, paramStr...)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseImplement) MustInsert(beans ...interface{}) int64 {
	result, err := this.Insert(beans...)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseImplement) MustInsertOne(bean interface{}) int64 {
	result, err := this.InsertOne(bean)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseImplement) MustUpdate(bean interface{}, condiBeans ...interface{}) int64 {
	result, err := this.Update(bean, condiBeans...)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseImplement) MustDelete(bean interface{}) int64 {
	result, err := this.Delete(bean)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseImplement) MustGet(bean interface{}) bool {
	result, err := this.Get(bean)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseImplement) MustFind(beans interface{}, condiBeans ...interface{}) {
	err := this.Find(beans, condiBeans...)
	if err != nil {
		panic(err)
	}
}

func (this *databaseImplement) MustCount(bean interface{}) int64 {
	result, err := this.Count(bean)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseImplement) UpdateBatch(rowsSlicePtr interface{}, indexColName string) (int64, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice {
		return 0, errors.New("needs a pointer to a slice")
	}
	if sliceValue.Len() == 0 {
		return 0, errors.New("update rows is empty")
	}

	bean := sliceValue.Index(0).Interface()
	elementValue := this.rValue(bean)
	table := this.autoMapType(elementValue)
	size := sliceValue.Len()

	var rows = make([][]interface{}, 0)
	var indexRow = make([]interface{}, 0)
	cols := make([]*core.Column, 0)
	updateCols := make([]bool, 0)
	var indexCol *core.Column

	//????????????
	for i := 0; i < size; i++ {
		v := sliceValue.Index(i)
		vv := reflect.Indirect(v)

		//???????????????update??????
		if i == 0 {
			for _, col := range table.Columns() {
				if col.Name == indexColName {
					indexCol = col
				} else {
					cols = append(cols, col)
					updateCols = append(updateCols, false)
				}
			}
			if indexCol == nil {
				return 0, errors.New("counld not found index col " + indexColName)
			}
		}

		//???????????????update??????
		var singleRow = make([]interface{}, 0)
		for colIndex, col := range cols {
			ptrFieldValue, err := col.ValueOfV(&vv)
			if err != nil {
				return 0, err
			}
			fieldValue := *ptrFieldValue
			var arg interface{}
			if this.isZero(fieldValue.Interface()) {
				arg = nil
			} else {
				var err error
				arg, err = this.value2Interface(fieldValue)
				if err != nil {
					return 0, err
				}
				updateCols[colIndex] = true
			}
			singleRow = append(singleRow, arg)
		}
		rows = append(rows, singleRow)
		ptrFieldValue, err := indexCol.ValueOfV(&vv)
		if err != nil {
			return 0, err
		}
		fieldValue := *ptrFieldValue
		arg, err := this.value2Interface(fieldValue)
		if err != nil {
			return 0, err
		}
		indexRow = append(indexRow, arg)
	}
	if len(cols) == 0 {
		return 0, errors.New("update cols is empty! " + fmt.Sprintf("%v", rowsSlicePtr))
	}

	//??????sql
	var sqlArgs = make([]interface{}, 0)
	var sql = "UPDATE " + table.Name + " SET "
	var isFirstUpdateCol = true
	for colIndex, col := range cols {
		if updateCols[colIndex] == false {
			continue
		}
		if isFirstUpdateCol == false {
			sql += " , "
		}
		sql += this.Engine.QuoteStr() + col.Name + this.Engine.QuoteStr()
		sql += " = CASE "
		sql += this.Engine.QuoteStr() + indexCol.Name + this.Engine.QuoteStr()
		for rowIndex, row := range rows {
			if row[colIndex] == nil {
				continue
			}
			sql += " WHEN ? THEN ? "
			sqlArgs = append(sqlArgs, indexRow[rowIndex])
			sqlArgs = append(sqlArgs, row[colIndex])
		}
		sql += " END "
		isFirstUpdateCol = false
	}
	sql += " WHERE " + this.Engine.QuoteStr() + indexCol.Name + this.Engine.QuoteStr() + " IN ( "
	for rowIndex, row := range indexRow {
		if rowIndex != 0 {
			sql += " , "
		}
		sql += " ? "
		sqlArgs = append(sqlArgs, row)
	}
	sql += " ) "

	//??????sql
	res, err := this.Exec(sql, sqlArgs...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (this *databaseImplement) MustUpdateBatch(rowsSlicePtr interface{}, indexColName string) int64 {
	result, err := this.UpdateBatch(rowsSlicePtr, indexColName)
	if err != nil {
		panic(err)
	}
	return result
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

func (this *databaseSessionImplement) Sql(querystring string, args ...interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.Sql(querystring, args...))
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

func (this *databaseSessionImplement) Id(id interface{}) DatabaseSession {
	return newDatabaseSession(this.Session.Id(id))
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
	//??????xorm???PageSize???0??????????????????????????????
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

func (this *databaseSessionImplement) MustExec(sql string, args ...interface{}) sql.Result {
	result, err := this.Exec(sql, args...)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseSessionImplement) MustQuery(sql string, paramStr ...interface{}) []map[string][]byte {
	result, err := this.Query(sql, paramStr...)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseSessionImplement) MustInsert(beans ...interface{}) int64 {
	result, err := this.Insert(beans...)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseSessionImplement) MustInsertOne(bean interface{}) int64 {
	result, err := this.InsertOne(bean)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseSessionImplement) MustUpdate(bean interface{}, condiBeans ...interface{}) int64 {
	result, err := this.Update(bean, condiBeans...)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseSessionImplement) MustDelete(bean interface{}) int64 {
	result, err := this.Delete(bean)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseSessionImplement) MustGet(bean interface{}) bool {
	result, err := this.Get(bean)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseSessionImplement) MustFind(beans interface{}, condiBeans ...interface{}) {
	err := this.Find(beans, condiBeans...)
	if err != nil {
		panic(err)
	}
}

func (this *databaseSessionImplement) MustCount(bean interface{}) int64 {
	result, err := this.Count(bean)
	if err != nil {
		panic(err)
	}
	return result
}

func (this *databaseSessionImplement) MustBegin() {
	err := this.Begin()
	if err != nil {
		panic(err)
	}
}

func (this *databaseSessionImplement) MustCommit() {
	err := this.Commit()
	if err != nil {
		panic(err)
	}
}
