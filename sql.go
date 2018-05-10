package orm

import (
	"fmt"
	//	"reflect"
	"strings"
)

type SQL struct {
	table   string
	action  string
	columns []string
	where   map[string]interface{}
	sort    string
	desc    string
	join    string
	group   string
	limit   int
}

func (s *SQL) Table(table string) *SQL {
	s.table = table
	return s
}

func (s *SQL) Columns(columns ...string) *SQL {
	s.columns = columns
	return s
}

func (s *SQL) Where(where map[string]interface{}) *SQL {
	s.where = where
	return s
}

func (s SQL) String() string {
	sql := "SELECT "
	columns := "*"
	if len(s.columns) > 0 {
		columns = strings.Join(s.columns, ", ")
	}
	sql += columns + " FROM " + s.table
	var whereList []string
	if len(s.where) > 0 {
		for k, v := range s.where {
			whereList = append(whereList, fmt.Sprintf("%s='%v'", k, v))
		}
		sql += " WHERE " + strings.Join(whereList, " AND ")
	}
	return sql
}
