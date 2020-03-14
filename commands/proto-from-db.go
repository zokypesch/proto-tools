package commands

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode"

	core "github.com/zokypesch/proto-lib/core"
	utils "github.com/zokypesch/proto-lib/utils"
	tmp "github.com/zokypesch/proto-tools/template"
)

// ProtoFromDB struct for creating proto from DB
type ProtoFromDB struct{}

// NewProtoFromDB for new protofrom db
func NewProtoFromDB() CommandInterfacing {
	return &ProtoFromDB{}
}

type dbscheme struct {
	Name   string
	Tables []*table
}

type table struct {
	Name           string
	NameInGo       string
	NameOriginal   string
	Fields         []*field
	Joins          []*join
	GetAll         *getall
	PrimaryKeyName string
}

type getall struct {
	Page    int
	PerPage int
}

type join struct {
	ReferencedTableName       string
	ReferencedTableOriginal   string
	ReferencedColumnName      string
	ReferencedColumnNameProto string
	TableName                 string
	ColumnName                string
	ConstraintName            string
	FieldName                 string
	Option                    string
	Repeated                  bool
	OrdinalPosition           int
}

type field struct {
	Name            string
	DataType        string
	DataTypeProto   string
	ColumnKey       string
	MaxLength       int
	Comment         string
	Option          string
	Required        bool
	PrimaryKey      bool
	IsNullable      string
	OrdinalPosition int
	NameProto       string
}

