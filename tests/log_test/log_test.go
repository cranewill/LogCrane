package log_test

import (
	"github.com/cranewill/logcrane/crane"
	"github.com/cranewill/logcrane/logs"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

const ServerId = "TestServer"

func TestCraneLog(t *testing.T) {
	crane.Start(ServerId, "root", "starunion", "test", 5)

	//for i := 0; i < 10; i++ {
	//	go doLog(i)
	//}
	oLog := logs.NewOnlineLog("TestPlayerId", "ss", "127.0.0.1", "sdfsdfsd")
	//crane.Instance().Execute(oLog)
	for i := 0; i < 1000; i++ {
		//oLog := logs.NewOnlineLog("TestPlayerId", strconv.Itoa(i), "127.0.0.1")
		crane.Instance().Execute(oLog)
	}
	for {
		time.Sleep(time.Second)
	}
}

func doLog(j int) {
	for {
		oLog := logs.NewOnlineLog("TestPlayerId", strconv.Itoa(j), "127.0.0.1", "test_action_id")
		crane.Instance().Execute(oLog)
		sc := rand.Int63n(3)
		time.Sleep(time.Duration(sc) * time.Second)
	}
}
