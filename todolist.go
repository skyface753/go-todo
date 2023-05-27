package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db, err = gorm.Open(postgres.Open("host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Europe/Berlin"), &gorm.Config{})

type TodoItem struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:255"`
	Description string `gorm:"size:255"`
	Completed   bool   `gorm:"default:false"`
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	log.Info("API Health is OK")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func GetTodoItems(w http.ResponseWriter, r *http.Request) {
	log.Info("GetTodoItems")
	var todoItems []TodoItem
	db.Find(&todoItems)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todoItems)
}

func GetTodoItem(w http.ResponseWriter, r *http.Request) {
	log.Info("GetTodoItem")
	var todoItem TodoItem
	vars := mux.Vars(r)
	todoItemID := vars["id"]
	db.First(&todoItem, todoItemID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todoItem)

}

func CreateTodoItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	title := r.FormValue("title")
	if title == "" {
		log.Error("Title is empty")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"success": false, "error": "Title is empty"}`)
		return
	}
	description := r.FormValue("description")

	completed, error := strconv.ParseBool(r.FormValue("completed"))
	if error != nil {
		completed = false
	}
	log.WithFields(log.Fields{"title": title, "description": description, "completed": completed}).Info("CreateTodoItem")
	todoItem := TodoItem{Title: title, Description: description, Completed: completed}
	db.Create(&todoItem)

	io.WriteString(w, `{"success": true}`)

}

func UpdateTodoItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	title := r.FormValue("title")
	if title == "" {
		log.Error("Title is empty")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"success": false, "error": "Title is empty"}`)
		return
	}
	description := r.FormValue("description")
	completed, _ := strconv.ParseBool(r.FormValue("completed"))
	log.WithFields(log.Fields{"title": title, "description": description, "completed": completed}).Info("UpdateTodoItem")
	var todoItem TodoItem
	vars := mux.Vars(r)
	todoItemID := vars["id"]
	db.First(&todoItem, todoItemID)
	todoItem.Title = title
	todoItem.Description = description
	todoItem.Completed = completed
	db.Save(&todoItem)
	io.WriteString(w, `{"success": true}`)
}

func DeleteTodoItem(w http.ResponseWriter, r *http.Request) {
	log.Info("DeleteTodoItem")
	var todoItem TodoItem
	vars := mux.Vars(r)
	todoItemID := vars["id"]
	db.First(&todoItem, todoItemID)
	db.Delete(&todoItem)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"success": true}`)
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func main() {
	if err != nil {
		log.Fatal(err)
	}
	db.Debug().AutoMigrate(&TodoItem{})
	log.Info("Starting Todolist API server")
	router := mux.NewRouter()
	router.HandleFunc("/healthz", Healthz).Methods("GET")
	router.HandleFunc("/todoitems", GetTodoItems).Methods("GET")
	router.HandleFunc("/todoitems/{id}", GetTodoItem).Methods("GET")
	router.HandleFunc("/todoitems", CreateTodoItem).Methods("POST")
	router.HandleFunc("/todoitems/{id}", UpdateTodoItem).Methods("PUT")
	router.HandleFunc("/todoitems/{id}", DeleteTodoItem).Methods("DELETE")
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", router))
}
