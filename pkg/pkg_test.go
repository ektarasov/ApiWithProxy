package pkg

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
)

const name = "TestovTest"
const age = 21

func delete(db *sql.DB) {

	db.Exec("delete from users where first_name=$1", name)

}

func create() *http.Response {

	textbytes := []byte(fmt.Sprintf("{ \"name\": \"%s\", \"age\": \"%d\"}", name, age))
	resp, err := http.Post("http://localhost:4000/create", "text/plain", bytes.NewBuffer(textbytes))
	if err != nil {
		log.Fatal(err)
	}
	return resp
}

func makeFriends() (string, string) {

	req1 := create().Body
	req2 := create().Body
	targetId := readId(req1)
	sourceId := readId(req2)
	textbytes := []byte(fmt.Sprintf("{ \"sourceid\": \"%s\", \"targetid\": \"%s\"}", sourceId, targetId))
	http.Post("http://localhost:4000/make_friends", "text/plain", bytes.NewBuffer(textbytes))
	return targetId, sourceId

}

func readId(req1 io.ReadCloser) string {

	req1answ, _ := ioutil.ReadAll(req1)
	parseId := string(req1answ)[strings.Index(string(req1answ), ":")+2:]
	return parseId

}

func TestCreate(t *testing.T) {
	exp := 0
	res := 0

	db := DataBaseConn()
	db.QueryRow("select count(*) from users where age =$1 and first_name =$2", age, name).Scan(&exp)

	create()

	db.QueryRow("select count(*) from users where age =$1 and first_name =$2", age, name).Scan(&res)

	delete(db)

	if res == exp {
		t.Fail()
	}
}

func TestDataBaseConn(t *testing.T) {
	exp := 1
	db := DataBaseConn()
	db.Ping()
	res := db.Stats().OpenConnections
	if exp != res {
		t.Fail()
	}
}

func TestMakeFriends(t *testing.T) {
	var friendship bool

	db := DataBaseConn()

	targetId, sourceId := makeFriends()

	db.QueryRow("select count(*)  from users where id = $1 and $2 =any(friends)", sourceId, targetId).Scan(&friendship)
	delete(db)
	if friendship != true {
		t.Fail()
	}

}

func TestDeleteUser(t *testing.T) {
	userCheck := 1

	db := DataBaseConn()

	req1 := create().Body
	targetId := readId(req1)
	textbytes := []byte(fmt.Sprintf("{\"targetid\": \"%s\"}", targetId))
	http.Post("http://localhost:4000/delete_user", "text/plain", bytes.NewBuffer(textbytes))
	db.QueryRow("select count(*)  from users where id=$1", targetId).Scan(&userCheck)
	if userCheck != 0 {
		t.Fail()
	}

}

func TestGetFriends(t *testing.T) {
	exp := "Друзья пользователя TestovTest :\nTestovTest\n"

	db := DataBaseConn()
	_, sourceId := makeFriends()

	req3, _ := http.Post("http://localhost:4000/friends/"+sourceId, "text/plain", nil)
	req3answ, _ := ioutil.ReadAll(req3.Body)
	req3str := string(req3answ)
	if exp != req3str {
		t.Fail()
	}
	delete(db)
}

func TestNewAgeUser(t *testing.T) {
	exp := 100
	res := 0

	db := DataBaseConn()
	req1 := create().Body
	targetId := readId(req1)

	textbytes := fmt.Sprintf("{ \"age\": \"%d\"}", exp)
	http.Post("http://localhost:4000/user_"+targetId, "text/plain", bytes.NewBuffer([]byte(textbytes)))
	db.QueryRow("select age from users where id = $1", targetId).Scan(&res)
	delete(db)
	if exp != res {
		t.Fail()
	}

}
