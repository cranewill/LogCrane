// The core package contains all the main logic code
package core

import (
	"container/list"
	"context"
	"database/sql"
	"fmt"
	"github.com/cranewill/logcrane/def"
	"github.com/cranewill/logcrane/utils"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"sync/atomic"
	"time"
)

var craneChan chan def.Logger

func init() {
	craneChan = make(chan def.Logger, 1024)
}

type LogCrane struct {
	MysqlDb                *sql.DB                    // the mysql database handle
	Running                bool                       // is running
	ServerId               string                     // server id
	CreateStatements       map[string]string          // tableName -> createSql
	SingleInsertStatements map[string]string          // tableName -> insertSql
	BatchInsertStatements  map[string]string          // tableName -> insertSql
	UpdateStatements       map[string]string          // table -> updateSql
	ExistTables            map[string]string          // tableName _. tableFullName
	LogCounters            map[string]*def.LogCounter // tableName -> logCounter
	LogChannels            map[string]chan def.Logger // tableName -> channel. every channel deal one type of cLog
}

// Execute throws the logs and put them into a channel to avoid from concurrent panic
func (c *LogCrane) Execute(cLog def.Logger) {
	if c == nil {
		log.Println("Log system not init!")
		return
	}
	if !c.Running {
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
			c.LogChannels[tableName] = make(chan def.Logger, 1024)
			c.LogCounters[tableName] = &def.LogCounter{TotalCount: 0, Count: 0}
			go c.Fly(c.LogChannels[tableName], tableName, cLog.RollType(), cLog.SaveType())
		}
		if !c.Running {
			return
		}
		logChan, _ := c.LogChannels[tableName]
		logChan <- cLog
	}
}

// Fly accepts a logs channel and deals the recording tasks of this logs according to the save type
func (c *LogCrane) Fly(logChan chan def.Logger, tableName string, rollType, saveType int32) {
	queue := list.New()
	for {
		switch saveType {
		case def.Single:
			cLog := <-logChan
			c.doSingle(cLog, tableName, rollType)
		case def.Batch:
			select {
			case clog := <-logChan:
				queue.PushBack(clog)
				if queue.Len() >= def.BatchNum {
					c.doBatch(queue, tableName, rollType)
					queue.Init()
				}
			case <-time.After(5 * time.Second):
				c.doBatch(queue, tableName, rollType)
				queue.Init()
			}
		case def.Update:
			select {
			case clog := <-logChan:
				queue.PushBack(clog)
				if queue.Len() >= def.BatchNum {
					c.doUpdate(queue, tableName)
					queue.Init()
				}
			case <-time.After(5 * time.Second):
				c.doUpdate(queue, tableName)
				queue.Init()
			}
		}
	}
}

// doSingle deals one log recording
func (c *LogCrane) doSingle(cLog def.Logger, tableName string, rollType int32) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(tableName, ":")
			log.Println(err)
		}
	}()
	tableFullName := utils.GetTableFullNameByTableName(tableName, rollType)
	count := c.LogCounters[tableName]
	existTableName, exist := c.ExistTables[tableName]
	if !exist || existTableName != tableFullName { // not exist in memory, check if this table created
		err := c.checkCreate(cLog, tableName, tableFullName, rollType)
		if err != nil {
			log.Println("Create table " + tableFullName + " error!")
			log.Println(err)
			return
		}
		insertStmt := utils.GetInsertSql(cLog) // do insert
		c.SingleInsertStatements[tableName] = insertStmt
	}
	err := c.doSingleInsert(cLog, tableFullName, c.SingleInsertStatements[tableName])
	if err != nil {
		log.Println("Insert log " + tableFullName + " error!")
		log.Println(err)
		return
	}
	atomic.AddUint64(&count.Count, 1)
	atomic.AddUint64(&count.TotalCount, 1)
}

// doBatch deals a batch of logs
func (c *LogCrane) doBatch(logs *list.List, tableName string, rollType int32) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(tableName, ":")
			log.Println(err)
		}
	}()
	tableFullName := utils.GetTableFullNameByTableName(tableName, rollType)
	count := c.LogCounters[tableName]
	existTableName, exist := c.ExistTables[tableName]
	if (!exist || existTableName != tableFullName) && logs.Len() > 0 {
		err := c.checkCreate(logs.Front().Value.(def.Logger), tableName, tableFullName, rollType)
		if err != nil {
			log.Println("Create table " + tableFullName + " error!")
			log.Println(err)
			return
		}
		insertStmt := utils.GetBatchInsertSql(logs.Front().Value.(def.Logger))
		c.BatchInsertStatements[tableName] = insertStmt
	}
	err := c.doBatchInsert(logs, tableFullName, c.BatchInsertStatements[tableName])
	if err != nil {
		log.Println("Insert log " + tableFullName + " error!")
		log.Println(err)
		return
	}
	atomic.AddUint64(&count.Count, uint64(logs.Len()))
	atomic.AddUint64(&count.TotalCount, uint64(logs.Len()))
}

