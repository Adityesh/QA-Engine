package controllerQuestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	// "fmt"
	"net/http"
	"time"
    "example.org/middlewares"
	"example.org/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type RequestQuestion struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Title    string `json:"title"`
	Content  string `json:"content"`
}

type UpVoteRequestQuestion struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	VoteUsername string `json:"voteusername"`
	VoteEmail    string `json:"voteemail"`
	VoteType     string `json:"votetype"`
}

type Result struct {
	Err     bool   `json:"error"`
	Message string `json:"message"`
}

type ResultSuccess struct {
	Err     bool             `json:"error"`
	Message string           `json:"message"`
	Data    []model.Question `json:"data"`
}

type AnswerRequestQuestion struct {
	AnswerUsername   string `json:"answerusername"`
	AnswerEmail      string `json:"answeremail"`
	QuestionUsername string `json:"questionusername"`
	QuestionEmail    string `json:"questionemail"`
	Answer           string `json:"answer"`
	Title            string `json:"title"`
}

func AddQuestion(response http.ResponseWriter, request *http.Request, QAEngineDatabase *mongo.Database) {
	err := middlewares.VerifyRequest(response, request)

	// Unauthorized access
	if err != nil {
		json.NewEncoder(response).Encode(Result{
			Err : false,
			Message : err.Error(),
		})
		return
	} 

	// Authorized

	var questionDetails RequestQuestion
	json.NewDecoder(request.Body).Decode(&questionDetails)
	var person model.UserModel
	err = checkUserInDatabase(&questionDetails, QAEngineDatabase, &person)
	if err != nil {
		// User is not present in the database
		json.NewEncoder(response).Encode(Result{Err: false, Message: "User not present in the database"})
	} else {
		// Check if the question with the given title is present or not

		err = checkQuestionInDatabase(&questionDetails, QAEngineDatabase)
		if err == nil {
			// Valid question
			// Add the new question to the database

			_, error := addQuestionToDatabase(&questionDetails, QAEngineDatabase)
			if error == nil {
				// Successfully added question to the database
				json.NewEncoder(response).Encode(Result{Err: false, Message: "Added question successfully"})

			} else {
				// Error adding the question to the database
				json.NewEncoder(response).Encode(Result{Err: true, Message: "Error adding question"})
			}
		} else {
			// Duplicate Question in the database
			json.NewEncoder(response).Encode(Result{Err: true, Message: "Question already present in the database"})
		}
	}

	defer request.Body.Close()
}

func checkUserInDatabase(questionDetails *RequestQuestion, QAEngineDatabase *mongo.Database, person *model.UserModel) error {
	error := QAEngineDatabase.Collection("users").FindOne(context.TODO(), bson.M{
		"email":    questionDetails.Email,
		"username": questionDetails.Username,
	}).Decode(&person)

	if error != nil {
		// Error finding an user
		return errors.New("User not found in that database")
	} else {
		return nil
	}
}

func checkQuestionInDatabase(questionDetails *RequestQuestion, QAEngineDatabase *mongo.Database) error {
	var question model.Question
	result := QAEngineDatabase.Collection("questions").FindOne(context.TODO(), bson.M{
		"title": questionDetails.Title,
	}).Decode(&question)

	if result != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if result == mongo.ErrNoDocuments {
			return nil
		}

	}

	return errors.New("Question with that title already found in the database")
}

func addQuestionToDatabase(questionDetails *RequestQuestion, QAEngineDatabase *mongo.Database) (*mongo.InsertOneResult, error) {
	newQuestion := model.Question{}
	newQuestion.Username = questionDetails.Username
	newQuestion.Content = questionDetails.Content
	newQuestion.Votes = 0
	newQuestion.Answers = []model.Answer{}
	newQuestion.Title = questionDetails.Title
	newQuestion.SelectedAnswer = model.Answer{}

	result, err := QAEngineDatabase.Collection("questions").InsertOne(context.TODO(), newQuestion)

	if err != nil {
		// Error adding question
		return nil, errors.New("Error adding question to the database")
	} else {
		return result, nil
	}

}

