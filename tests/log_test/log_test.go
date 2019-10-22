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

	oLog := logs.NewOnlineLog("TestPlayerId", "ss", "127.0.0.1", "sdfsdfsd")
	for i := 0; i < 1000; i++ {
		crane.Instance().Execute(oLog)
	}
	for {
		time.Sleep(time.Second)
	}
}

func TestUpdate(t *testing.T) {
	crane.Start(ServerId, "root", "starunion", "test", 5)

	for i := 0; i < 1000; i++ {
		id := rand.Int31n(1000)
		randStr := strconv.Itoa(int(id))
		pLog := logs.NewPlayerInfo(randStr, "server"+randStr, "location"+randStr, "1"+randStr, id, time.Now().Unix())

		crane.Instance().Execute(pLog)
	}
	for {
		time.Sleep(time.Second * 3)
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
