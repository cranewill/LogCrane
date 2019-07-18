package utils

import (
	"github.com/cranewill/logcrane/core"
	"reflect"
	"time"
)

// GetLogName returns the log type name
func GetLogName(log *core.ILog) string {
	return reflect.TypeOf(log).Name()
}

// GetTableName returns the DB table name of which the log is to be inserted,
// and the name differs due to the RollType of this log
func GetTableName(log *core.ILog, rollType int32) string {
	logType := GetLogName(log)
	year, month, day := time.Now().Date()
	var timeStr string
	switch rollType {
	case core.RollTypeDay:
		timeStr = string(year*10000 + int(month)*100 + day)
	case core.RollTypeMonth:
		timeStr = string(year*100 + int(month))
	case core.RollTypeYear:
		timeStr = string(year)
	}
	return logType + "_" + timeStr
}
