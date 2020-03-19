package core

import (
	"container/list"
	"context"
	"database/sql"
	"fmt"
	"github.com/cranewill/logcrane/def"
	"github.com/cranewill/logcrane/utils"
	"log"
	"sync/atomic"
	"time"
)

type Worker struct {
	Crane                 *LogCrane
	CurrentTable          string
	TableName             string
	CreateStatement       string
	SingleInsertStatement string
	BatchInsertStatement  string
	UpdateStatement       string
	LogCounter            *def.LogCounter
}

// NewWorker initializes a new worker
func NewWorker(crane *LogCrane, tableName string) *Worker {
	worker := &Worker{
		Crane:      crane,
		TableName:  tableName,
		LogCounter: &def.LogCounter{},
	}
	return worker
}

// doSingle deals one log recording
func (w *Worker) doSingle(cLog def.Logger, tableName string, rollType int32) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(tableName, ":")
			log.Println(err)
		}
	}()
	tableFullName := utils.GetTableFullNameByTableName(tableName, rollType)
	if w.CurrentTable == "" || w.CurrentTable != tableFullName {
		err := w.checkCreate(cLog, tableName, tableFullName, rollType)
		if err != nil {
			log.Println("Create table " + tableFullName + " error!")
			log.Println(err)
			return
		}
		insertStmt := utils.GetInsertSql(cLog) // do insert
		w.SingleInsertStatement = insertStmt
	}
	err := w.doSingleInsert(cLog, tableFullName, w.SingleInsertStatement)
	if err != nil {
		log.Println("Insert log " + tableFullName + " error!")
		log.Println(err)
		return
	}
	atomic.AddUint64(&w.LogCounter.Count, 1)
	atomic.AddUint64(&w.LogCounter.TotalCount, 1)
}

// doBatch deals a batch of logs
func (w *Worker) doBatch(logs *list.List, tableName string, rollType int32) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(tableName, ":")
			log.Println(err)
		}
	}()
	tableFullName := utils.GetTableFullNameByTableName(tableName, rollType)
	if (w.CurrentTable == "" || w.CurrentTable != tableFullName) && logs.Len() > 0 {
		err := w.checkCreate(logs.Front().Value.(def.Logger), tableName, tableFullName, rollType)
		if err != nil {
			log.Println("Create table " + tableFullName + " error!")
			log.Println(err)
			return
		}
		insertStmt := utils.GetBatchInsertSql(logs.Front().Value.(def.Logger))
		w.BatchInsertStatement = insertStmt
	}
	err := w.doBatchInsert(logs, tableFullName, w.BatchInsertStatement)
	if err != nil {
		log.Println("Insert log " + tableFullName + " error!")
		log.Println(err)
		return
	}
	atomic.AddUint64(&w.LogCounter.Count, uint64(logs.Len()))
	atomic.AddUint64(&w.LogCounter.TotalCount, uint64(logs.Len()))
}

// doUpdate updates logs if they exist in db, insert a new log otherwise
func (w *Worker) doUpdate(logs *list.List, tableName string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(tableName, ":")
			log.Println(err)
		}
	}()
	tableFullName := utils.GetTableFullNameByTableName(tableName, def.Never)
	if (w.CurrentTable == "" || w.CurrentTable != tableFullName) && logs.Len() > 0 {
		err := w.checkCreate(logs.Front().Value.(def.Logger), tableName, tableFullName, def.Never)
		if err != nil {
			log.Println("Create table " + tableFullName + " error!")
			log.Println(err)
			return
		}
		stmt := utils.GetUpdateSql(logs)
		w.UpdateStatement = stmt
	}
	err := w.doUpdateInsert(logs, tableFullName, w.UpdateStatement)
	if err != nil {
		log.Println("Update-Insert log " + tableFullName + " error!")
		log.Println(err)
		return
	}
	atomic.AddUint64(&w.LogCounter.Count, uint64(logs.Len()))
	atomic.AddUint64(&w.LogCounter.TotalCount, uint64(logs.Len()))
}

// checkCreate creates the table
func (w *Worker) checkCreate(cLog def.Logger, tableName, tableFullName string, rollType int32) error {
	var s string
	err := w.Crane.MysqlDb.QueryRow("SHOW TABLES LIKE '" + tableFullName + "';").Scan(&s)
	if err != nil {
		if err == sql.ErrNoRows { // table not exist in db
			var createStmt string
			if w.CreateStatement == "" {
				createStmt = utils.GetNewCreateSql(cLog)
				w.CreateStatement = createStmt
			}
			stmt := fmt.Sprintf(w.CreateStatement, tableFullName)
			_, err := w.Crane.MysqlDb.Exec(stmt)
			log.Println("Create table ", tableFullName)
			if err != nil {
				log.Println(stmt)
				return err
			}
		} else {
			return err
		}
	}
	w.CurrentTable = tableFullName
	return nil
}

// doSingleInsert inserts a single cLog
func (w *Worker) doSingleInsert(cLog def.Logger, tableFullName, insertStmt string) error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	values := utils.GetInsertValues(cLog)
	preparedStmt := insertStmt + "(" + values + ");"
	stmt := fmt.Sprintf(preparedStmt, tableFullName)
	_, err := w.Crane.MysqlDb.ExecContext(ctx, stmt)
	if err != nil {
		log.Println(stmt)
		return err
	}
	return nil
}

// doBatchInsert inserts numbers of logs at one time
func (w *Worker) doBatchInsert(logs *list.List, tableFullName, insertStmt string) error {
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
	_, err := w.Crane.MysqlDb.ExecContext(ctx, insertStmt)
	if err != nil {
		log.Println(insertStmt)
		return err
	}
	return nil
}

// doUpdateInsert executes update
func (w *Worker) doUpdateInsert(logs *list.List, tableFullName, updateStmt string) error {
	if logs.Len() == 0 {
		return nil
	}
	updateStmt = fmt.Sprintf(updateStmt, tableFullName) + ";"
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err := w.Crane.MysqlDb.ExecContext(ctx, updateStmt)
	if err != nil {
		log.Println(updateStmt)
		return err
	}
	return nil
}
