package api_server

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/Bookstore-App/utils"
	"github.com/gorilla/mux"
)

func StartAPIServer(port string) {
	dbName := "test1"
	server := getApiServer(dbName)
	utils.ApplyIndices(server.getDatabase())
	defer server.closeClient()
	http.ListenAndServe(fmt.Sprintf(":%s", port), recoveryMid(handler(server)))
}

func handler(server *api_server) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/books/", server.getBooks).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/books/{bookid}", server.getBook).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/book", server.createBook).Methods(http.MethodPost)
	return router
}

//RecoveryMid function will recover from the panic situation.
//If any fatal error or panic occurs it will recover error.
func recoveryMid(app http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				stack := debug.Stack()
				log.Println(string(stack))
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "unexpected error occured")
			}
		}()
		app.ServeHTTP(w, r)
	}
}
