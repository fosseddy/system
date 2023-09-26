package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	_ "github.com/go-sql-driver/mysql"
)

type context struct {
	db *sql.DB
}

func (ctx context) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("allow", http.MethodPost)
		return
	}

	if r.Header.Get("content-type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "application/json")

	fmt.Fprintln(w, "hello, from login")
}

func (ctx context) check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello, from check")
}

func (ctx context) refresh(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello, from refresh")
}

func connectDatabase() *sql.DB {
	dsn := fmt.Sprintf("%s:%s@/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return db
}

func loadenv() {
	src, err := os.ReadFile(".env")
	if err != nil {
		panic(err)
	}

	valid := true
	for i, line := range strings.Split(string(src), "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		kv := strings.Split(line, "=")
		if len(kv) != 2 {
			fmt.Printf(".env:%d: invalid line\n", i+1)
			valid = false
			continue
		}

		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if len(key) == 0 || len(val) == 0 {
			fmt.Printf(".env:%d: empty key or value\n", i+1)
			valid = false
			continue
		}

		if err := os.Setenv(key, val); err != nil {
			panic(err)
		}
	}

	if !valid {
		os.Exit(1)
	}
}

func init() {
	loadenv()

	f, err := os.OpenFile("auth.log", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)
}

func main() {
	ctx := context{connectDatabase()}

	http.HandleFunc("/login", ctx.login)
	http.HandleFunc("/check", ctx.check)
	http.HandleFunc("/refresh", ctx.refresh)

	port := os.Getenv("PORT")
	fmt.Printf("Server is listening on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
