# LogCrane 一个游戏日志打点解决方案 by golang

## 该系统中的日志分为三种：

* 玩家日志（PlayerBaseLog）：

针对记录玩家的个人行为

* 服务器日志（ServerBaseLog）：

针对记录服务器的事件

**建议所有增量日志都应该继承自这两者之一**


* 自定义插入-更新日志：

使用于部分特殊日志，比如玩家基本信息表这种不是增量而是"无则插入，有则更新"的表，由使用者自由设计字段

## 日志定义方案：

直接手动定义每个日志的结构，对应sql结构写在tag里面，tag定义按照如下规则：
* type：字段对应日志表中列的类型；
* length：字段对应日志表中列的长度，一般varchar类定义，如果没定义默认`255`。其他类型都不用定义，使用mysql的默认长度；
* explain：字段的注释，现在这个字段并不会体现在mysql表中；
* name：字段对应日志表中列名。如果没有指定，则默认是该字段名小写；
* key: 字段是否为主键或索引，值为"primary"时为主键，为其他时为普通索引。一个字段可以对应多个key，key的名字以","隔开。含有相同key的多个字段会作为联合索引。

**注意当日志里有字段"pk_id"时，"primary"不会生效**

例如基本玩家日志结构定义：

```go
type BasePlayerLog struct {
	PkId       string `type:"int" explain:"自增主键" name:"pk_id"`
	PlayerId   string `type:"varchar" length:"255" explain:"玩家id" name:"player_id" key:"player_id,player_server_id"`
	ServerId   string `type:"varchar" length:"255" explain:"服务器id" name:"server_id" key:"player_server_id"`
	CreateTime int64  `type:"bigint" explain:"创建时间" name:"create_time"`
	SaveTime   int64  `type:"bigint" explain:"保存时间" name:"save_time"`
	ActionId   string `type:"varchar" length:"255" explain:"行为id" name:"action_id"`
}

```
继承自基本玩家日志的玩家登录日志定义：

```go
type OnlineLog struct {
	Base   def.BasePlayerLog
	Source string `type:"varchar" length:"255" explain:"来源" key:"source"`
	Ip     string `type:"varchar" length:"255" explain:"IP"`
}
```

所有日志都要实现Logger接口，需要实现一下三个方法：

```go
func (log OnlineLog) TableName() string { // 日志对应的数据库中table的名字
	return "log_online"
}

func (log OnlineLog) RollType() int32 { // 日志表分表规则
	return def.RollTypeDay
}

func (log OnlineLog) SaveType() int32 { // 保存类型，分为单条插入、批量插入和批量插入-更新
	return def.Single
}
```

## 调用方法：

```go
crane.Start(ServerId, "username", "password", "log_db", monitor_tick) // 启动日志系统，指定日志库和监控日志打印频率，以秒为单位
logger := logs.NewOnlineLog("playerId", "source", "127.0.0.1", "") // 创建日志对象
crane.Instance().Execute(logger) // 执行日志记录
```

## 停止系统：

```go
crane.Stop()
```

## Todo：

* 优化Test文件内容
* 支持更多mysql表属性定义
* 支持更多mysql表操作，例如修改表结构
* 支持游戏后台统一管理
* 支持更多种类数据库，例如mongodb