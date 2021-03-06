package sqlf

import (
	"encoding/json"
	. "github.com/milkbobo/fishgoweb/app/log"
	. "github.com/milkbobo/fishgoweb/assert"
	. "github.com/milkbobo/fishgoweb/language"
	"testing"
	"time"
)

func initSqliteDatabase() SqlfDB {
	log, err := NewLog(LogConfig{
		Driver: "console",
	})
	if err != nil {
		panic(err)
	}
	db, err := NewSqlfDB(log, nil, SqlfDBConfig{
		Driver:     "sqlite3",
		SourceName: ":memory:?_loc=auto",
		Debug:      true,
	})
	if err != nil {
		panic(err)
	}
	db.MustExec(`
	create table t_user(
		userId integer primary key autoincrement,
		name char(32) not null,
		age integer not null,
		money decimal(14,2) not null,
		loginTime timestamp not null,
		createTime timestamp not null default 0,
		modifyTime timestamp not null default 0
	);

	create table t_item(
		itemId integer primary key autoincrement,
		name char(32) not null,
		onShelfTime timestamp not null,
		createTime timestamp not null default 0,
		modifyTime timestamp not null default 0
	);

	create table t_article(
		articleId integer primary key autoincrement,
		data text not null,
		remark text not null,
		createTime timestamp not null default 0,
		modifyTime timestamp not null default 0
	);
	`)
	return db
}

func initMySqlDatabase() SqlfDB {
	log, err := NewLog(LogConfig{
		Driver: "console",
	})
	if err != nil {
		panic(err)
	}
	db, err := NewSqlfDB(log, nil, SqlfDBConfig{
		Driver:     "mysql",
		SourceName: "root:Yinghao23367847@tcp(localhost:3306)/test?parseTime=true&loc=Local",
		Debug:      true,
	})
	if err != nil {
		panic(err)
	}
	db.MustExec(`
	drop table if exists t_user;
	`)
	db.MustExec(`
	drop table if exists t_item;
	`)
	db.MustExec(`
	drop table if exists t_article;
	`)
	db.MustExec(`
	create table t_user(
		userId int not null auto_increment,
		name char(32) not null,
		age integer not null,
		money decimal(14,2) not null,
		loginTime datetime not null,
		createTime datetime not null default '1970-01-01 08:00:00',
		modifyTime datetime not null default '1970-01-01 08:00:00',
		primary key(userId)
	)engine=innodb default charset=utf8mb4;`)

	db.MustExec(`
	create table t_item(
		itemId integer not null auto_increment,
		name char(32) not null,
		onShelfTime datetime not null default '1970-01-01 08:00:00',
		createTime datetime not null default '1970-01-01 08:00:00',
		modifyTime datetime not null default '1970-01-01 08:00:00',
		primary key(itemId)
	)engine=innodb default charset=utf8mb4;`)

	db.MustExec(`
	create table t_article(
		articleId integer not null auto_increment,
		data mediumtext not null,
		remark mediumtext not null,
		createTime datetime not null default '1970-01-01 08:00:00',
		modifyTime datetime not null default '1970-01-01 08:00:00',
		primary key(articleId)
	)engine=innodb default charset=utf8mb4;`)
	return db
}

func checkNowTime(t *testing.T, inTime time.Time) {
	now := time.Now()
	AssertEqual(t, now.Sub(inTime) <= time.Second, true)
}

func checkTime(t *testing.T, inTime time.Time, targetTime time.Time) {
	AssertEqual(t, targetTime.Sub(inTime) <= time.Second || targetTime.Sub(inTime) >= time.Second, true)
}

type User struct {
	UserId     int `sqlf:"autoincr"`
	Name       string
	Age        int
	Money      Decimal
	LoginTime  time.Time
	CreateTime time.Time `sqlf:"created"`
	ModifyTime time.Time `sqlf:"updated"`
}

