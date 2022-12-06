package main

import (
	"RestApiWithProxy/pkg"
	"flag"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/create", pkg.Create)
	mux.HandleFunc("/make_friends", pkg.MakeFriends)
	mux.HandleFunc("/friends/", pkg.GetFriends)
	mux.HandleFunc("/delete_user", pkg.DeleteUser)
	mux.HandleFunc("/", pkg.NewAgeUser)

	srv1 := flag.String("srv1", "localhost:4000", "Сетевой адрес HTTP")
	flag.Parse()

	log.Printf("Запуск сервера на %s", *srv1)
	err := http.ListenAndServe(*srv1, mux)
	log.Fatal(err)

}