func AddAnswer(response http.ResponseWriter, request *http.Request, QAEngineDatabase *mongo.Database) {
	// The body of the request should contain username, email of the person who is answering the question
	// and the title of the question as the title is unique in the database
	// 1. Get the username and email of the person who is posting the answer
	// 2. Get the email and username of the person who has posted the question
	// 3. Get the userid of the person who is posting the answer
	// 4. Get the answer content for the question
	// Steps 1, 2 ,4

	err := middlewares.VerifyRequest(response, request)

	// Unauthorized access
	if err != nil {
		json.NewEncoder(response).Encode(Result{
			Err : false,
			Message : err.Error(),
		})
		return
	} 

	// Authorized
	var answerRequestDetails AnswerRequestQuestion
	json.NewDecoder(request.Body).Decode(&answerRequestDetails)

	// Step 3
	result, err := getUserId(QAEngineDatabase, &answerRequestDetails)

	if err != nil {
		// Error fetching the id of the user
		json.NewEncoder(response).Encode(Result{
			Err:     true,
			Message: err.Error(),
		})
		return
	} else {
		// Id present in result variable
		result, err = addAnswerToDatabase(QAEngineDatabase, &answerRequestDetails, result)

		if err != nil {
			json.NewEncoder(response).Encode(Result{
				Err:     true,
				Message: err.Error(),
			})
			return
		} else {
			json.NewEncoder(response).Encode(Result{
				Err:     false,
				Message: result,
			})
			return
		}
	}

}

func AddUpVoteToQuestion(response http.ResponseWriter, request *http.Request, QAEngineDatabase *mongo.Database) {

	// 1. Get the Username and email of the person who posted the question
	// 2. Get the title and content of the question to be voted
	// 3. Get the username and email of the person who is casting the vote
	// 4. Get the type of the vote (upvote or downvote)
	// 5. Check if the person voting has already voted the question or not
	// 6 Update the vote count in the question document of the collection if he has not voted
	// 7. Add the vote or downvote to the user document of the person who voted if not voted

	// Steps 1, 2, 3

	err := middlewares.VerifyRequest(response, request)

	// Unauthorized access
	if err != nil {
		json.NewEncoder(response).Encode(Result{
			Err : false,
			Message : err.Error(),
		})
		return
	} 

	// Authorized
	var questionDetails UpVoteRequestQuestion

	json.NewDecoder(request.Body).Decode(&questionDetails)

	// Step 4. Checking the type of the vote
	// Add +1 vote to the question vote count
	// Step 5 Check if the vote is already casted by the user
	var voteDoc model.Votes
	err = QAEngineDatabase.Collection("votes").FindOne(context.TODO(), bson.M{
		"username": questionDetails.VoteUsername,
		"email":    questionDetails.VoteEmail,
	}).Decode(&voteDoc)

	// Check if the document was found or not
	if err == mongo.ErrNoDocuments {
		// No document found
		// This will happen one time only as the user hasnt voted for a question before
		// Create a new vote document and insert it
		error := addNewVoteToUserVoteCollection(QAEngineDatabase, &questionDetails, &voteDoc)
		if error != nil {
			json.NewEncoder(response).Encode(Result{Err: true, Message: error.Error()})
			return
		} else {
			json.NewEncoder(response).Encode(Result{Err: false, Message: "Upvote Successful"})
			return
		}
	} else {
		// Check in the upvotes array if the user has cast the vote or not
		if questionDetails.VoteType == "upvote" {
			for _, element := range voteDoc.Upvotes {
				if element.Title == questionDetails.Title {
					// User has already cast the vote
					json.NewEncoder(response).Encode(Result{Err: true, Message: "User already cast the upvote respond to the request "})
					return
				}
			}
		} else if questionDetails.VoteType == "downvote" {
			for _, element := range voteDoc.Downvotes {
				if element.Title == questionDetails.Title {
					// User has already cast the vote
					json.NewEncoder(response).Encode(Result{Err: true, Message: "User already cast the downvote respond to the request "})
					return
				}
			}
		}

		// End of loop user hasnt casted this is a new vote
		// update
		filter := bson.M{
			"username": questionDetails.Username,
			"title":    questionDetails.Title,
			"content":  questionDetails.Content,
		}
		voteCountType := 1

		if questionDetails.VoteType == "upvote" {
			voteCountType = 1
		} else {
			voteCountType = -1
		}

		update := bson.M{
			"$inc": bson.M{
				"votes": voteCountType,
			},
		}

		// Step 6
		result, _ := QAEngineDatabase.Collection("questions").UpdateOne(context.TODO(), filter, update)

		// Check the error and send the response accordingly

		if result.MatchedCount == 0 {
			// No matching documnets with the filter provided
			json.NewEncoder(response).Encode(Result{Err: true, Message: "Failed to vote for the question"})
			return
		} else {
			// Else updated vote count

			// Now add the vote document to user collection who casted the vote
			filter = bson.M{
				"username": questionDetails.VoteUsername,
				"email":    questionDetails.Email,
			}
			var votes []model.VoteDoc
			if questionDetails.VoteType == "upvote" {
				votes = voteDoc.Upvotes
			} else {
				votes = voteDoc.Downvotes
			}

			votes = append(votes, model.VoteDoc{
				Title: questionDetails.Title,
			})

			if questionDetails.VoteType == "upvote" {
				update = bson.M{
					"$set": bson.M{
						"upvotes": votes,
					},
				}
			} else {
				update = bson.M{
					"$set": bson.M{
						"downvotes": votes,
					},
				}
			}
			// Step 7
			_, err = QAEngineDatabase.Collection("votes").UpdateOne(context.TODO(), filter, update)
			if err != nil {
				// Error updating the document
				json.NewEncoder(response).Encode(Result{Err: true, Message: "Error updating the document"})
				return
			} else {
				// Success
				json.NewEncoder(response).Encode(Result{Err: false, Message: "Successfully added the new vote document to the casters collection"})
				return
			}

		}

	}

}

