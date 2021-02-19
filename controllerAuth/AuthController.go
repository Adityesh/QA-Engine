package controllerAuth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	// "os"
	"time"

	"example.org/model"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)


type Result struct {
	Err bool `json:"error"`
	Message string `json:"message"`
}


func UserRegisterController(response http.ResponseWriter, request *http.Request, QAEngineDatabase *mongo.Database)  {
	response.Header().Add("Content-Type", "application/json")
	
	user := model.UserModel{}
	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(response).Encode(Result{
			Err : true,
			Message : "Error parsing data provided",
		})
		return
	}

	// Before generating hash check if the username and email combination is found in the database
	err1 := checkEmailInDatabase(QAEngineDatabase, user.Email)
	err2 := checkUsernameInDatabase(QAEngineDatabase, user.Username)

	if err1 == nil {
		// User with that email found
		response.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(response).Encode(Result{
			Err : true,
			Message : "Email already taken",
		})
		return
	}

	if err2 == nil {
		// User with that username found
		response.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(response).Encode(Result{
			Err : true,
			Message : "Username already taken",
		})
		return
	}
	var hash string
	hash, err = generateHashPassword(user.Password,&user)

	if err != nil {
		// Handle the error to the response of the request
	}
	// else Add the new user to the database

	_, err = addUserToDatabase(&user, hash, QAEngineDatabase)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		responseResult := Result{
			Err : true,
			Message : "Failed to add user to the database",
		}
		json.NewEncoder(response).Encode(responseResult)
	} else {
		response.WriteHeader(http.StatusOK)
		responseResult := Result{
			Err : false,
			Message : "User added to the database",
		}
		json.NewEncoder(response).Encode(responseResult)
	}

	defer request.Body.Close()
}

func UserLoginController(response http.ResponseWriter, request *http.Request, QAEngineDatabase *mongo.Database)  {
	// 1. Get the login credentials from the request body
	// 2. Check if the given email or username is there in the database or not
	// 3. Check the given password against the one in the database
	
	jwtKey := []byte("my_secret_key")
	// secret := "suifhapqihopakn;OKR01"
	// Step 1
	var loginCreds model.UserLogin

	json.NewDecoder(request.Body).Decode(&loginCreds)
	var err error
	// Step 2.
	if loginCreds.Username == "" {
		// Only email was provided
		err = checkEmailInDatabase(QAEngineDatabase, loginCreds.Email)
	} else if loginCreds.Email == "" {
		// Only Username was provided
		err = checkUsernameInDatabase(QAEngineDatabase, loginCreds.Username)
	} 

	if err != nil {
		// Means the document with the given email and username was not
		// found in the database
		response.WriteHeader(http.StatusBadRequest);
		json.NewEncoder(response).Encode(Result{
			Err : true,
			Message : "Invalid Credentials",
		})
		return

	} else {
		// User name and email found in the database, Safe to login the user
		// Step 3 Check the given password in the database
		username, email, err := comparePassword(QAEngineDatabase, &loginCreds)

		if err != nil {
			// Error comparing or finding the user
			response.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(response).Encode(Result{
				Err : true,
				Message : "Invalid password",
			})
			return
		} else {
			// Password valid
			expirationTime := time.Now().Add(20 * time.Minute)
			claims := &model.Claims{
				Username:  username,
				Email: email,
				StandardClaims: jwt.StandardClaims{
					ExpiresAt : expirationTime.Unix(),
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

			tokenString, err := token.SignedString(jwtKey)

			if err != nil {
				
				response.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(response).Encode(Result{
					Err : true,
					Message : "Error sigining token",
				})
				return
			} else {
				
				http.SetCookie(response, &http.Cookie{
					Name : "token",
					Value : tokenString,
					Expires : expirationTime,
				})
				response.WriteHeader(http.StatusOK)
				json.NewEncoder(response).Encode(Result{
					Err : false,
					Message : "A-OK",
				})
				return
			}
		}
	}

	
}


func generateHashPassword(password string, user *model.UserModel) (string, error){
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)

	if err != nil {
		return "", errors.New("Failed to hash the password") 
	}
	return string(hash), nil
}

func addUserToDatabase(user *model.UserModel, hashedPass string, QAEngineDatabase *mongo.Database) (*mongo.InsertOneResult, error) {
	// Check if the given username and email is already there in the database or not
	count, err := QAEngineDatabase.Collection("users").CountDocuments(context.TODO(), bson.M{"email" : user.Email, "username" : user.Username})
	if count > 0 {
		// User already exists
		return nil, errors.New("User already exists in the database")
	}

	if err != nil {
		return nil, errors.New("Internal Server Error")
	}

	user.Password = hashedPass
	// User not found , insert a new user in the database
	result, err := QAEngineDatabase.Collection("users").InsertOne(context.TODO(), user)
	if err != nil {
		return nil, errors.New("Internal Server Error")
	} else {
		return result, nil
	}
}

func checkEmailInDatabase(QAEngineDatabase *mongo.Database, email string) error {
	result := QAEngineDatabase.Collection("users").FindOne(context.TODO(), bson.M{
		"email" : email,
	})

	if result.Err() == mongo.ErrNoDocuments {
		// No Documents with that email found in the database
		return errors.New("No documents with that email found in the database")
	} else {
		// User found with that email
		return nil
	}
}

func checkUsernameInDatabase(QAEngineDatabase *mongo.Database, username string) error {
	result := QAEngineDatabase.Collection("users").FindOne(context.TODO(), bson.M{
		"username" : username,
	})

	if result.Err() == mongo.ErrNoDocuments {
		// No Documents with that email found in the database
		return errors.New("No documents with that username found in the database")
	} else {
		// User found with that email
		return nil
	}
}

func comparePassword(QAEngineDatabase *mongo.Database, loginCreds *model.UserLogin) (string, string, error) {
	// Find the user with the given credentials
	var user model.UserReturnModel
	var result *mongo.SingleResult
	if loginCreds.Email == "" {
		result = QAEngineDatabase.Collection("users").FindOne(context.TODO(), bson.M{
			"username" : loginCreds.Username,
		})
	} else {
		result = QAEngineDatabase.Collection("users").FindOne(context.TODO(), bson.M{
			"email" : loginCreds.Email,
		})
	}
	

	if result.Err() == mongo.ErrNoDocuments {
		// No documents found with that credentials
		return "", "", errors.New("Invalid credentials")
	} else {
		result.Decode(&user)
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginCreds.Password))

		if err != nil {
			// Password incorrect
			return "", "", errors.New("Incorrect password")
		} else {
			return user.Username, user.Email, nil
		}
	}

}
