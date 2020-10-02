package commands

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	core "github.com/zokypesch/proto-lib/core"
	utils "github.com/zokypesch/proto-lib/utils"
	"github.com/zokypesch/proto-tools/config"
)

// ETL struct for creating proto from DB
type ETL struct {
	db     *gorm.DB
	dbPipe *gorm.DB
}

type tableEtl struct {
	DBName         string
	Name           string
	NameInGo       string
	NameOriginal   string
	Fields         []field
	PrimaryKeyName string
}

var etl *ETL

// NewETL for new protofrom db
func NewETL() CommandInterfacing {

	if etl == nil {
		cfg := config.Get()
		db := core.InitDB(cfg.DBAddressETL, cfg.DBNameETL, cfg.DBUserETL, cfg.DBPasswordETL, cfg.DBPort, false, 10, 5)
		dbPipe := core.InitDB(cfg.DBAddressETL, cfg.DBNamePipe, cfg.DBUserETL, cfg.DBPasswordETL, cfg.DBPort, false, 10, 5)

		log.Println("db connection success")
		etl = &ETL{db, dbPipe}
	}

	return etl
}

var (
	checkKey = []string{
		"host",
		"user",
		"password",
		"db",
		"kafka",
		"server",
		"group",
	}
)

// Execute for executing command
func (depl *ETL) Execute(args map[string]string) error {

	// var host, user, password, db, kafka, server, group string
	var err error

	for _, v := range checkKey {
		_, ok := args[v]
		if !ok {
			err = fmt.Errorf("cannot found key: %s", v)
			break
		}
	}

	if err != nil {
		return err
	}

	var tables []string
	db := core.InitDB(args["host"], args["db"], args["user"], args["password"], 3306, false, 10, 5)
	err = db.Raw("SHOW TABLES").Pluck(fmt.Sprintf("Tables_in_%s", args["db"]), &tables).Error
	if err != nil {
		return err
	}

	var tablesRes []tableEtl
	for _, tbl := range tables {
		var newfields []field
		db.Raw(`SELECT COLUMN_NAME AS name, DATA_TYPE as data_type, COLUMN_COMMENT AS comment, 
		CHARACTER_MAXIMUM_LENGTH as max_length, COLUMN_KEY as column_key, IS_NULLABLE as is_nullable, ORDINAL_POSITION AS ordinal_position
		FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ? AND table_schema = ?`, tbl, args["db"]).Scan(&newfields)

		tablesRes = append(tablesRes, tableEtl{NameOriginal: tbl, Name: utils.ConvertUnderscoreToCamel(tbl), Fields: newfields, DBName: args["db"]})
	}

	cfg := config.Get()

	for _, vTable := range tablesRes {
		var conn []Connection
		var tbl []Table
		var fl []KeyPair

		primaryName := ""
		foundPrimary := false

		for _, vField := range vTable.Fields {
			if vField.ColumnKey == "PRI" {
				foundPrimary = true
			}
		}

		for _, vField := range vTable.Fields {
			isType := 1
			isOmitEmpty := false

			if vField.ColumnKey == "PRI" {
				primaryName = vField.Name
				isType = 3
			}

			if vField.IsNullable == "YES" {
				isOmitEmpty = true
			}

			if !foundPrimary && vField.ColumnKey == "UNI" {
				primaryName = vField.Name
				isType = 3
			}

			fl = append(fl, KeyPair{
				FieldName:   vField.Name,
				ZQL:         fmt.Sprintf("[%s]", vField.Name),
				IsType:      isType,
				IsOmitEmpty: isOmitEmpty,
			})

		}

		// initial connection
		conn = append(conn, Connection{
			Name:         args["db"],
			DBAddress:    args["host"],
			DBSourceName: args["db"],
			DBUser:       args["user"],
			DBPassword:   args["password"],
			DBPort:       3306,
		})

		// etl for override connection
		conn = append(conn, Connection{
			Name:         cfg.DBNameETL,
			DBAddress:    cfg.DBAddressETL,
			DBSourceName: cfg.DBNameETL,
			DBUser:       cfg.DBUserETL,
			DBPassword:   cfg.DBPasswordETL,
			DBPort:       3306,
		})

		tbl = append(tbl, Table{
			ConnectionOverride: true,
			ConnectionName:     cfg.DBNameETL,
			TableName:          vTable.NameOriginal,
			Trigger:            "insert",
			Key:                fmt.Sprintf("[%s]", primaryName),
			Precheck:           true,
			PrecheckAg: []Agregator{
				{
					ConnectionName: cfg.DBNameETL,
					TableName:      vTable.NameOriginal,
					Fields:         "COUNT(1) AS exist",
					Where: []KeyPair{
						{
							FieldName: primaryName,
							ZQL:       fmt.Sprintf("[%s]", primaryName),
							IsType:    3,
						},
					},
				},
			},
			PrecheckCond:    []string{fmt.Sprintf("[precheck.%s.exist] == '1'", vTable.NameOriginal)},
			PrecheckAction:  "skip",
			TargetTableName: vTable.NameOriginal,
			TargetMode:      "insert",
			NextAction:      "saving_db",
			TopicTarget:     "",
			Initiate: Initiate{
				Status: false,
			},
			TargetTableField: fl,
		})

		// for update
		tbl = append(tbl, Table{
			ConnectionOverride: true,
			ConnectionName:     cfg.DBNameETL,
			TableName:          vTable.NameOriginal,
			Trigger:            "update",
			Key:                fmt.Sprintf("[%s]", primaryName),
			Precheck:           true,
			PrecheckAg: []Agregator{
				{
					ConnectionName: cfg.DBNameETL,
					TableName:      vTable.NameOriginal,
					Fields:         "COUNT(1) AS exist",
					Where: []KeyPair{
						{
							FieldName: primaryName,
							ZQL:       fmt.Sprintf("[%s]", primaryName),
							IsType:    3,
						},
					},
				},
			},
			PrecheckCond:    []string{fmt.Sprintf("[precheck.%s.exist] == '0'", vTable.NameOriginal)},
			PrecheckAction:  "skip",
			TargetTableName: vTable.NameOriginal,
			TargetMode:      "update",
			NextAction:      "saving_db",
			TopicTarget:     "",
			Initiate: Initiate{
				Status: false,
			},
			TargetTableField: fl,
		})

		dataFinal := Conf{
			KafkaAddress: args["kafka"],
			Group:        args["group"],
			Server:       args["server"],
			DBName:       args["db"],
			Tables:       tbl,
		}

		resJSON, err := json.Marshal(&dataFinal)
		coll := SourceConnection{
			Collection: conn,
		}

		if err != nil {
			log.Fatalln("failed serelize: ", err.Error())
		}
		resJSONConn, err := json.Marshal(&coll)

		if err != nil {
			log.Fatalln("failed serelize: ", err.Error())
		}

		var existPipe []int64
		err = depl.dbPipe.Raw("SELECT COUNT(1) AS total FROM pipeline where code = ?", vTable.NameOriginal).Pluck("total", &existPipe).Error

		if err != nil {
			log.Println("failed to query pipeline")
			break
		}

		if existPipe[0] > 0 {
			err = depl.dbPipe.Exec("UPDATE pipeline SET connection = ?, config = ? WHERE code = ?", resJSONConn, resJSON, vTable.NameOriginal).Error
			if err != nil {
				log.Println("failed to update pipeline")
				break
			}
			continue
		}

		err = depl.dbPipe.Exec("INSERT INTO pipeline (code, connection, config) VALUES(?, ?, ?) ", vTable.NameOriginal, resJSONConn, resJSON).Error
		if err != nil {
			log.Println("failed to insert pipeline: ", err)
			break
		}
	}

	if err != nil {
		return err
	}

	return nil
}

