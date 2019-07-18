package log

import (
	. "github.com/cranewill/logcrane/core"
	"time"
)

type OnlineLog struct {
	Base   *BasePlayerLog
	Source string
	Ip     string
}

func NewOnlineLog(playerId, serverId, source, ip string) *OnlineLog {
	log := &OnlineLog{}
	log.Base = &BasePlayerLog{
		Id:         11,
		PlayerId:   playerId,
		ServerId:   serverId,
		RollType:   1,
		SaveMethod: 1,
		CreateTime: time.Now().Unix(),
	}
	log.Source = source
	log.Ip = ip
	return log
}

func (log OnlineLog) ToString() {

}

func (log OnlineLog) GetRollType() int32 {
	return log.Base.RollType
}
