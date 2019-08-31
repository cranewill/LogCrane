// The logs package defines the logs
package logs

import (
	"github.com/cranewill/logcrane/def"
	"time"
)

type OnlineLog struct {
	Base   def.BasePlayerLog
	Source string `type:"varchar" length:"255" explain:"来源"`
	Ip     string `type:"varchar" length:"255" explain:"IP"`
}

// NewOnlineLog constructs a new OnlineLog
func NewOnlineLog(playerId, source, ip string) OnlineLog {
	return OnlineLog{
		Base: def.BasePlayerLog{
			PlayerId:   playerId,
			CreateTime: time.Now().Unix(),
		},
		Source: source,
		Ip:     ip,
	}
}

func (log OnlineLog) TableName() string {
	return "log_online"
}

func (log OnlineLog) RollType() int32 {
	return def.RollTypeDay
}

func (log OnlineLog) SaveType() int32 {
	//return def.Single
	return def.Batch
}
