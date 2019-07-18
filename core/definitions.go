package core

import ()

// Log database type
const (
	MySql = 1
	Mongo = 2
)

// Log record type
const (
	Single = 1
	Batch  = 2
)

// Log table split type
const (
	RollTypeDay   = 1
	RollTypeMonth = 2
	RollTypeYear  = 3
)

const DataBase = MySql

type ILog interface {
	ToString() string
	GetRollType() int32
}

type BasePlayerLog struct {
	Id         int64  `type:"bigint" explain:"pk_id"`
	PlayerId   string `type:"varchar" length:"255" explain:"玩家id"`
	ServerId   string `type:"varchar" length:"255" explain:"服务器id"`
	RollType   int32
	SaveMethod int32
	CreateTime int64
	SaveTime   int64
}

type LogCrane struct {
	// todo ... mysql 连接实例
	// todo ... mongo 连接实例

	LogPools map[string]chan *ILog
}
