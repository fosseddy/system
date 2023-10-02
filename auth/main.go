package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"io"
	"strings"
	"time"
	"errors"
	"encoding/json"

	_ "github.com/golang-jwt/jwt/v5"

	_ "github.com/go-sql-driver/mysql"
)

type context struct {
	db *sql.DB
}

/*
response {
	data: {} || [{}, {}],
	error: {},
}
data: {
	items: []
	items_count: 0
	items_per_page: 0
	items_total: 0
	page_index: 0
	page_total: 0
}

error: {
	code: 400
	errors: [{message: "something"}]
}
*/

type apiResponse struct {
	Data any `json:"data,omitempty"`
	DataMany any `json:"data,omitempty"`
	Error apiError `json:"error,omitempty"`
}

type apiError struct {
	Code int `json:"code"`
	Errors []apiErrorValue `json:"errors"`
}

type apiErrorValue struct {
	Message string `json:"message"`
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	var res apiResponse

	res.Error.Code = status
	res.Error.Errors = append(res.Error.Errors, apiErrorValue{msg})

	b, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(status)
	w.Write(b)
}

func writeServerErr(w http.ResponseWriter, err error) {
	log.Print(err)
	writeErr(w, http.StatusInternalServerError, "internal server error")
}

func (ctx context) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "invalid method")
		w.Header().Set("allow", http.MethodPost)
		return
	}

	if r.Header.Get("content-type") != "application/json" {
		writeErr(w, http.StatusUnsupportedMediaType, "invalid content type")
		return
	}

	w.Header().Set("content-type", "application/json")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		writeServerErr(w, err)
		return
	}

	body := struct{
		Username string
		Password string
	}{}

	if err := json.Unmarshal(b, &body); err != nil {
		var umarsherr *json.InvalidUnmarshalError
		if errors.As(err, &umarsherr) {
			writeServerErr(w, err)
			return
		}

		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}

	body.Username = strings.TrimSpace(body.Username)
	body.Password = strings.TrimSpace(body.Password)

	if body.Username == "" {
		writeErr(w, http.StatusBadRequest, "username is required")
		return
	}

	if body.Password == "" {
		writeErr(w, http.StatusBadRequest, "password is required")
		return
	}

	fmt.Println(body)
	fmt.Fprintln(w, body)
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
	log.SetOutput(io.MultiWriter(f, os.Stderr))
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
