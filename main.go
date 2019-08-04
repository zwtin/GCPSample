package main

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Person struct {
	gorm.Model
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Image string `json:"image"`
}

var db *gorm.DB
var rs1Letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world")
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
			r.ParseMultipartForm(1024)
			fileHeader := r.MultipartForm.File["uploaded"][0]
			name := r.MultipartForm.Value["dataA"][0]
			age := r.MultipartForm.Value["dataB"][0]
			file, err := fileHeader.Open()
			defer file.Close()

			ctx := context.Background()

			client, err := storage.NewClient(ctx)
			if err != nil {
				log.Fatalf("Failed to create client: %v", err)
			}

			bucketName := "hello-world-243909.appspot.com"
			bucket := client.Bucket(bucketName)

			randString := RandString1(16)

			wc := bucket.Object(randString).NewWriter(ctx)
			if _, err = io.Copy(wc, file); err != nil {
				return
			}
			if err := wc.Close(); err != nil {
				return
			}
			db = DB()
			defer db.Close()

			person := Person{}
			person.Name = name
			person.Age, _ = strconv.Atoi(age)
			person.Image = "https://storage.googleapis.com/hello-world-243909.appspot.com/" + randString
			db.Create(&person)

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

func RandString1(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = rs1Letters[rand.Intn(len(rs1Letters))]
	}
	return string(b)
}
