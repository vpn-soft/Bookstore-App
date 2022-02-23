package api_server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Bookstore-App/models"
	"github.com/Bookstore-App/utils"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type api_server struct {
	client *mongo.Client
	dbName string
}

func getApiServer(db string) *api_server {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cFunc := context.WithTimeout(context.Background(), 10*time.Second)
	_ = cFunc
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return &api_server{client: client, dbName: db}
}

func (server *api_server) getDatabase() *mongo.Database {
	return server.client.Database(server.dbName)
}

func (server *api_server) closeClient() error {
	return server.client.Disconnect(context.Background())
}

func (server *api_server) getBook(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	var id = pathParams["bookid"]
	if id == "" {
		log.Println("empty book id found")
		http.Error(w, "empty book id", http.StatusBadRequest)
		return
	}
	bookID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		msg := fmt.Sprintf("invalid book id provided %s ", id)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	var book models.Book
	err = server.getDatabase().Collection(utils.BOOKS).FindOne(r.Context(), primitive.M{"_id": bookID}).Decode(&book)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println("DB DOWN")
		http.Error(w, "some error occured while processing request, try again after some time", http.StatusInternalServerError)
		return
	} else if err != nil && err == mongo.ErrNoDocuments {
		log.Printf("book not found for id %s ", id)
		http.Error(w, fmt.Sprintf("no book found with given id %s", id), http.StatusOK)
		return
	}
	b, err := json.Marshal(book)
	if err != nil {
		msg := fmt.Sprintf("error while marshalling book :%s :: ERROR:%v\n", id, err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		msg := fmt.Sprintf("error while writing data to response book :%s :: ERROR:%v\n", id, err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (server *api_server) getBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, offset, err := getQueryParams(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var books []models.Book
	cursor, err := server.getDatabase().Collection(utils.BOOKS).Find(ctx, primitive.M{}, options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)))
	if err != nil {
		log.Printf("error while fetching books from db : %s", err.Error())
		http.Error(w, "some error occured while processing request, try again after some time", http.StatusInternalServerError)
		return
	}
	err = cursor.All(ctx, &books)
	if err != nil {
		msg := fmt.Sprintf("error while marshalling books from DB :: ERROR:%v\n", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	var response = models.BooksResponse{Books: books, Pagination: models.Pagination{Offset: int64(offset), Limit: int64(limit), Total: int64(len(books))}}
	b, err := json.Marshal(response)
	if err != nil {
		msg := fmt.Sprintf("error while marshalling books :: ERROR:%v\n", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		msg := fmt.Sprintf("error while writing data to response:: ERROR:%v\n", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (server *api_server) createBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var books []models.Book
	var book models.Book
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&book)
	if err != nil {
		err = decoder.Decode(&books)
		if err != nil {
			log.Printf("error while reading request body :: ERROR:%v\n", err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
	} else {
		books = append(books, book)
	}
	var recordsCreated []interface{}
	var recordsConflicted []string
	log.Println(books)
	for _, bk := range books {
		var result *mongo.InsertOneResult
		result, err = server.getDatabase().Collection(utils.BOOKS).InsertOne(ctx, bk)
		if err == nil {
			var ok bool
			bk.ID, ok = result.InsertedID.(primitive.ObjectID)
			if !ok {
				log.Printf("invalid insertID=[%v] returned by mongo", result.InsertedID)
			}
			recordsCreated = append(recordsCreated, bk)
			continue
		} else {
			log.Printf("error while creating books into db : %s", err.Error())
			recordsConflicted = append(recordsConflicted, bk.Name)
		}
	}
	var resMap = make(map[string]interface{})
	resMap["recordsCreated"] = recordsCreated
	resMap["recordsConflicted"] = recordsConflicted
	b, err := json.Marshal(resMap)
	if err != nil {
		msg := fmt.Sprintf("error while marshalling response :: ERROR:%v\n", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		msg := fmt.Sprintf("error while writing data to response:: ERROR:%v\n", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getQueryParams(r *http.Request) (limit, offset int, err error) {
	var limitStr = r.URL.Query().Get(utils.LIMIT)
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return
		}
	} else {
		limit = utils.DefaultLimit
	}
	var skip = r.URL.Query().Get(utils.OFFSET)
	if skip != "" {
		offset, err = strconv.Atoi(skip)
		if err != nil {
			return
		}
	}
	return
}
