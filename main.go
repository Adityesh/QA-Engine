package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"example.org/controllerAuth"
	"example.org/controllerQuestion"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var QAEngineDatabase *mongo.Database

func HomeHandlerEndpoint(response http.ResponseWriter, request *http.Request)  {
	fmt.Fprintf(response, "Hello from home page")
}


func main() {
	os.Setenv("TOKEN_SECRET", "[031-024H0ODFHSKDOFASDF]SDFASDFASD")
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	client, e := mongo.Connect(context.TODO(), clientOptions)
	
	if e != nil {
		log.Fatal(e)
	}
	e = client.Ping(context.TODO(), nil)

	if e != nil {
		log.Fatal(e)
	}

	QAEngineDatabase = client.Database("QAEngine")
	
	router := mux.NewRouter()
	router.HandleFunc("/", HomeHandlerEndpoint)

	// Register a user to the database
	router.HandleFunc("/user/register", func(response http.ResponseWriter, request *http.Request) {
		controllerAuth.UserRegisterController(response, request, QAEngineDatabase)
	}).Methods("POST")

	// Login a user to the application
	router.HandleFunc("/user/login", func(rw http.ResponseWriter, r *http.Request) {
		controllerAuth.UserLoginController(rw, r, QAEngineDatabase)
	})

	// Add a new question to the database
	router.HandleFunc("/user/question", func(rw http.ResponseWriter, r *http.Request) {
		controllerQuestion.AddQuestion(rw, r, QAEngineDatabase)
	})

	// Upvote route to add an upvote to the question
	router.HandleFunc("/user/question/vote", func(rw http.ResponseWriter, r *http.Request) {
		controllerQuestion.AddUpVoteToQuestion(rw, r, QAEngineDatabase)
	}).Methods("POST")

	// Add a new answer to an existing question
	router.HandleFunc("/user/question/answer", func(rw http.ResponseWriter, r *http.Request) {
		controllerQuestion.AddAnswer(rw, r, QAEngineDatabase)
	}).Methods("POST")

	// Get All Questions
	router.HandleFunc("/user/questions/all", func(rw http.ResponseWriter, r *http.Request) {
		controllerQuestion.GetAllQuestions(rw, r, QAEngineDatabase)
	}).Methods("GET")

	// Order the questions in the database and return it
	router.HandleFunc("/user/questions/order", func(rw http.ResponseWriter, r *http.Request) {
		controllerQuestion.GetAllQuestionsByOrder(rw, r, QAEngineDatabase)
	})

	http.ListenAndServe(":3000", router)

	defer os.Unsetenv("TOKEN_SECRET")
}