type Article struct {
	ArticleId  int `sqlf:"autoincr"`
	Data       []byte
	Remark     json.RawMessage
	CreateTime time.Time `sqlf:"created"`
	ModifyTime time.Time `sqlf:"updated"`
}

func testStructType(t *testing.T, db SqlfCommon) {
	//????????????????????????
	users := []User{}
	db.MustQuery(&users, "select * from t_user", User{})

	AssertEqual(t, users, []User{})

	//????????????????????????????????????
	userAdds := []User{
		User{Name: "fish", Age: 12, Money: "", LoginTime: time.Unix(1, 0)},
		User{Name: "cat", Age: 34, Money: "102.35", LoginTime: time.Unix(2, 0)},
	}
	db.MustExec("insert into t_user(?.insertColumn) values ?.insertValue", userAdds, userAdds)

	db.MustQuery(&users, "select ?.column from t_user", users)

	now := time.Now()
	for _, user := range users {
		checkNowTime(t, user.CreateTime)
		checkNowTime(t, user.ModifyTime)
	}
	db.MustExec("update t_user set createTime = ?,modifyTime = ?", time.Unix(0, 0), time.Unix(0, 0))

	db.MustQuery(&users, "select * from t_user")

	AssertEqual(t, users, []User{
		User{UserId: 1, Name: "fish", Age: 12, Money: "0", LoginTime: time.Unix(1, 0), CreateTime: time.Unix(0, 0), ModifyTime: time.Unix(0, 0)},
		User{UserId: 2, Name: "cat", Age: 34, Money: "102.35", LoginTime: time.Unix(2, 0), CreateTime: time.Unix(0, 0), ModifyTime: time.Unix(0, 0)},
	})

	//??????????????????
	db.MustExec("delete from t_user where userId = ?", 1)

	db.MustQuery(&users, "select ?.column from t_user", users)

	AssertEqual(t, users, []User{
		User{UserId: 2, Name: "cat", Age: 34, Money: "102.35", LoginTime: time.Unix(2, 0), CreateTime: time.Unix(0, 0), ModifyTime: time.Unix(0, 0)},
	})

	prevTime := now
	time.Sleep(time.Second * 2)

	//??????????????????
	userMod := User{
		UserId:    10001,
		Name:      "cat2",
		Age:       789,
		Money:     "91.23",
		LoginTime: time.Unix(3, 0),
	}
	db.MustExec("update t_user set ?.updateColumnValue where userId = ?", userMod, 2)

	db.MustQuery(&users, "select ?.column from t_user", users)

	for _, user := range users {
		checkTime(t, user.CreateTime, prevTime)
		checkNowTime(t, user.ModifyTime)
	}

	db.MustExec("update t_user set createTime = ?,modifyTime = ?", time.Unix(0, 0), time.Unix(0, 0))

	db.MustQuery(&users, "select ?.column from t_user", users)

	AssertEqual(t, users, []User{
		User{UserId: 2, Name: "cat2", Age: 789, Money: "91.23", LoginTime: time.Unix(3, 0), CreateTime: time.Unix(0, 0), ModifyTime: time.Unix(0, 0)},
	})

	//??????????????????
	userAdd := User{
		UserId:    10004,
		Name:      "bird",
		Age:       56,
		Money:     "33",
		LoginTime: time.Unix(4, 0),
	}
	//???????????????&???????????????????????????????????????????????????????????????????????????
	db.MustExec("insert into t_user(?.insertColumn) values ?.insertValue", &userAdd, &userAdd)

	db.MustExec("update t_user set createTime = ?,modifyTime = ?", time.Unix(0, 0), time.Unix(0, 0))

	db.MustQuery(&users, "select ?.column from t_user", users)

	AssertEqual(t, users, []User{
		User{UserId: 2, Name: "cat2", Age: 789, Money: "91.23", LoginTime: time.Unix(3, 0), CreateTime: time.Unix(0, 0), ModifyTime: time.Unix(0, 0)},
		User{UserId: 3, Name: "bird", Age: 56, Money: "33", LoginTime: time.Unix(4, 0), CreateTime: time.Unix(0, 0), ModifyTime: time.Unix(0, 0)},
	})

	//????????????????????????
	articles := []Article{}
	db.MustQuery(&articles, "select * from t_article", Article{})

	AssertEqual(t, articles, []Article{})

	//????????????
	articleAdds := []Article{
		Article{Data: []byte(`{"name":"fish","age":123}`), Remark: json.RawMessage(`{"name2":"fish","age2":123}`)},
		Article{Data: []byte(`{"name":"cat","age":789}`), Remark: json.RawMessage(`{"name2":"cat","age2":789}`)},
	}

	db.MustExec("insert into t_article(?.insertColumn) values ?.insertValue", articleAdds, articleAdds)

	db.MustExec("update t_article set createTime = ?,modifyTime = ?", time.Unix(0, 0), time.Unix(0, 0))

	db.MustQuery(&articles, "select ?.column from t_article", articles)

	AssertEqual(t, articles, []Article{
		Article{ArticleId: 1, Data: []byte(`{"name":"fish","age":123}`), Remark: json.RawMessage(`{"name2":"fish","age2":123}`), CreateTime: time.Unix(0, 0), ModifyTime: time.Unix(0, 0)},
		Article{ArticleId: 2, Data: []byte(`{"name":"cat","age":789}`), Remark: json.RawMessage(`{"name2":"cat","age2":789}`), CreateTime: time.Unix(0, 0), ModifyTime: time.Unix(0, 0)},
	})

}

