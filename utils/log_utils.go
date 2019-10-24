// The utils package provides all the tool functions of log dealing
package utils

import (
	"container/list"
	"fmt"
	"github.com/cranewill/logcrane/def"
	log2 "log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var LogFieldsDefinitions map[string][]def.ColumnDef

func init() {
	LogFieldsDefinitions = make(map[string][]def.ColumnDef)
}

// GetTableFullNameByTableName returns the DB table name of the specific log name
func GetTableFullNameByTableName(tableName string, rollType int32) string {
	year, month, day := time.Now().Date()
	var timeStr string
	switch rollType {
	case def.Never:
		return tableName
	case def.RollTypeDay:
		timeStr = strconv.Itoa(year*10000 + int(month)*100 + day)
	case def.RollTypeMonth:
		timeStr = strconv.Itoa(year*100 + int(month))
	case def.RollTypeYear:
		timeStr = strconv.Itoa(year)
	}
	return tableName + "_" + timeStr
}

// GetTableFullName returns the DB table name of the specific log name
func GetTableFullName(log def.Logger, rollType int32) string {
	return GetTableFullNameByTableName(log.TableName(), rollType)
}

// GetFields returns a slice contains logs's every attributes table column def from memory.
// If this is called to construct the field part of sql statement, it returns from memory.
// Otherwise get attributes from log itself.
func GetFields(log interface{}, onlyDef bool) []def.ColumnDef {
	typ := reflect.TypeOf(log)
	logName := typ.Name()
	if onlyDef {
		_, exist := LogFieldsDefinitions[logName]
		if !exist {
			LogFieldsDefinitions[logName] = GetFieldDefs(log, true)
		}
		return LogFieldsDefinitions[logName]
	} else {
		return GetFieldDefs(log, false)
	}
}

// GetFieldDefs returns the all the logs's attributes by reflection
func GetFieldDefs(log interface{}, onlyDef bool) []def.ColumnDef {
	val := reflect.ValueOf(log)
	typ := reflect.TypeOf(log)
	fields := make([]def.ColumnDef, 0)
	for i := 0; i < val.NumField(); i++ {
		fVal := val.Field(i)
		fTyp := typ.Field(i)
		if fVal.Kind() == reflect.Struct {
			fields = append(fields, GetFields(fVal.Interface(), onlyDef)...)
		} else {
			field := def.ColumnDef{}
			tag := fTyp.Tag
			if value, ok := tag.Lookup("name"); ok { // field name
				field.Name = value
			} else {
				field.Name = strings.ToLower(fTyp.Name)
			}
			if value, ok := tag.Lookup("type"); ok { // field type
				field.Type = value
			} else {
				log2.Println("Error def of " + typ.Name() + ". Lack of column DB type!")
				return fields
			}
			if value, ok := tag.Lookup("length"); ok { // field length
				colLen, _ := strconv.Atoi(value)
				field.Length = int32(colLen)
			}
			if value, ok := tag.Lookup("explain"); ok { // field explain
				field.Explain = value
			}
			if value, ok := tag.Lookup("key"); ok { // field index
				field.Index = value
			}
			switch strings.ToLower(field.Name) { // field value
			case def.NamePkId:
				break
			case def.NameServerId:
				field.Value = def.ServerId
			case def.NameSaveTime:
				field.Value = strconv.FormatInt(time.Now().Unix(), 10)
			default:
				field.Value = GetValueString(fVal.Interface())
			}
			fields = append(fields, field)
		}
	}
	return fields
}

// GetValueString returns the string-type value of v. If the type of v is not included here, return NULL
func GetValueString(v interface{}) string {
	var valueStr string
	switch v.(type) {
	case string:
		valueStr = v.(string)
	case int:
		valueStr = strconv.Itoa(v.(int))
	case int32:
		valueStr = strconv.Itoa(int(v.(int32)))
	case int64:
		valueStr = strconv.FormatInt(v.(int64), 10)
	case float32:
		valueStr = strconv.FormatFloat(float64(v.(float32)), 'E', -1, 32)
	case float64:
		valueStr = strconv.FormatFloat(v.(float64), 'E', -1, 64)
	case bool:
		valueStr = strconv.FormatBool(v.(bool))
	default:
		valueStr = "NULL"
	}
	return valueStr
}

// GetFieldDefString return the table column def statement part of the CREATE sql statement
func GetFieldDefString(fieldDef def.ColumnDef) string {
	fdStr := "`" + fieldDef.Name + "` " + fieldDef.Type
	if strings.ToLower(fieldDef.Type) == def.VARCHAR || strings.ToLower(fieldDef.Type) == def.TEXT {
		if fieldDef.Length <= 0 {
			fieldDef.Length = 255
		}
		fdStr += "(" + strconv.Itoa(int(fieldDef.Length)) + ")"
	}
	if strings.ToLower(fieldDef.Name) == def.NamePkId {
		fdStr += " AUTO_INCREMENT"
	}
	fdStr += ",\n"
	return fdStr
}

// GetNewCreateSql is the new create sql statement function. It allows you
// to customize primary key and normal key (Attention: If there has been 'pk_id'
// column, it will use 'pk_id' as primary key).
func GetNewCreateSql(log def.Logger) string {
	sqlFormer := "CREATE TABLE IF NOT EXISTS `%s`\n "
	sqlValue := "( %s "
	sqlTail := "%s) ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	havePkId := false
	combinedKeys := make(map[string][]string)
	var fieldsStr, indexStr, pkIdStr string
	var createTimeTemp, saveTimeTemp, actionIdTemp string
	fields := GetFields(log, true)
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		fieldStr := GetFieldDefString(field)
		if strings.ToLower(field.Name) == def.NamePkId {
			pkIdStr = "PRIMARY KEY (`" + field.Name + "`),\n"
			havePkId = true
		}
		if strings.ToLower(field.Name) == def.NameCreateTime {
			createTimeTemp = fieldStr
			continue
		}
		if strings.ToLower(field.Name) == def.NameSaveTime {
			saveTimeTemp = fieldStr
			continue
		}
		if strings.ToLower(field.Name) == def.NameActionId {
			actionIdTemp = fieldStr
			continue
		}
		if field.Index != "" {
			keys := strings.Split(field.Index, ",")
			for _, key := range keys {
				if !havePkId && key == def.IndexTypePK {
					pkIdStr = "PRIMARY KEY (`" + field.Name + "`),\n"
				} else {
					if key == def.IndexTypePK {
						continue
					}
					combinedKeys[key] = append(combinedKeys[key], field.Name)
				}
			}
		}
		fieldsStr += fieldStr
	}
	for key, fields := range combinedKeys {
		indexStr += "KEY `" + key + "` ("
		for _, field := range fields {
			indexStr += "`" + field + "`,"
		}
		indexStr = strings.TrimSuffix(indexStr, ",")
		indexStr += "),\n"
	}
	fieldsStr += createTimeTemp + saveTimeTemp + actionIdTemp
	indexStr = pkIdStr + indexStr
	indexStr = strings.TrimSuffix(indexStr, ",\n")
	if indexStr == "" {
		fieldsStr = strings.TrimSuffix(fieldsStr, ",\n")
	}
	return sqlFormer + fmt.Sprintf(sqlValue, fieldsStr) + fmt.Sprintf(sqlTail, indexStr)
}

