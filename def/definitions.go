// The def package defines the all the const and struct we need
package def

// Log database type
const (
	MySql = 1
	Mongo = 2
)

// Log record type
const (
	Single = 1
	Batch  = 2
)

// Over BatchCleanTime, clean all the logs in the channel buffer
const (
	BatchCleanTime = 10
)

// Log table split type
const (
	RollTypeDay   = 1
	RollTypeMonth = 2
	RollTypeYear  = 3
)

// Mysql column types
const (
	TINY_INT  = "tinyint"
	SMALL_INT = "smallint"
	MEDIUMINT = "mediumint"
	INT       = "int"
	BIG_INT   = "bigint"
	FLOAT     = "float"
	DOUBLE    = "double"
	DATE      = "date"
	TIME      = "time"
	DATETIME  = "datetime"
	TIMESTAMP = "timestamp"
	CHAR      = "char"
	VARCHAR   = "varchar"
	TEXT      = "text"
)

const DataBase = MySql

const (
	NamePkId       = "pk_id"
	NameServerId   = "server_id"
	NameCreateTime = "create_time"
	NameSaveTime   = "save_time"
	NameActionId   = "action_id"
)

var ServerId string
var BatchNum int

// Logger is the interface which all the logs MUST implement
type Logger interface {
	TableName() string // return the name of the db table where the log is going to insert
	RollType() int32   // return the db table slice type
	SaveType() int32   // return the log should be recorded single or batch
}

// BasePlayerLog contains the basic player log's attributes, all the player logs should extend of it
// or includes them in it by yourself
type BasePlayerLog struct {
	PkId       string `type:"int" explain:"自增主键" name:"pk_id"`
	PlayerId   string `type:"varchar" length:"255" explain:"玩家id" name:"player_id"`
	ServerId   string `type:"varchar" length:"255" explain:"服务器id" name:"server_id"`
	CreateTime int64  `type:"bigint" explain:"创建时间" name:"create_time"`
	SaveTime   int64  `type:"bigint" explain:"保存时间" name:"save_time"`
	ActionId   string `type:"varchar" length:"255" explain:"行为id" name:"action_id"`
}

// BasePlayerLog contains the basic server log's attributes, all the server logs should extend of it
// or includes them in it by yourself
type BaseServerLog struct {
	PkId       string `type:"int" explain:"自增主键" name:"pk_id"`
	ServerId   string `type:"varchar" length:"255" explain:"服务器id" name:"server_id"`
	CreateTime int64  `type:"bigint" explain:"创建时间" name:"create_time"`
	SaveTime   int64  `type:"bigint" explain:"保存时间" name:"save_time"`
	ActionId   string `type:"varchar" length:"255" explain:"行为id" name:"action_id"`
}

// ColumnDef defines the field info of logs, and it helps to build CREATE and INSERT sql statements
type ColumnDef struct {
	Name    string // the column name
	Type    string // the column type
	Length  int32  // length of this column
	Value   string // value of this column
	Explain string // explain of this column
}

// LogCounter counts the logs number we deal successfully
type LogCounter struct {
	TotalCount uint64 // the total count
	Count      uint64 // the count in one of the monitor tick
}