func testStructTypeAll(t *testing.T, initDatabase func() SqlfDB) {
	db := initDatabase()
	testStructType(t, db)

	db2 := initDatabase().MustBegin()
	defer db2.MustClose()
	testStructType(t, db2)
	db2.MustCommit()

}

func testBuildInType(t *testing.T, db SqlfCommon) {
	users := []User{}

	//????????????type??????
	db.MustExec("insert into t_user(name,age,money,loginTime) values(?,?,?,?)", "fish", 123, Decimal("23"), time.Unix(1, 0))

	db.MustQuery(&users, "select ?.column from t_user", users)

	AssertEqual(t, users, []User{
		User{UserId: 1, Name: "fish", Age: 123, Money: "23", LoginTime: time.Unix(1, 0), CreateTime: time.Unix(0, 0), ModifyTime: time.Unix(0, 0)},
	})

	//??????[]type??????
	db.MustQuery(&users, "select * from t_user where name in (?) and age in (?) and money in (?) and loginTime in (?)",
		[]string{"12", "23"},
		[]int{1, 2, 3},
		[]Decimal{"123", "456", "789", "0ab"},
		[]time.Time{time.Unix(1, 0)},
	)
	AssertEqual(t, users, []User{})

	//??????*int???*[]int?????????
	db.MustExec("insert into t_user(name,age,money,loginTime) values(?,?,?,?)", "cat", 456, Decimal("78.13"), time.Unix(2, 0))

	var count int
	db.MustQuery(&count, "select count(*) from t_user")
	AssertEqual(t, count, 2)

	var userIds []int
	db.MustQuery(&userIds, "select userId from t_user")
	AssertEqual(t, userIds, []int{1, 2})

	//??????*string???*[]string?????????
	var name string
	db.MustQuery(&name, "select name from t_user where userId = 2")
	AssertEqual(t, name, "cat")

	var names []string
	db.MustQuery(&names, "select name from t_user")
	AssertEqual(t, names, []string{"fish", "cat"})

	//??????*Decimal???*[]Decimal?????????
	var money Decimal
	db.MustQuery(&money, "select money from t_user where userId = 2")
	AssertEqual(t, money, Decimal("78.13"))

	var moneys []Decimal
	db.MustQuery(&moneys, "select money from t_user")
	AssertEqual(t, moneys, []Decimal{"23", "78.13"})

	//??????*time.Time???*[]time.Time?????????
	var loginTime time.Time
	db.MustQuery(&loginTime, "select loginTime from t_user where userId = 2")
	AssertEqual(t, loginTime, time.Unix(2, 0))

	var loginTimes []time.Time
	db.MustQuery(&loginTimes, "select loginTime from t_user")
	AssertEqual(t, loginTimes, []time.Time{time.Unix(1, 0), time.Unix(2, 0)})

	//??????[]byte??????
	db.MustExec("insert into t_article(data,remark) values(?,?)", []byte(`{"name":"fish","age":123}`), json.RawMessage(`{"name2":"fish","age2":123}`))
	db.MustExec("insert into t_article(data,remark) values(?,?)", []byte(`{"name":"cat","age":456}`), json.RawMessage(`{"name2":"cat","age2":456}`))

	articles := []Article{}

	db.MustQuery(&articles, "select * from t_article where data in (?) and remark in (?)",
		[][]byte{[]byte(`{"name":"fish","age":123}`), []byte(`{"name":"cat","age":456}`)},
		[]json.RawMessage{json.RawMessage("123"), json.RawMessage("456")},
	)
	AssertEqual(t, articles, []Article{})

	var data []byte
	db.MustQuery(&data, "select data from t_article limit 0,1")
	AssertEqual(t, data, []byte(`{"name":"fish","age":123}`))

	var datas [][]byte
	db.MustQuery(&datas, "select data from t_article")
	AssertEqual(t, datas, [][]byte{[]byte(`{"name":"fish","age":123}`), []byte(`{"name":"cat","age":456}`)})

	var remark json.RawMessage
	db.MustQuery(&remark, "select remark from t_article limit 0,1")
	AssertEqual(t, remark, json.RawMessage(`{"name2":"fish","age2":123}`))

	var remarks []json.RawMessage
	db.MustQuery(&remarks, "select remark from t_article")
	AssertEqual(t, remarks, []json.RawMessage{json.RawMessage(`{"name2":"fish","age2":123}`), json.RawMessage(`{"name2":"cat","age2":456}`)})

}

