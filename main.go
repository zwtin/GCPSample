package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Person struct {
	gorm.Model
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var db *gorm.DB

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(1024)
		fileHeader := r.MultipartForm.File["uploaded"][0]
		file, err := fileHeader.Open()
		if err == nil {
			data, err := ioutil.ReadAll(file)
			if err == nil {
				fmt.Fprintln(w, string(data))
			}
		}
		return
	})

	http.HandleFunc("/getMyData", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprint(w, "{\"name\":\"Masato Ikezawa\", \"screen_name\":\"zwtin\", \"bio\":\"Nice to me too\", \"id\":3874, \"method\":\"get\"}")
			return
		} else if r.Method == http.MethodPost {
			fmt.Fprint(w, "{\"name\":\"Masato Ikezawa\", \"screen_name\":\"zwtin\", \"bio\":\"Nice to me too\", \"id\":3874, \"method\":\"post\"}")
			return
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed) // 405
			w.Write([]byte("http method was not found"))
			return
		}
	})

	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			db = DB()
			defer db.Close()
			var people []Person
			db.Find(&people)
			str, _ := json.Marshal(people)
			fmt.Fprintf(w, "%s\n", str)
			return
		} else if r.Method == http.MethodPost {
			db = DB()
			defer r.Body.Close()
			defer db.Close()
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			jsonBytes := ([]byte)(string(body))
			data := new(Person)
			if err := json.Unmarshal(jsonBytes, data); err != nil {
				fmt.Println("JSON Unmarshal error:", err)
				return
			}
			db.Create(&data)
			return
		} else {
			db = DB()
			defer db.Close()
			w.WriteHeader(http.StatusMethodNotAllowed) // 405
			w.Write([]byte("http method was not found"))
			return
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func DB() *gorm.DB {
	var (
		connectionName = os.Getenv("CLOUDSQL_CONNECTION_NAME")
		user           = os.Getenv("CLOUDSQL_USER")
		password       = os.Getenv("CLOUDSQL_PASSWORD")
		socket         = os.Getenv("CLOUDSQL_SOCKET_PREFIX")
		databaseName   = os.Getenv("CLOUDSQL_DATABASE_NAME")
		option         = os.Getenv("CLOUDSQL_OPTION")
	)

	if socket == "" {
		socket = "/cloudsql"
	}
	if databaseName == "" {
		databaseName = "bookshelf"
	}
	if option == "" {
		option = "?charset=utf8&parseTime=True&loc=Local"
	}

	dbURI := fmt.Sprintf("%s:%s@unix(%s/%s)/%s%s", user, password, socket, connectionName, databaseName, option)
	conn, err := gorm.Open("mysql", dbURI)
	if err != nil {
		panic(fmt.Sprintf("DB: %v", err))
	}

	return conn
}
