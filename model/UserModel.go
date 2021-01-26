package model

import (
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserModel struct {

	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Email string `json:"email" bson:"email"`
	Country string `json:"country" bson:"country"`
	Phone int64 `json:"phone" bson:"phone"`
	City string `json:"city" bson:"city"`
}

type UserReturnModel struct {
	ID primitive.ObjectID `json:"userid" bson:"_id"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Email string `json:"email" bson:"email"`
	Country string `json:"country" bson:"country"`
	Phone int64 `json:"phone" bson:"phone"`
	City string `json:"city" bson:"city"`
}

type UserLogin struct {
	Email string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`

}

type Claims struct {
	Username string `json:"username"`
	Email string `json:"email"`
	jwt.StandardClaims
}