func testBuildInTypeAll(t *testing.T, initDatabase func() SqlfDB) {
	db := initDatabase()
	testBuildInType(t, db)

	db2 := initDatabase().MustBegin()
	defer db2.MustClose()
	testBuildInType(t, db2)
	db2.MustCommit()
}

func testTxCommit(t *testing.T, initDatabase func() SqlfDB) {
	db := initDatabase()

	tx := db.MustBegin()

	//??????????????????
	userAdd := User{
		Name:      "bird",
		Age:       56,
		Money:     "33",
		LoginTime: time.Unix(4, 0),
	}
	tx.MustExec("insert into t_user(?.insertColumn) values ?.insertValue", userAdd, userAdd)

	tx.MustCommit()

	db.MustExec("update t_user set createTime = ?,modifyTime = ?", time.Unix(0, 0), time.Unix(0, 0))

	var users []User
	db.MustQuery(&users, "select ?.column from t_user", users)

	AssertEqual(t, users, []User{
		User{UserId: 1, Name: "bird", Age: 56, Money: "33", LoginTime: time.Unix(4, 0), CreateTime: time.Unix(0, 0), ModifyTime: time.Unix(0, 0)},
	})
}

func testTxRollBack(t *testing.T, initDatabase func() SqlfDB) {
	db := initDatabase()

	tx := db.MustBegin()

	//??????????????????
	userAdd := User{
		Name:      "bird",
		Age:       56,
		Money:     "33",
		LoginTime: time.Unix(4, 0),
	}
	tx.MustExec("insert into t_user(?.insertColumn) values ?.insertValue", userAdd, userAdd)

	tx.MustRollback()

	db.MustExec("update t_user set createTime = ?,modifyTime = ?", time.Unix(0, 0), time.Unix(0, 0))

	var users []User

	db.MustQuery(&users, "select ?.column from t_user", users)

	AssertEqual(t, users, []User{})
}

