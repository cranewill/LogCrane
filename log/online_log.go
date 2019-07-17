package log

import (
	. "github.com/cranewill/logcrane/core"
	"time"
)

type OnlineLog struct {
	base   *BasePlayerLog
	source string
	ip     string
}

func NewOnlineLog(playerId, serverId, source, ip string) *OnlineLog {
	log := &OnlineLog{}
	log.base = &BasePlayerLog{
		Id:         11,
		PlayerId:   playerId,
		ServerId:   serverId,
		RollType:   1,
		SaveMethod: 1,
		CreateTime: time.Now().Unix(),
	}
	log.source = source
	log.ip = ip
	return log
}

func (log OnlineLog) ToString() {

}

func (log OnlineLog) GetRollType() int32 {
	return log.base.RollType
}
