package main
import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"fmt"
	"os/exec"
	"github.com/op/go-logging"
	//"io"
	"strings"
	"strconv"
)

var log = logging.MustGetLogger("test")
var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)


type xen_scan interface {
	Scan() bool;
	Open() bool;
	Init() bool;
}

type xen_list_struct struct {
	name string
	id int
	state string
	mem int
	vcores int
	uptime string
}

type xen_scan_struct struct {
	sql_db *sql.DB;
	sql_string string
}

func (x xen_scan_struct) Open() bool{
	var err interface{} = nil
	println("Test1: " + x.sql_string)
	x.sql_db , err = sql.Open("sqlite3", x.sql_string)
	if err != nil {
		log.Error( "OPEN XEN CONFIG FROM")
		return false
	}
	//defer db.Close()
	//arr := make([]byte, unsafe.Sizeof(*db))
	//copy(arr, *db)


	//x.sql_db = db
	return true

}

func (x xen_scan_struct) Init() bool{
	os.Remove(x.sql_string)
	x.Open()
	sqlStmt := `create table data (id integer not null primary key, name text);`
	_ , err := x.sql_db.Exec(sqlStmt)
	if err != nil {
		log.Error("OPEN XEN CONFIG FROM" +  sqlStmt)
		return false
	}

	return true
}

func (x xen_scan_struct) Scan() bool{
	/*
	db, err := sql.Open("sqlite3", x.sql_string)
	if err != nil {
		log.Error("scan XEN open sql")
		return false
	}

	sqlStmt := `create table data (id integer not null primary key, name text);`
	_ , err = db.Exec(sqlStmt)
	if err != nil {
		log.Error(" Scan " + sqlStmt)
		return false
	}

	exec.Command("xl", "list")

	var data = "test"
	sqlStmt = "`INSERT INTO Relation [( Attribut+ )] VALUES name " +  data + ");`"
	_ , err = db.Exec(sqlStmt)
	if err != nil {
		log.Error(" insert " + sqlStmt)
		return false
	}
	*/

	ret, err:= exec.Command("xl", "list").CombinedOutput()
	if err != nil {
		println("error occured")
		fmt.Printf("%s", err)
	}

	test := string(ret)
	test2 := strings.Fields(test)

	bound := len(test2)/6

	test_struct := [20]xen_list_struct{}

	for x := 1; x < bound; x++{
		test_struct[x - 1].name = test2[0 + x*6]
		i, err := strconv.Atoi(test2[1 + x*6])
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		test_struct[x - 1].id = i
		i, err = strconv.Atoi(test2[2 + x*6])
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		test_struct[x - 1].mem = i
		i, err = strconv.Atoi(test2[3 + x*6])
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		test_struct[x - 1].vcores = i
		test_struct[x - 1].state = test2[4 + x*6]
		test_struct[x - 1].uptime = test2[5 + x*6]

	}
	println(test2[6])
	println(test_struct[0].vcores)
	//println(ret[1])

	return true
}

func main(){
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")

	// Set the backends to be used.
	logging.SetBackend(backend1Leveled, backend2Formatter)

	x := xen_scan_struct{sql_db:&sql.DB{},sql_string:"./foo.db" }
	x.Scan()
}