func testTxCloseCommit(t *testing.T, initDatabase func() SqlfDB) {
	db := initDatabase()

	tx := db.MustBegin()

	func() {
		defer tx.MustClose()

		//??????????????????
		userAdd := User{
			Name:      "bird",
			Age:       56,
			Money:     "33",
			LoginTime: time.Unix(4, 0),
		}
		tx.MustExec("insert into t_user(?.insertColumn) values ?.insertValue", userAdd, userAdd)

		tx.MustCommit()
	}()

	db.MustExec("update t_user set createTime = ?,modifyTime = ?", time.Unix(0, 0), time.Unix(0, 0))

	var users []User

	db.MustQuery(&users, "select ?.column from t_user", users)

	AssertEqual(t, users, []User{
		User{UserId: 1, Name: "bird", Age: 56, Money: "33", LoginTime: time.Unix(4, 0), CreateTime: time.Unix(0, 0), ModifyTime: time.Unix(0, 0)},
	})
}

func testTxCloseRollback(t *testing.T, initDatabase func() SqlfDB) {
	db := initDatabase()

	tx := db.MustBegin()

	func() {
		defer CatchCrash(func(e Exception) {

		})
		defer tx.MustClose()

		//??????????????????
		userAdd := User{
			Name:      "bird",
			Age:       56,
			Money:     "33",
			LoginTime: time.Unix(4, 0),
		}
		tx.MustExec("insert into t_user(?.insertColumn) values ?.insertValue", userAdd, userAdd)

		panic("ud")

		tx.MustCommit()
	}()

	db.MustExec("update t_user set createTime = ?,modifyTime = ?", time.Unix(0, 0), time.Unix(0, 0))

	var users []User

	db.MustQuery(&users, "select ?.column from t_user", users)

	AssertEqual(t, users, []User{})
}

type Item struct {
	ItemId      int `sqlf:"autoincr"`
	Name        string
	OnShelfTime time.Time
	CreateTime  time.Time
	ModifyTime  time.Time
}

func testZeroTime(t *testing.T, initDatabase func() SqlfDB) {
	//??????struct???time??????insert
	db := initDatabase()

	db.MustExec("insert into t_item(?.insertColumn) values ?.insertValue", Item{}, Item{
		Name:        "fish",
		OnShelfTime: ZERO_TIME,
		CreateTime:  ZERO_TIME,
		ModifyTime:  ZERO_TIME,
	})

	var items []Item
	db.MustQuery(&items, "select * from t_item")

	t.Logf("%v", items[0].OnShelfTime)
	AssertEqual(t, items[0].OnShelfTime == ZERO_TIME, true)

	db.MustExec("update t_item set ?.updateColumnValue", Item{
		Name:        "fish",
		OnShelfTime: ZERO_TIME,
		CreateTime:  ZERO_TIME,
		ModifyTime:  ZERO_TIME,
	})

	db.MustQuery(&items, "select * from t_item")

	AssertEqual(t, items[0].OnShelfTime == ZERO_TIME, true)

	//??????time??????
	db2 := initDatabase()

	db2.MustExec("insert into t_item(name,onShelfTime,createTime) values (?,?,?)", "cat", ZERO_TIME, &ZERO_TIME)

	var items2 []Item
	db2.MustQuery(&items2, "select * from t_item")

	AssertEqual(t, items2[0].OnShelfTime == ZERO_TIME, true)
	AssertEqual(t, items2[0].CreateTime == ZERO_TIME, true)
}

func testAll(t *testing.T, initDatabase func() SqlfDB) {
	testStructTypeAll(t, initDatabase)
	testBuildInTypeAll(t, initDatabase)
	testTxCommit(t, initDatabase)
	testTxRollBack(t, initDatabase)
	testTxCloseCommit(t, initDatabase)
	testTxCloseRollback(t, initDatabase)
	testZeroTime(t, initDatabase)
}

func TestAll(t *testing.T) {
	testAll(t, initSqliteDatabase)
	testAll(t, initMySqlDatabase)
}