func addNewVoteToUserVoteCollection(QAEngineDatabase *mongo.Database, questionDetails *UpVoteRequestQuestion, voteDoc *model.Votes) error {
	if questionDetails.VoteType == "upvote" {
		for _, element := range voteDoc.Upvotes {
			if element.Title == questionDetails.Title {
				// User has already cast the vote
				return errors.New("User already cast the vote respond to the request ")

			}
		}
	} else if questionDetails.VoteType == "downvote" {
		for _, element := range voteDoc.Downvotes {
			if element.Title == questionDetails.Title {
				// User has already cast the vote
				return errors.New("User already cast the vote respond to the request ")
			}
		}
	}

	filter := bson.M{
		"username": questionDetails.Username,
		"title":    questionDetails.Title,
		"content":  questionDetails.Content,
	}

	voteCountType := 1

	if questionDetails.VoteType == "upvote" {
		voteCountType = 1
	} else {
		voteCountType = -1
	}

	update := bson.M{
		"$inc": bson.M{
			"votes": voteCountType,
		},
	}

	// Step 6
	result, _ := QAEngineDatabase.Collection("questions").UpdateOne(context.TODO(), filter, update)

	if result.MatchedCount == 0 {
		// Error incrementing the vote for the question
		return errors.New("Error incrementing the vote for the question")
	} else {
		var votes []model.VoteDoc
		votes = append(votes, model.VoteDoc{
			Title: questionDetails.Title,
		})

		var newVoteDoc model.Votes

		if questionDetails.VoteType == "upvote" {
			newVoteDoc.Username = questionDetails.VoteUsername
			newVoteDoc.Email = questionDetails.VoteEmail
			newVoteDoc.Upvotes = votes
		} else if questionDetails.VoteType == "downvote" {
			newVoteDoc.Username = questionDetails.VoteUsername
			newVoteDoc.Email = questionDetails.VoteEmail
			newVoteDoc.Downvotes = votes
		}

		_, err := QAEngineDatabase.Collection("votes").InsertOne(context.TODO(), newVoteDoc)
		if err != nil {
			// Error inserting the vote documnet
			return errors.New("Error inserting the new document")
		} else {
			// No error, respond accordingly
			return nil
		}
	}

}

