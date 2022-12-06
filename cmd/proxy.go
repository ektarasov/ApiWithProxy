package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	counter int = 0
)

func main() {

	srvProxy := flag.String("proxy", "localhost:9000", "Сетевой адрес HTTP")
	flag.Parse()
	serversList := flag.Args()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		textbytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		textUrl := r.URL.Path
		fmt.Println(string(textbytes))

		resp, err := http.Post("http://"+serversList[counter]+textUrl, "text/plain", bytes.NewBuffer(textbytes))
		if err != nil {
			log.Fatal(err)
		}
		textbytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		w.Write([]byte("Адрес сервера: " + serversList[counter] + "\n"))
		w.Write(textbytes)

		if counter < len(serversList)-1 {
			counter++
		} else {
			counter = 0
		}

	})
	http.ListenAndServe(*srvProxy, nil)

}
