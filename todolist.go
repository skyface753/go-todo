package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// var db, err = gorm.Open(postgres.Open("host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Europe/Berlin"), &gorm.Config{})

type App struct {
	Router *mux.Router
	DB     *gorm.DB
}

func (a *App) Initialize() {
	var connectionString = "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Europe/Berlin"
	var err error

	a.DB, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		panic("failed to connect database")
	}
	a.DB.AutoMigrate(&TodoItem{})
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/healthz", a.Healthz).Methods("GET")
	a.Router.HandleFunc("/todoitems", a.GetTodoItems).Methods("GET")
	a.Router.HandleFunc("/todoitems/{id}", a.GetTodoItem).Methods("GET")
	a.Router.HandleFunc("/todoitems", a.CreateTodoItem).Methods("POST")
	a.Router.HandleFunc("/todoitems/{id}", a.UpdateTodoItem).Methods("PUT")
	a.Router.HandleFunc("/todoitems/{id}", a.DeleteTodoItem).Methods("DELETE")
}

func (a *App) Run(addr string) {
	log.Info("Starting Todolist API server")
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

type TodoItem struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:255"`
	Description string `gorm:"size:255"`
	Completed   bool   `gorm:"default:false"`
}

func (a *App) Healthz(w http.ResponseWriter, r *http.Request) {
	log.Info("API Health is OK")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func (a *App) GetTodoItems(w http.ResponseWriter, r *http.Request) {
	log.Info("GetTodoItems")
	var todoItems []TodoItem
	a.DB.Find(&todoItems)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todoItems)
}

func (a *App) GetTodoItem(w http.ResponseWriter, r *http.Request) {
	log.Info("GetTodoItem")
	var todoItem TodoItem
	vars := mux.Vars(r)
	todoItemID := vars["id"]
	if result := a.DB.First(&todoItem, todoItemID); result.Error != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"success": false, "error": "Item not found"}`)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todoItem)
	}

}

func (a *App) CreateTodoItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var todoItem TodoItem
	err := decoder.Decode(&todoItem)
	if err != nil {
		log.Error("Error decoding JSON")
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"success": false, "error": "Bad arguments"}`)
		return
	}
	defer r.Body.Close()
	// Check if title is empty
	if todoItem.Title == "" {
		log.Error("Title is empty")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"success": false, "error": "Title is empty"}`)
		return
	}
	a.DB.Create(&todoItem)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todoItem)

}

func (a *App) UpdateTodoItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var todoItem TodoItem
	vars := mux.Vars(r)
	todoItemID := vars["id"]
	if result := a.DB.First(&todoItem, todoItemID); result.Error != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"success": false, "error": "Item not found"}`)
	} else {
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&todoItem)
		if err != nil {
			log.Error("Error decoding JSON")
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, `{"success": false, "error": "Bad arguments"}`)
			return
		}
		defer r.Body.Close()
		// Check if title is empty
		if todoItem.Title == "" {
			log.Error("Title is empty")
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, `{"success": false, "error": "Title is empty"}`)
			return
		}
		a.DB.Save(&todoItem)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(todoItem)
	}
}

func (a *App) DeleteTodoItem(w http.ResponseWriter, r *http.Request) {
	var todoItem TodoItem
	vars := mux.Vars(r)
	todoItemID := vars["id"]
	if result := a.DB.First(&todoItem, todoItemID); result.Error != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"success": false, "error": "Item not found"}`)
	} else {
		a.DB.Delete(&todoItem)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"success": true}`)
	}
}

func main() {
	a := App{}
	a.Initialize()
	a.Run(":8000")
}