func getUserId(QAEngineDatabase *mongo.Database, answerRequestDetails *AnswerRequestQuestion) (string, error) {
	var answerReturn model.UserReturnModel
	result := QAEngineDatabase.Collection("users").FindOne(context.TODO(), bson.M{
		"username": answerRequestDetails.AnswerUsername,
		"email":    answerRequestDetails.AnswerEmail,
	}).Decode(&answerReturn)

	if result == mongo.ErrNoDocuments {
		return "", errors.New("No document found")
	} else {
		return answerReturn.ID.String(), nil
	}
}

func addAnswerToDatabase(QAEngineDatabase *mongo.Database, answerRequestDetails *AnswerRequestQuestion, userId string) (string, error) {
	// Step 1 Get the user document of the person who posted the question
	// Step 2 Make a new Answer document and add it to the answers array of the user who posted the question
	// Step 3 ??
	// Step 4 Profit
	var answerReturn model.Question
	// Step 1
	result := QAEngineDatabase.Collection("questions").FindOne(context.TODO(), bson.M{
		"username": answerRequestDetails.QuestionUsername,
		"title":    answerRequestDetails.Title,
	}).Decode(&answerReturn)

	if result == mongo.ErrNoDocuments {
		// No document found
		return "", errors.New("No documnet found with the username and email")
	} else {
		// Step 2
		answerModel := model.Answer{
			UserID:     userId,
			Answer:     answerRequestDetails.Answer,
			Username:   answerRequestDetails.AnswerUsername,
			Email:      answerRequestDetails.AnswerEmail,
			ISSelected: false,
			Votes:      0,
			DatePosted: time.Now(),
		}

		answersArr := answerReturn.Answers
		answersArr = append(answersArr, answerModel)
		// Update the array in the database

		filter := bson.M{
			"username": answerRequestDetails.QuestionUsername,
			"title":    answerRequestDetails.Title,
		}

		update := bson.M{
			"$set": bson.M{
				"answers": answersArr,
			},
		}

		result, _ := QAEngineDatabase.Collection("questions").UpdateOne(context.TODO(), filter, update)

		if result.MatchedCount == 0 {
			// Failed to update document
			return "", errors.New("Failed to add the answer to the database")
		} else {
			// Updated successfully
			return "Added the answer to the database", nil
		}
	}
}

// Gets all questions in the order they were posted
func GetAllQuestions(response http.ResponseWriter, request *http.Request, QAEngineDatabase *mongo.Database) {
	var questions []model.Question

	cursor, err := QAEngineDatabase.Collection("questions").Find(context.TODO(), bson.D{})
	defer cursor.Close(context.TODO())
	if err != nil {
		// Error fetching all the questions
		json.NewEncoder(response).Encode(Result{
			Err:     false,
			Message: err.Error(),
		})
		return
	} else {
		for cursor.Next(context.TODO()) {
			var question model.Question
			cursor.Decode(&question)

			questions = append(questions, question)
		}

		json.NewEncoder(response).Encode(ResultSuccess{
			Err:     false,
			Message: "Successfully fetched all questions",
			Data:    questions,
		})
		return
	}

}

func GetAllQuestionsByOrder(response http.ResponseWriter, request *http.Request, QAEngineDatabase *mongo.Database) {
	query := request.URL.Query()

	sort, present := query["sort"]

	if !present || len(sort) == 0 {
		fmt.Println("sort not present")
	} else {

		if sort[0] == "top" {
			var questions []model.Question

			cursor, err := QAEngineDatabase.Collection("questions").Find(context.TODO(), bson.M{
				"$sort": bson.M{
					"votes": 1,
				},
			})
			defer cursor.Close(context.TODO())
			if err != nil {
				// Error fetching all the questions
				json.NewEncoder(response).Encode(Result{
					Err:     false,
					Message: err.Error(),
				})
				return
			} else {
				for cursor.Next(context.TODO()) {
					var question model.Question
					cursor.Decode(&question)

					questions = append(questions, question)
				}

				json.NewEncoder(response).Encode(ResultSuccess{
					Err:     false,
					Message: "Successfully fetched all questions",
					Data:    questions,
				})
				return
			}
		}

	}
}
