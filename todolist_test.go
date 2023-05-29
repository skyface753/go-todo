package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

var a App

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize()
	a.DB.Migrator().DropTable(&TodoItem{})
	a.DB.Migrator().AutoMigrate(&TodoItem{})
	m.Run()
}

// TestHealthz tests the /healthz endpoint
func TestHealthz(t *testing.T) {
	req, _ := http.NewRequest("GET", "/healthz", nil)
	response := executeRequest(req, a.Router)
	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	checkResponseKeyBool(t, true, m["alive"].(bool))
}

// TestGetNonExistentTodoItem tests the /todoitems/{id} endpoint
func TestGetNonExistentTodoItem(t *testing.T) {
	req, _ := http.NewRequest("GET", "/todoitems/11", nil)
	response := executeRequest(req, a.Router)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	checkResponseKeyString(t, "Item not found", m["error"])
}

// TestCreateTodoItem tests the /todoitems endpoint
func TestCreateTodoItem(t *testing.T) {
	payload := []byte(`{"title":"test title","description":"test description","completed":false}`)
	req, _ := http.NewRequest("POST", "/todoitems", bytes.NewBuffer(payload))
	response := executeRequest(req, a.Router)
	checkResponseCode(t, http.StatusCreated, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	checkResponseKeyString(t, "test title", m["Title"].(string))
	checkResponseKeyString(t, "test description", m["Description"].(string))
	checkResponseKeyBool(t, false, m["Completed"].(bool))
}

// TestGetTodoItem tests the /todoitems/{id} endpoint
func TestGetTodoItem(t *testing.T) {
	req, _ := http.NewRequest("GET", "/todoitems/1", nil)
	response := executeRequest(req, a.Router)
	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	checkResponseKeyString(t, "test title", m["Title"].(string))
	checkResponseKeyString(t, "test description", m["Description"].(string))
	checkResponseKeyBool(t, false, m["Completed"].(bool))
}

func TestGetTodoItems(t *testing.T) {
	req, _ := http.NewRequest("GET", "/todoitems", nil)
	response := executeRequest(req, a.Router)
	checkResponseCode(t, http.StatusOK, response.Code)
}

// TestUpdateTodoItem tests the /todoitems/{id} endpoint
func TestUpdateTodoItem(t *testing.T) {
	payload := []byte(`{"title":"test title updated","description":"test description updated","completed":true}`)
	req, _ := http.NewRequest("PUT", "/todoitems/1", bytes.NewBuffer(payload))
	response := executeRequest(req, a.Router)
	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	checkResponseKeyString(t, "test title updated", m["Title"].(string))
	checkResponseKeyString(t, "test description updated", m["Description"].(string))
	checkResponseKeyBool(t, true, m["Completed"].(bool))
}

// TestUpdateNonExistentTodoItem tests the /todoitems/{id} endpoint
func TestUpdateNonExistentTodoItem(t *testing.T) {
	payload := []byte(`{"title":"test title updated","description":"test description updated","completed":true}`)
	req, _ := http.NewRequest("PUT", "/todoitems/2", bytes.NewBuffer(payload))
	response := executeRequest(req, a.Router)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	checkResponseKeyString(t, "Item not found", m["error"])
}

// TestDeleteTodoItem tests the /todoitems/{id} endpoint
func TestDeleteTodoItem(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/todoitems/1", nil)
	response := executeRequest(req, a.Router)
	checkResponseCode(t, http.StatusOK, response.Code)
}

// TestDeleteNonExistentTodoItem tests the /todoitems/{id} endpoint
func TestDeleteNonExistentTodoItem(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/todoitems/1", nil)
	response := executeRequest(req, a.Router)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	checkResponseKeyString(t, "Item not found", m["error"])
}

func executeRequest(req *http.Request, router *mux.Router) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

/*
 * CheckResponseKey for string and bool
 */
func checkResponseKeyString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected response key %s. Got %s\n", expected, actual)
	}
}

func checkResponseKeyBool(t *testing.T, expected, actual bool) {
	if expected != actual {
		t.Errorf("Expected response key %t. Got %t\n", expected, actual)
	}
}
