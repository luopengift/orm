package orm

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"strings"
	"time"
)

type MySQL struct {
	*sql.DB
}

func Open(driver, dsn string) (*MySQL, error) {
	mysql := new(MySQL)
	err := mysql.Init(dsn)
	return mysql, err
}

func (mysql *MySQL) Init(dsn string) (err error) {
	mysql.DB, err = sql.Open("mysql", dsn)
	mysql.SetConnMaxLifetime(6 * time.Hour)
	mysql.SetMaxOpenConns(200)
	mysql.SetMaxIdleConns(100)
	return
}

func (mysql *MySQL) CreateTable(v interface{}) (sql.Result, error) {
	sql := CreateTableSQL(v)
	fmt.Println(strings.Join(strings.Split(sql, ","), ",\n"))
	return mysql.Exec(sql)
}

func (mysql *MySQL) DropTable(v interface{}) (sql.Result, error) {
	sql := DropTableSQL(v)
	return mysql.Exec(sql)
}

func (mysql *MySQL) Insert(tableName string, row map[string]interface{}) (sql.Result, error) {
	if err := mysql.Ping(); err != nil {
		return nil, err
	}
	var (
		keyList   []string
		valueList []interface{}
		markList  []string
	)
	for key, value := range row {
		keyList = append(keyList, key)
		valueList = append(valueList, value)
		markList = append(markList, "?")
	}
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(keyList, ", "), strings.Join(markList, ", "))
	fmt.Println(sql, fmt.Sprintf("%#v", valueList))
	stmt, err := mysql.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return nil, err
	}
	return stmt.Exec(valueList...)
	return nil, nil
}

// @tableName: 表名
// @selector: 更新的条件
// @row: 待更新数据
func (mysql *MySQL) Update(tableName string, selector, row map[string]interface{}) (sql.Result, error) {
	if err := mysql.Ping(); err != nil {
		return nil, err
	}
	var (
		updateList   []string
		selectorList []string
		valueList    []interface{}
	)
	for k, v := range row {
		updateList = append(updateList, k+"=?")
		valueList = append(valueList, v)
	}
	for k, v := range selector {
		selectorList = append(selectorList, k+"=?")
		valueList = append(valueList, v)

	}
	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, strings.Join(updateList, ", "), strings.Join(selectorList, ", "))
	fmt.Println(sql, fmt.Sprintf("%#v", valueList))
	stmt, err := mysql.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return nil, err
	}
	return stmt.Exec(valueList...)

}

func (mysql *MySQL) Delete(tableName string, selector map[string]interface{}) (sql.Result, error) {
	if err := mysql.Ping(); err != nil {
		return nil, err
	}
	var (
		selectorList []string
		valueList    []interface{}
	)
	for k, v := range selector {
		selectorList = append(selectorList, k+"=?")
		valueList = append(valueList, v)

	}
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, strings.Join(selectorList, ", "))
	fmt.Println(sql, fmt.Sprintf("%#v", valueList))
	stmt, err := mysql.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return nil, err
	}
	return stmt.Exec(valueList...)
}

func (mysql *MySQL) Query(tableName string, selector map[string]interface{}) (*sql.Rows, error) {
	sql := fmt.Sprintf("SELECT %s FROM %s", "*", tableName)
	if len(selector) > 0 {
		var whereList []string
		for k, v := range selector {
			whereList = append(whereList, fmt.Sprintf("%s='%v'", k, v))
		}
		sql = sql + " WHERE " + strings.Join(whereList, " AND ")
	}
	sql += ";"
	fmt.Println(sql)
	stmt, err := mysql.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return nil, err
	}
	return stmt.Query()
}

func ParseRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	typeList, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	var columnsList []reflect.Type
	for _, columnType := range typeList {
		columnsList = append(columnsList, columnType.ScanType())
		fmt.Println(columnType.Name())
	}
	fmt.Println(columnsList)

	num := len(columns)
	for rows.Next() {
		var (
			row    = make([][]byte, num)
			rowptr = make([]interface{}, num)
		)
		for idx, _ := range row {
			rowptr[idx] = &row[idx]
		}
		if err = rows.Scan(rowptr...); err != nil {
			return nil, err
		}
		var result = make(map[string]interface{}, num)
		for idx, value := range row {
			result[columns[idx]] = string(value)
		}
		results = append(results, result)
	}
	return results, nil
}

func (mysql *MySQL) And(selector map[string]interface{}) string {
	if len(selector) > 0 {
		var andList []string
		for k, v := range selector {
			andList = append(andList, fmt.Sprintf("%s='%v'", k, v))
		}
		return " WHERE " + strings.Join(andList, " AND ")
	}
	return ""
}

func ParseResult(result sql.Result) string {
	lastInsertId, err1 := result.LastInsertId()
	rowAffected, err2 := result.RowsAffected()
	return fmt.Sprintf("LastInsertId: %d, RowAffected: %d, err: %v %v", lastInsertId, rowAffected, err1, err2)

}

func TableName(v interface{}) string {
	structName := reflect.TypeOf(v).String()
	idx := strings.LastIndex(structName, ".")
	return structName[idx+1:]
}

func CreateTableSQL(v interface{}) string {
	var (
		columns []string
		tags    reflect.StructTag
	)
	sql := "CREATE TABLE IF NOT EXISTS %s (%s) ENGINE=Innodb DEFAULT CHARSET=utf8;"
	t := reflect.TypeOf(v).Elem()
	for i := 0; i < t.NumField(); i++ {
		tags = t.Field(i).Tag
		orm := tags.Get("orm")
		comment := tags.Get("comment")
		if comment != "" {
			orm += fmt.Sprintf(" COMMENT '%s'", comment)
		}
		columns = append(columns, orm)
	}
	return fmt.Sprintf(sql, TableName(v), strings.Join(columns, ","))
}

func AddColumnSQL(tableName, tags string, offset ...string) string {
	sql := "ALTER TABLE %s ADD COLUMN %s %s;"
	location := ""
	if len(offset) == 1 {
		switch offset[0] {
		case "first", "":
			location = offset[0]
		default:
			location = "AFTER" + offset[0]
		}
	}
	return fmt.Sprintf(sql, tableName, tags, location)
}

func DropTableSQL(v interface{}) string {
	sql := "DROP TABLE %s;"
	return fmt.Sprintf(sql, TableName(v))
}
