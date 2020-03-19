// The core package contains all the main logic code
package core

import (
	"container/list"
	"database/sql"
	"github.com/cranewill/logcrane/def"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var craneChan chan def.Logger

func init() {
	def.ChannelBuffer = 10000
	def.BatchNum = 100
	craneChan = make(chan def.Logger, def.ChannelBuffer)
}

type LogCrane struct {
	MysqlDb     *sql.DB                    // the mysql database handle
	Running     bool                       // is running
	ServerId    string                     // server id
	LogChannels map[string]chan def.Logger // tableName -> channel. every channel deal one type of cLog
	Workers     map[string]*Worker         // tableName -> worker
	Wgp         *sync.WaitGroup
}

// Execute throws the logs and put them into a channel to avoid from concurrent panic
func (c *LogCrane) Execute(cLog def.Logger) {
	if c == nil {
		log.Println("Log system not init!")
		return
	}
	if !c.Running {
		log.Println("Log system not running!")
		return
	}
	craneChan <- cLog
}

// Lift gives every cLog to its own channel waiting for saving
func (c *LogCrane) Lift() {
	for {
		if !c.Running {
			return
		}
		cLog := <-craneChan
		tableName := cLog.TableName()
		if _, exist := c.LogChannels[tableName]; !exist {
			c.LogChannels[tableName] = make(chan def.Logger, def.ChannelBuffer)
			c.Workers[tableName] = NewWorker(c, tableName)
			task := func() {
				c.Fly(c.Wgp, c.LogChannels[tableName], tableName, cLog.RollType(), cLog.SaveType())
			}
			c.Wgp.Add(1)
			go task()
		}
		if !c.Running {
			return
		}
		logChan, _ := c.LogChannels[tableName]
		logChan <- cLog
	}
}

// Fly accepts a logs channel and deals the recording tasks of this logs according to the save type
func (c *LogCrane) Fly(wgp *sync.WaitGroup, logChan chan def.Logger, tableName string, rollType, saveType int32) {
	defer wgp.Done()
	queue := list.New()
	worker, exist := c.Workers[tableName]
	if !exist {
		log.Println("Get worker [", tableName, "] failed!")
		return
	}
	for {
		if !c.Running {
			log.Println("Stop log worker ", tableName)
			break
		}
		switch saveType {
		case def.Single:
			cLog := <-logChan
			worker.doSingle(cLog, tableName, rollType)
		case def.Batch:
			select {
			case clog := <-logChan:
				queue.PushBack(clog)
				if queue.Len() >= def.BatchNum {
					worker.doBatch(queue, tableName, rollType)
					queue.Init()
				}
			case <-time.After(5 * time.Second):
				worker.doBatch(queue, tableName, rollType)
				queue.Init()
			}
		case def.Update:
			select {
			case clog := <-logChan:
				queue.PushBack(clog)
				if queue.Len() >= def.BatchNum {
					worker.doUpdate(queue, tableName)
					queue.Init()
				}
			case <-time.After(5 * time.Second):
				worker.doUpdate(queue, tableName)
				queue.Init()
			}
		}
	}
}

// Monitor creates a time ticker with  duration, and prints the monitor log
// of the log system every tick
func (c *LogCrane) Monitor(duration time.Duration) {
	t := time.NewTicker(duration)
	for range t.C {
		for tableName, worker := range c.Workers {
			counter := worker.LogCounter
			tCount := &counter.TotalCount
			count := &counter.Count
			log.Println(tableName + ": New " + strconv.Itoa(int(atomic.LoadUint64(count))) + ", Total " + strconv.Itoa(int(atomic.LoadUint64(tCount))))
			counter.Count = 0
		}
	}
}

// Stop ends all the goroutine and finish all the logs left,
// use batch insert to finish the logs
func (c *LogCrane) Stop() {
	c.Running = false
	c.Wgp.Wait() // wait for the end of every worker goroutine
	for tableName, logChan := range c.LogChannels {
		size := len(logChan)
		if size <= 0 {
			continue
		}
		unFinished := list.New()
		for i := 0; i < size; i++ {
			unFinished.PushBack(<-logChan)
		}
		cLog := unFinished.Front().Value.(def.Logger)
		rollType := cLog.RollType()
		worker, exist := c.Workers[tableName]
		if !exist {
			log.Println("Get worker [", tableName, "] failed!")
			continue
		}
		log.Println("Clean ", size, " logs ", tableName, " when system stop ...")
		worker.doBatch(unFinished, tableName, rollType)
	}
}