// SourceConnection for connection source
type SourceConnection struct {
	Collection []Connection `json:"collection"`
}

// Connection for connection type
type Connection struct {
	Name         string `json:"name"`
	DBAddress    string `json:"dbAddress"`
	DBSourceName string `json:"dbSourceName"`
	DBUser       string `json:"dbUser"`
	DBPassword   string `json:"dbPassword"`
	DBPort       int    `json:"dbPort"`
	DBType       int    `json:"dbType"` // 0 mysql, 1: postgresql
}

// Conf data yaml for configuration
type Conf struct {
	KafkaAddress   string              `json:"kafkaAddress"`
	Group          string              `json:"group"`
	Server         string              `json:"server"`
	DBName         string              `json:"dbName"`
	Tables         []Table             `json:"tables"`
	Offset         string              `json:"offset"`
	ConnCollection map[string]*gorm.DB `json:",omitempty"`
}

// Table for table configuration
type Table struct {
	ConnectionOverride bool        `json:"connectionOverride"`
	ConnectionName     string      `json:"connectionName"`
	TableName          string      `json:"tableName"`              // as table name for topic, ie: fmt.Sprintf("%s.%s.%s", Server, DBName, TableName)
	Trigger            string      `json:"trigger"`                // insert, update, insert_or_update -> precheckCond, precheckAg cannot be empty
	ConditionZQL       []string    `json:"conditionZQL,omitempty"` //ie: [name||fullname == 'udin'] see zokypesch Query Language
	Key                string      `json:"key"`
	Precheck           bool        `json:"precheck"`
	PrecheckAg         []Agregator `json:"precheckAg"`
	PrecheckCond       []string    `json:"precheckCond"`
	PrecheckAction     string      `json:"precheckAction"` // update or skip
	Agregator          []Agregator `json:"agregator,omitempty"`
	TargetTableName    string      `json:"targetTableName"`
	TargetTableField   []KeyPair   `json:"targetTableField"` //ie: fieldname = table_name.fieldname
	TargetMode         string      `json:"targetMode"`       // insert, update
	NextAction         string      `json:"nextAction"`       // saving_db, publish_message, both
	TopicTarget        string      `json:"topicTarget"`
	Initiate           Initiate    `json:"initiate"`
}

// Initiate for initiate struct
type Initiate struct {
	Status     bool     `json:"status"`
	Condition  []string `json:"condition,omitempty"` // ["id > 1000","id < 1000"]
	Connection string   `json:"connection,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Offset     int      `json:"offset,omitempty"`
	Fields     []string `json:"fields,omitempty"`
	OrderBy    []string `json:"orderBy,omitempty"`
}

// Agregator for agregator svc
type Agregator struct {
	ConnectionName string    `json:"connectionName"`
	TableName      string    `json:"tableName"`
	AliasName      string    `json:"aliasName"`
	Fields         string    `json:"fields"` // ie: field_name1 AS name, field_email AS email
	Where          []KeyPair `json:"where"`  // ie: join_batch = '' OR field_from_another = '[table_name.fieldname]' AND (field_from_another2 = [table_name.fieldname])
	OrderLimit     string    `json:"orderLimit"`
}

// KeyPair for key pair where
type KeyPair struct {
	FieldName   string `json:"fieldName"`
	ZQL         string `json:"ZQL"`
	IsType      int    `json:"isType"` // 0: default, 1: for insert, 2: for update where condition, 3: all
	IsOmitEmpty bool   `json:"isOmitEmpty"`
}
