package pkg

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Users struct {
	Name    string `json:"name"`
	Age     string `json:"age"`
	Friends []int  `json:"friends"`
}
type Mfriends struct {
	SourceId string `json:"sourceid"`
	TargetId string `json:"targetid"`
}

var id int

const (
	host     = "localhost"
	port     = 49153
	user     = "postgres"
	password = ""
	dbname   = "postgres"
)

func Create(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		var u Users

		if err = json.Unmarshal(content, &u); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		db := DataBaseConn()
		db.QueryRow("select max(id) from users u").Scan(&id)
		if id == 0 {
			id = 1
		} else {
			id++
		}
		sqlStatement := `
		INSERT INTO users (id, age, first_name)
		VALUES ($1,$2,$3)`
		_, err = db.Exec(sqlStatement, id, u.Age, u.Name)
		if err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("ID пользователя: %d", id)))

		defer db.Close()
		defer r.Body.Close()
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func MakeFriends(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		content, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		var mf Mfriends
		if err := json.Unmarshal(content, &mf); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		var friendship bool
		db := DataBaseConn()
		defer db.Close()
		defer r.Body.Close()
		db.QueryRow("select count(*)  from users where id = $1 and $2 =any(friends)", mf.SourceId, mf.TargetId).Scan(&friendship)

		if friendship != false {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Пользователи уже дружат"))
			return
		} else {
			var verab int
			db.QueryRow("select count(*) from users where id = $1 or id = $2", mf.SourceId, mf.TargetId).Scan(&verab)
			if verab == 2 {
				db.Exec("update users set friends = array_append(friends,$1) where id =$2", mf.SourceId, mf.TargetId)
				db.Exec("update users set friends = array_append(friends,$1) where id =$2", mf.TargetId, mf.SourceId)

				names, _ := db.Query("select first_name from users u where id = $1 or id = $2", mf.SourceId, mf.TargetId)

				var listOfNames [2]string
				for i := 0; names.Next(); i++ {
					names.Scan(&listOfNames[i])
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(listOfNames[0] + " и " + listOfNames[1] + " теперь друзья"))

				return

			} else {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte("Пользователя не существует"))
				return
			}

		}

	}
	w.WriteHeader(http.StatusBadRequest)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {

	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	var mf Mfriends
	if err := json.Unmarshal(content, &mf); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	db := DataBaseConn()
	var userName string

	db.QueryRow("select first_name from users where id=$1", mf.TargetId).Scan(&userName)
	db.Exec("DELETE FROM users where id=$1", mf.TargetId)
	db.Exec("update users set friends = array_remove(friends,$1)", mf.TargetId)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Пользователь %s успешно удален.", userName)))
	defer db.Close()
	defer r.Body.Close()
	return

}

func GetFriends(w http.ResponseWriter, r *http.Request) {

	text := r.URL.Path
	ind := strings.Index(text, "s/")
	str := text[ind+2:]

	w.WriteHeader(http.StatusOK)

	var name string
	var names string
	var nameUser string
	db := DataBaseConn()

	rows, _ := db.Query("select first_name FROM users u WHERE $1 = any (friends)", str)
	for rows.Next() {
		rows.Scan(&name)
		names = names + name + "\n"
	}

	db.QueryRow("select first_name FROM users u WHERE id = $1", str).Scan(&nameUser)

	result := nameUser + " :\n" + names

	w.Write([]byte("Друзья пользователя " + result))
	defer db.Close()
	defer r.Body.Close()
	return

}

func NewAgeUser(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	text := r.URL.Path
	ind := strings.Index(text, "_")
	str := text[ind+1:]

	var u Users
	if err := json.Unmarshal(content, &u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	db := DataBaseConn()
	var userName string
	db.QueryRow("select first_name from users where id=$1", str).Scan(&userName)
	db.Query("UPDATE users SET age=$1 WHERE id=$2", u.Age, str)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Возраст пользователя %s обновлен", userName)))
	defer db.Close()
	defer r.Body.Close()
	return

}

func DataBaseConn() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, errPs := sql.Open("postgres", psqlInfo)
	if errPs != nil {
		panic("failed conn" + errPs.Error())
	}

	return db
}
