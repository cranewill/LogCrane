package logs

import (
	"github.com/cranewill/logcrane/def"
)

type PlayerInfo struct {
	//Base             def.BasePlayerLog
	PlayerId         string `type:"varchar" length:"255" explain:"玩家id" name:"player_id"`
	ServerId         string `type:"varchar" length:"255" explain:"服务器id" name:"server_id"`
	Level            int32  `type:"int" explain:"等级" name:"level"`
	Location         string `type:"varchar" length:"255" explain:"地区" name:"location"` // varchar(255)	地区
	Language         string `type:"varchar" length:"255" explain:"语言" name:"language"` // varchar(255)	语言
	Ip               string `type:"varchar" length:"255" explain:"ip" name:"ip"`       // varchar(255)	ip
	System           string `type:"varchar" length:"255" explain:"系统" name:"system"`   // varchar(255)	系统
	Device           string `type:"varchar" length:"255" explain:"设备" name:"device"`   // varchar(255)	设备
	Source           string `type:"varchar" length:"255" explain:"来源" name:"source"`   // varchar(255)	来源
	PlayerCreateTime int64  `type:"bigint" explain:"玩家创建时间" name:"player_create_time"` // bigint	玩家创建时间
}

func (login PlayerInfo) TableName() string {
	return "player_info"
}

func (login PlayerInfo) RollType() int32 {
	return def.Never
}

func (login PlayerInfo) SaveType() int32 {
	return def.Update
}

func NewPlayerInfo(playerId, serverId, location, language string, level int32, playerCreateTime int64) PlayerInfo {
	return PlayerInfo{
		PlayerId:         playerId,
		ServerId:         serverId,
		Level:            level,
		Location:         location,
		Language:         language,
		PlayerCreateTime: playerCreateTime,
	}
}
