// The crane package provides the api to use this log system
package crane

import (
	"database/sql"
	"fmt"
	"github.com/cranewill/logcrane/core"
	"github.com/cranewill/logcrane/def"
	"log"
)

var crane *core.LogCrane

// Instance returns the singleton instance of LogCrane
func Instance() *core.LogCrane {
	if !crane.Init {
		log.Println("Log service not started!")
		return nil
	}
	return crane
}

// Start starts LogCrane
func Start(serverId, user, pwd, db string) {
	crane = &core.LogCrane{
		LogChannels:            make(map[string]chan def.Logger),
		SingleInsertStatements: make(map[string]string),
		BatchInsertStatements:  make(map[string]string),
		CreateStatements:       make(map[string]string),
		ServerId:               serverId,
		Init:                   false,
	}
	switch def.DataBase {
	case def.MySql:
		driver := "%s:%s@/%s"
		db, err := sql.Open("mysql", fmt.Sprintf(driver, user, pwd, db))
		if err != nil {
			panic(err.Error())
		}
		// test db handle
		err = db.Ping()
		if err != nil {
			panic(err.Error())
		}
		crane.MysqlDb = db
	case def.Mongo:
		// todo ... init mongo
	}
	crane.Init = true
	def.ServerId = serverId
	def.BatchNum = 100
	go crane.Lift()
	log.Println("Log System Started!")
}

// Stop stops the log system
func Stop() {
	_ = crane.MysqlDb.Close()
}