// doUpdate updates logs if they exist in db, insert a new log otherwise
func (c *LogCrane) doUpdate(logs *list.List, tableName string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(tableName, ":")
			log.Println(err)
		}
	}()
	tableFullName := utils.GetTableFullNameByTableName(tableName, def.Never)
	count := c.LogCounters[tableName]
	existTableName, exist := c.ExistTables[tableName]
	if (!exist || existTableName != tableFullName) && logs.Len() > 0 {
		err := c.checkCreate(logs.Front().Value.(def.Logger), tableName, tableFullName, def.Never)
		if err != nil {
			log.Println("Create table " + tableFullName + " error!")
			log.Println(err)
			return
		}
		stmt := utils.GetUpdateSql(logs)
		c.UpdateStatements[tableName] = stmt
	}
	err := c.doUpdateInsert(logs, tableFullName, c.UpdateStatements[tableName])
	if err != nil {
		log.Println("Update-Insert log " + tableFullName + " error!")
		log.Println(err)
		return
	}
	atomic.AddUint64(&count.Count, uint64(logs.Len()))
	atomic.AddUint64(&count.TotalCount, uint64(logs.Len()))
}

// checkCreate creates the table if the table doesn't exist in db
func (c *LogCrane) checkCreate(cLog def.Logger, tableName, tableFullName string, rollType int32) error {
	var result string
	err := c.MysqlDb.QueryRow("SHOW TABLES LIKE '" + tableFullName + "';").Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows { // table not exist in db
			createStmt, exist := c.CreateStatements[tableName]
			if !exist {
				if rollType == def.Never {
					createStmt = utils.GetPlayerIdPKCreateSql(cLog)
					c.CreateStatements[tableName] = createStmt
				} else {
					createStmt = utils.GetCreateSql(cLog)
					c.CreateStatements[tableName] = createStmt
				}
			}
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			stmt := fmt.Sprintf(createStmt, tableFullName)
			_, err := c.MysqlDb.ExecContext(ctx, stmt)
			log.Println("Create table ", tableFullName)
			if err != nil {
				log.Println(stmt)
				return err
			}
		} else {
			return err
		}
	}
	c.ExistTables[tableName] = tableFullName
	return nil
}

// doSingleInsert inserts a single cLog
func (c *LogCrane) doSingleInsert(cLog def.Logger, tableFullName, insertStmt string) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	values := utils.GetInsertValues(cLog)
	preparedStmt := insertStmt + "(" + values + ");"
	stmt := fmt.Sprintf(preparedStmt, tableFullName)
	_, err := c.MysqlDb.ExecContext(ctx, stmt)
	if err != nil {
		log.Println(stmt)
		return err
	}
	return nil
}

// doBatchInsert inserts numbers of logs at one time
func (c *LogCrane) doBatchInsert(logs *list.List, tableFullName, insertStmt string) error {
	if logs.Len() == 0 {
		return nil
	}
	for cLog := logs.Front(); cLog != nil; cLog = cLog.Next() {
		sep := ","
		if cLog.Next() == nil {
			sep = ""
		}
		insertStmt += "(" + utils.GetInsertValues(cLog.Value.(def.Logger)) + ")" + sep
	}
	insertStmt = fmt.Sprintf(insertStmt, tableFullName) + ";"
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err := c.MysqlDb.ExecContext(ctx, insertStmt)
	if err != nil {
		log.Println(insertStmt)
		return err
	}
	return nil
}

// doUpdateInsert executes update
func (c *LogCrane) doUpdateInsert(logs *list.List, tableFullName, updateStmt string) error {
	if logs.Len() == 0 {
		return nil
	}
	updateStmt = fmt.Sprintf(updateStmt, tableFullName) + ";"
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err := c.MysqlDb.ExecContext(ctx, updateStmt)
	if err != nil {
		log.Println(updateStmt)
		return err
	}
	return nil
}

// Monitor creates a time ticker with  duration, and prints the monitor log
// of the log system every tick
func (c *LogCrane) Monitor(duration time.Duration) {
	t := time.NewTicker(duration)
	for range t.C {
		for tableName, counter := range c.LogCounters {
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
	//defer c.MysqlDb.Close()
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
		tableFullName := utils.GetTableFullNameByTableName(tableName, rollType)
		existTableName, exist := c.ExistTables[tableName]
		if !exist && existTableName != tableFullName && size > 0 {
			err := c.checkCreate(cLog, tableName, tableFullName, rollType)
			if err != nil {
				log.Println("Create table " + tableFullName + " error!")
				continue
			}
			insertStmt := utils.GetBatchInsertSql(cLog)
			c.BatchInsertStatements[tableName] = insertStmt
		}
		err := c.doBatchInsert(unFinished, tableFullName, c.BatchInsertStatements[tableName])
		if err != nil {
			log.Println("Insert cLog " + tableFullName + " error!")
			break
		}
	}
}
