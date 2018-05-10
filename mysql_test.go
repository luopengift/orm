package orm

import (
	"fmt"
	"github.com/luopengift/log"
	//	"github.com/luopengift/types"
	"testing"
	"time"
)

type User struct {
	Id      int64     `json:"omitempty" orm:"id int(11) NOT NULL PRIMARY KEY AUTO_INCREMENT"`
	Name    string    `json:"omitempty" orm:"name varchar(255) UNIQUE"`
	Age     int       `json:"omitempty" orm:"age int(3)"`
	Created time.Time `json:"omitempty" orm:"created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP"`
	Updated time.Time `json:"omitempty" orm:"updated timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

// Dsn: "user:password@tcp(127.0.0.1:3306)/test",
func Test_MySQL(t *testing.T) {
	mysql := MySQL{}
	if err := mysql.Init("root:5pPWC5dIrGtw%5sz@tcp(127.0.0.1)/mytest"); err != nil {
		fmt.Println(err)
		return
	}
	if err := mysql.Ping(); err != nil {
		fmt.Println(err)
		return
	}

	res, err := mysql.DropTable(&User{})
	log.Info("drop table: %#v, %v", res, err)
	res, err = mysql.CreateTable(new(User))
	log.Info("create table: %#v, %v", res, err)
	res, err = mysql.Insert("User", map[string]interface{}{"name": "CLC", "age": 3})
	res, err = mysql.Insert("User", map[string]interface{}{"name": "LCL", "age": 23})
	fmt.Println(res, err)

	//time.Sleep(10 * time.Second)
	//res, err = mysql.Update("User", map[string]interface{}{"name": "luopeng"}, map[string]interface{}{"age": 5})

	//time.Sleep(10 * time.Second)
	//res, err = mysql.Delete("User", map[string]interface{}{"age": 5})

	result, err := mysql.Query("User", nil)
	fmt.Println(result, "$$", err)
	defer result.Close()

	rest, err := ParseRows(result)
	fmt.Println("=>", rest, err)

	for _, v := range rest {
		log.Info("==> %#v", v)
	}

	//sql := SQL{}
	//sql.Table("User").Where(map[string]interface{}{"name":"CLC"}).Columns("name", "age")
	//fmt.Println(sql)

}
