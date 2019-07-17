package core

import (
	"github.com/cranewill/logcrane/utils"
)

var crane *LogCrane

// Start starts LogCrane
func Start() error {
	crane = &LogCrane{
		LogPools: make(map[string]chan *ILog),
	}
	switch DataBase {
	case MySql:
		// todo ... 初始化MySql
	case Mongo:
		// todo ... 初始化Mongo
	}

	return nil
}

// Instance returns the singleton instance of LogCrane
func Instance() *LogCrane {
	return crane
}

// Execute throws the log to its own channel waiting for saving
func (crane *LogCrane) Execute(log *ILog) {
	logName := utils.GetLogName(log)
	if _, exist := crane.LogPools[logName]; !exist {
		crane.LogPools[logName] = make(chan *ILog)
		// Create a new dealing Goroutine of this log
		go crane.Fly(crane.LogPools[logName])
	}
	logChan, _ := crane.LogPools[logName]
	logChan <- log
}

// Fly accepts a log channel and deals the recording task of this log
func (crane *LogCrane) Fly(logChan chan *ILog) {
	// todo ... 考虑是每次从logChan取一个存到一个List中一起处理还是指定时间从logChan中取一次一起处理
	//for {
	//log := <-logChan
	//
	//}
}
