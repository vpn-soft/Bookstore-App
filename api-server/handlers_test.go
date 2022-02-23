package api_server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bookstore-App/models"
	"github.com/Bookstore-App/utils"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetEntries(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/books/", nil)
	if err != nil {
		t.Fatal(err)
	}

	server := getApiServer("test")
	defer server.closeClient()
	collection := server.client.Database("test").Collection(utils.BOOKS)
	var data = models.Book{ID: primitive.NewObjectID(), Name: "test", Author: "test"}
	_, err = collection.InsertOne(req.Context(), data)
	if err != nil {
		t.Error("error while inserting doc")
	}
	vars := map[string]string{
		"bookid": data.ID.Hex(),
	}

	req = mux.SetURLVars(req, vars)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.getBook)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	_, err = collection.DeleteMany(req.Context(), primitive.M{})
	if err != nil {
		t.Error("error while removing doc")
	}
}

func TestGetBooks(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/api/v1/books", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	server := getApiServer("test")
	rec := httptest.NewRecorder()
	server.getBooks(rec, req)
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected Status OK, recieved status %d", res.StatusCode)
	}
}

func TestCreateBooks(t *testing.T) {
	var payload = models.Book{Name: "ttestabc", Author: "test"}
	data, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, "/api/v1/book", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	server := getApiServer("test")
	rec := httptest.NewRecorder()
	server.createBook(rec, req)
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected Status OK, recieved status %d", res.StatusCode)
	}
}
