// The crane package provides the api to use this log system
package crane

import (
	"database/sql"
	"fmt"
	"github.com/cranewill/logcrane/core"
	"github.com/cranewill/logcrane/def"
	"log"
	"sync"
	"time"
)

var crane *core.LogCrane

// Instance returns the singleton instance of LogCrane
func Instance() *core.LogCrane {
	if !crane.Running {
		log.Println("Log service not started!")
		return nil
	}
	return crane
}

// Start starts LogCrane.
// if monitorTick > 0, a log monitor will be started and it prints monitor log every tick(second)
func Start(serverId, user, pwd, db string, monitorTick int32) {
	if crane != nil {
		return
	}
	crane = &core.LogCrane{
		LogChannels: make(map[string]chan def.Logger),
		ServerId:    serverId,
		Running:     false,
		Workers:     make(map[string]*core.Worker),
		Wgp:         &sync.WaitGroup{},
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
	crane.Running = true
	def.ServerId = serverId
	def.BatchNum = 100
	go crane.Lift()
	if monitorTick > 0 {
		go crane.Monitor(time.Duration(monitorTick) * time.Second)
	}
	log.Println("Log System Started!")
}

// Stop stops the log system
func Stop() {
	log.Println("Stop Log System ...")
	crane.Stop()
	log.Println("Log System Stopped!")
}
