package log_test

import (
	"github.com/cranewill/logcrane/crane"
	"github.com/cranewill/logcrane/logs"
	"strconv"
	"testing"
	"time"
)

const ServerId = "TestServer"

func TestCraneLog(t *testing.T) {
	crane.Start(ServerId, "root", "starunion", "test")

	for i := 0; i < 100; i++ {
		go doLog(i)
	}
	//for i := 0; i < 10; i++ {
	//	oLog := logs.NewOnlineLog("TestPlayerId", strconv.Itoa(i), "127.0.0.1")
	//	crane.Instance().Execute(oLog)
	//}
	for {
		time.Sleep(time.Second)
	}
}

func doLog(j int) {
	for i := 0; i < 5000; i++ {
		oLog := logs.NewOnlineLog("TestPlayerId", strconv.Itoa(j), "127.0.0.1")
		crane.Instance().Execute(oLog)
	}
}
