package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
	respondWithJSON(w, http.StatusOK, map[string]bool{"alive": true})
}

func (a *App) GetTodoItems(w http.ResponseWriter, r *http.Request) {
	var todoItems []TodoItem
	a.DB.Find(&todoItems)
	respondWithJSON(w, http.StatusOK, todoItems)
}

func (a *App) GetTodoItem(w http.ResponseWriter, r *http.Request) {
	var todoItem TodoItem
	vars := mux.Vars(r)
	todoItemID := vars["id"]
	if result := a.DB.First(&todoItem, todoItemID); result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Item not found")
	} else {
		respondWithJSON(w, http.StatusOK, todoItem)
	}
}

func (a *App) CreateTodoItem(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var todoItem TodoItem
	err := decoder.Decode(&todoItem)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad arguments")
		return
	}
	defer r.Body.Close()
	// Check if title is empty
	if todoItem.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is empty")
	}
	a.DB.Create(&todoItem)
	respondWithJSON(w, http.StatusCreated, todoItem)

}

func (a *App) UpdateTodoItem(w http.ResponseWriter, r *http.Request) {
	var todoItem TodoItem
	vars := mux.Vars(r)
	todoItemID := vars["id"]
	if result := a.DB.First(&todoItem, todoItemID); result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Item not found")
	} else {
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&todoItem)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Bad arguments")
			return
		}
		defer r.Body.Close()
		// Check if title is empty
		if todoItem.Title == "" {
			respondWithError(w, http.StatusBadRequest, "Title is empty")
			return
		}
		a.DB.Save(&todoItem)
		respondWithJSON(w, http.StatusOK, todoItem)
	}
}

func (a *App) DeleteTodoItem(w http.ResponseWriter, r *http.Request) {
	var todoItem TodoItem
	vars := mux.Vars(r)
	todoItemID := vars["id"]
	if result := a.DB.First(&todoItem, todoItemID); result.Error != nil {
		respondWithError(w, http.StatusNotFound, "Item not found")
	} else {
		a.DB.Delete(&todoItem)
		respondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	a := App{}
	a.Initialize()
	a.Run(":8000")
}