// Execute for executing command
func (cmd *ProtoFromDB) Execute(args map[string]string) error {
	log.Println("Converting database scheme from database")

	if len(args) < 4 {
		return fmt.Errorf("Not enough parameter, host={your_host_mysql_database} name={your_db_name} user={your_user_db} pass={your_pass_db}")
	}

	var ok bool
	var host, dbName, user, pass string

	host, ok = args["host"]
	if !ok {
		return fmt.Errorf("Not enough parameter, host not found, put host=your_host_mysql_database")
	}

	dbName, ok = args["name"]
	if !ok {
		return fmt.Errorf("Not enough parameter, dbName not found, put name=your_db_name")
	}

	user, ok = args["user"]
	if !ok {
		return fmt.Errorf("Not enough parameter, user not found, put port=your_user_db")
	}

	pass, ok = args["password"]
	if !ok {
		return fmt.Errorf("Not enough parameter, pass not found, put port=your_pass_db")
	}

	// Initial database
	dbName = utils.ToLowerFirst(dbName)

	var tables []string
	var err error

	db := core.InitDBWithoutLog(host, dbName, user, pass, 3306)
	err = db.Raw("SHOW TABLES").Pluck(fmt.Sprintf("Tables_in_%s", dbName), &tables).Error

	if err != nil {
		return err
	}

	var tablesRes []*table
	for _, tbl := range tables {
		var newfields []*field
		db.Raw(`SELECT COLUMN_NAME AS name, DATA_TYPE as data_type, COLUMN_COMMENT AS comment, 
		CHARACTER_MAXIMUM_LENGTH as max_length, COLUMN_KEY as column_key, IS_NULLABLE as is_nullable, ORDINAL_POSITION AS ordinal_position
		FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = ?`, tbl).Scan(&newfields)

		tablesRes = append(tablesRes, &table{NameOriginal: tbl, Name: utils.ConvertUnderscoreToCamel(tbl), Fields: newfields})
	}

	scheme := dbscheme{Name: utils.ToLowerFirst(dbName), Tables: tablesRes}

	for _, vTable := range scheme.Tables {
		primaryName := ""
		for _, vField := range vTable.Fields {
			switch vField.DataType {
			case "bigint", "int", "tinyint", "float", "smallint":
				vField.DataTypeProto = "int64"
			case "decimal", "double":
				vField.DataTypeProto = "float"
			case "varchar", "text", "json", "char", "enum", "tinytext":
				vField.DataTypeProto = "string"
			case "timestamp", "date", "datetime", "time":
				vField.DataTypeProto = "google.protobuf.Timestamp"
			default:
				vField.DataTypeProto = vField.DataType
			}

			usingComma := false
			options := "["

			if vField.ColumnKey == "PRI" {
				vField.PrimaryKey = true
				usingComma = true
				options += "(isPrimaryKey) = true,"
				primaryName = vField.Name
			}

			com := strings.Replace(vField.Comment, ",", "", -1)
			args := strings.Split(com, " ")
			param := utils.SplitSliceParamsToMap(args, "=")

			additional := ""
			if vField.MaxLength > 0 {
				additional = fmt.Sprintf(",max=%d", vField.MaxLength)
			}

			_, ok := param["required"]
			if ok || (vField.IsNullable == "NO" && vField.DataType != "timestamp") {
				vField.Required = true
				usingComma = true

				options += fmt.Sprintf("(required) = true,(required_type)=\"required%s\"", additional)
			}
			// options = strings.TrimSuffix(options, ",")
			if usingComma {
				options += ","
			}
			nameInProto := utils.ToLowerFirst(utils.ConvertUnderscoreToCamel(vField.Name))
			vField.NameProto = nameInProto
			options += fmt.Sprintf("json_name=\"%s\"];", nameInProto)
			vField.Option = options

		}
		vTable.PrimaryKeyName = primaryName
	}

	for _, vTable := range scheme.Tables {
		var jn []*join
		db.Raw(`SELECT 
			ke.referenced_table_name,
			ke.referenced_column_name,
			ke.table_name,
			ke.column_name,
			ke.constraint_name
		FROM
			information_schema.KEY_COLUMN_USAGE ke
		WHERE
			ke.referenced_table_name IS NOT NULL
				AND table_schema = 'transaction'
						AND table_name=?
		ORDER BY ke.referenced_table_name;`, vTable.NameOriginal).Scan(&jn)

		vTable.Joins = jn
	}

	for _, vTable := range scheme.Tables {
		lastField := len(vTable.Fields)
		for _, vJoin := range vTable.Joins {
			lastField++
			vJoin.ReferencedTableOriginal = vJoin.ReferencedTableName
			vJoin.ReferencedTableName = utils.ConvertUnderscoreToCamel(vJoin.ReferencedTableName)
			opt := strings.Split(vJoin.ConstraintName, "__")
			if len(opt) == 2 && opt[1] == "many" {
				vJoin.Repeated = true
			}
			vJoin.ReferencedColumnNameProto = utils.ConvertUnderscoreToCamel(vJoin.ReferencedColumnName)
			vJoin.Option = fmt.Sprintf("[(foreignKey) = \"%s\"]", vJoin.ReferencedColumnNameProto)
			vJoin.OrdinalPosition = lastField
		}
		//cahge per page
		lastField++
		newGetAll := &getall{
			Page:    lastField,
			PerPage: lastField + 1,
		}
		vTable.GetAll = newGetAll
	}

	// mapping datas to template
	tmpl := tmp.TmplProtoFromDb
	buf := bytes.NewBuffer(nil)
	err = template.Must(template.New("").Funcs(
		template.FuncMap{
			"unescape":     unescape,
			"ucfirst":      ucFirst,
			"allowRequest": allowRequest,
		}).
		Parse(tmpl)).Execute(buf, scheme)

	if err != nil {
		return err
	}

	location := fmt.Sprintf("%s/grpc/proto/%s", utils.GetFullPath(), scheme.Name)
	if _, err = os.Stat(location); os.IsNotExist(err) {
		os.MkdirAll(location, 0755)
	}
	log.Println(err)
	filePath := fmt.Sprintf("%s/%s.proto", location, scheme.Name)
	if _, err := os.Stat(location); os.IsNotExist(err) {
		os.Mkdir(location, 0755)
	}
	log.Println("success generate proto file " + utils.GetFullPath() + "/grpc/proto/" + scheme.Name + ".proto")
	err = ioutil.WriteFile(filePath, buf.Bytes(), 0644)

	if err != nil {
		return err
	}

	return nil
}

func unescape(s string) template.HTML {
	return template.HTML(s)
}

func ucFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

var ignoreField = []string{
	"id",
	"created_at",
	"updated_at",
	"created_by",
	"updated_by",
}

func allowRequest(field string) bool {
	for _, v := range ignoreField {
		if v == field {
			return false
		}
	}
	return true
}

/**
Query for show comment tables
SELECT table_comment
    FROM INFORMATION_SCHEMA.TABLES
    WHERE table_schema='my_cool_database'
        AND table_name='user_skill';
*/
