package web

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDbDatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Passowrd string
	Database string
}


func NewMongoDatabase(config MongoDbDatabaseConfig) (*mongo.Database, error) {
	if config.Host == "" {
		return nil, nil
	}
	dblink := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/",
		config.User,
		config.Passowrd,
		config.Host,
		config.Port,
	)
	client, err := mongo.NewClient(options.Client().ApplyURI(dblink))
	if err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}

	return client.Database(config.Database), nil
}

func NewMongoDatabaseFromConfig(configName string) (*mongo.Database, error) {
	config := MongoDbDatabaseConfig{}
	dataConfig := AppConfigInfoMongoDB{}
	switch configName {
	case "mdb":
		dataConfig = globalBasic.Config.Get().Mdb
	case "mdb2":
		dataConfig = globalBasic.Config.Get().Mdb2
	case "mdb3":
		dataConfig = globalBasic.Config.Get().Mdb3
	case "mdb4":
		dataConfig = globalBasic.Config.Get().Mdb4
	case "mdb5":
		dataConfig = globalBasic.Config.Get().Mdb5
	}

	config.Host = dataConfig.Host
	config.Port = dataConfig.Port
	config.User = dataConfig.User
	config.Passowrd = dataConfig.Password
	config.Database = dataConfig.Database

	return NewMongoDatabase(config)
}
