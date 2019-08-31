// The utils package provides all the tool functions of log dealing
package utils

import (
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
func GetFields(log interface{}) []def.ColumnDef {
	typ := reflect.TypeOf(log)
	logName := typ.Name()
	_, exist := LogFieldsDefinitions[logName]
	if !exist {
		LogFieldsDefinitions[logName] = GetFieldDefs(log)
	}
	return LogFieldsDefinitions[logName]
}

// GetFieldDefs returns the all the logs's attributes by reflection
func GetFieldDefs(log interface{}) []def.ColumnDef {
	val := reflect.ValueOf(log)
	typ := reflect.TypeOf(log)
	fields := make([]def.ColumnDef, 0)
	for i := 0; i < val.NumField(); i++ {
		fVal := val.Field(i)
		fTyp := typ.Field(i)
		if fVal.Kind() == reflect.Struct {
			fields = append(fields, GetFields(fVal.Interface())...)
		} else {
			field := def.ColumnDef{}
			tag := fTyp.Tag
			if value, ok := tag.Lookup("name"); ok { // 字段名
				field.Name = value
			} else {
				field.Name = strings.ToLower(fTyp.Name)
			}
			if value, ok := tag.Lookup("type"); ok { // 字段类型
				field.Type = value
			} else {
				log2.Println("Error def of " + typ.Name() + ". Lack of column DB type!")
				return fields
			}
			if value, ok := tag.Lookup("length"); ok { // 字段长度
				colLen, _ := strconv.Atoi(value)
				field.Length = int32(colLen)
			}
			if value, ok := tag.Lookup("explain"); ok { // 字段长度
				field.Explain = value
			}
			switch strings.ToLower(field.Name) { // 字段值
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

// GetCreateSql returns the CREATE sql statement of the logs
func GetCreateSql(log def.Logger) string {
	sqlFormer := "CREATE TABLE IF NOT EXISTS `%s`\n "
	sqlBack := "( %sPRIMARY KEY (`%s`)\n ) ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	var fieldsStr, pkIdStr string
	fields := GetFields(log)
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		fieldsStr += GetFieldDefString(field)
		if strings.ToLower(field.Name) == def.NamePkId {
			pkIdStr = strings.ToLower(field.Name)
		}
	}
	return sqlFormer + fmt.Sprintf(sqlBack, fieldsStr, pkIdStr)
}

// GetInsertSql returns the INSERT sql prepared statement of the logs
func GetInsertSql(log def.Logger) string {
	sqlFormer := "INSERT INTO `%s`"
	sqlBack := "( %s ) VALUES "
	var fieldsStr string
	fields := GetFields(log)
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
	fields := GetFields(log)
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
	fields := GetFieldDefs(log)
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

// GetPreparedInsertValues returns the slice containing all the values in the same order as the prepared insert sql statement,
// except PkId
func GetPreparedInsertValues(log def.Logger) []string {
	fields := GetFieldDefs(log)
	values := make([]string, 0)
	for i := 0; i < len(fields); i++ {
		if strings.ToLower(fields[i].Name) == def.NamePkId {
			continue
		}
		values = append(values, fields[i].Value)
	}
	return values
}