// GetInsertSql returns the INSERT sql prepared statement of the logs
func GetInsertSql(log def.Logger) string {
	sqlFormer := "INSERT INTO `%s`"
	sqlBack := "( %s ) VALUES "
	var fieldsStr string
	fields := GetFields(log, true)
	for i := 0; i < len(fields); i++ {
		if strings.ToLower(fields[i].Name) == def.NamePkId {
			continue
		}
		sep := ","
		if i >= len(fields)-1 {
			sep = ""
		}
		fieldsStr += "`" + fields[i].Name + "`" + sep
	}
	return sqlFormer + fmt.Sprintf(sqlBack, fieldsStr)
}

// GetBatchInsertSql returns former part of INSERT sql prepared statement
func GetBatchInsertSql(log def.Logger) string {
	sqlHead := "INSERT INTO `%s`"
	sqlBack := "( %s ) VALUES "
	var fieldsStr string
	fields := GetFields(log, false)
	for i := 0; i < len(fields); i++ {
		if strings.ToLower(fields[i].Name) == def.NamePkId {
			continue
		}
		sep := ","
		if i >= len(fields)-1 {
			sep = ""
		}
		fieldsStr += "`" + fields[i].Name + "`" + sep
	}
	return sqlHead + fmt.Sprintf(sqlBack, fieldsStr)
}

// GetInsertValues returns the string values in batch insert sql
func GetInsertValues(log def.Logger) string {
	fields := GetFieldDefs(log, false)
	valueStrList := make([]string, 0)
	for j := 0; j < len(fields); j++ {
		field := fields[j]
		if strings.ToLower(field.Name) == def.NamePkId {
			continue
		}
		value := field.Value
		if field.Type == def.VARCHAR || field.Type == def.TEXT || field.Type == def.DATETIME {
			value = "'" + value + "'"
		}
		valueStrList = append(valueStrList, value)
	}
	return strings.Join(valueStrList, ",")
}

// GetUpdateSql returns the update sql statement
func GetUpdateSql(logs *list.List) string {
	sqlHead := "INSERT INTO `%s`"
	sqlBack := "( %s ) VALUES "
	var fieldsStr string
	fields := GetFields(logs.Front().Value.(def.Logger), false)
	for i := 0; i < len(fields); i++ {
		if strings.ToLower(fields[i].Name) == def.NamePkId {
			continue
		}
		sep := ","
		if i >= len(fields)-1 {
			sep = ""
		}
		fieldsStr += "`" + fields[i].Name + "`" + sep
	}
	sqlHead = sqlHead + fmt.Sprintf(sqlBack, fieldsStr)
	for v := logs.Front(); v != nil; v = v.Next() {
		log := v.Value.(def.Logger)
		values := GetInsertValues(log)
		sqlHead = sqlHead + "(" + values + ")"
		if v.Next() != nil {
			sqlHead += ","
		} else {
			sqlHead += " ON DUPLICATE KEY UPDATE "
		}
	}
	for i, field := range fields {
		sqlHead += "`" + field.Name + "`" + "=VALUES(`" + field.Name + "`)"
		if i != len(fields)-1 {
			sqlHead += ","
		}
	}
	return sqlHead
}
