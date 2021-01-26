package model

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VoteDoc struct {
	Title string    `json:"title" bson:"title"`
	Date  time.Time `json:"upvoteTime" bson:"upvoteTime"`
}

type Votes struct {
	ID primitive.ObjectID `json:"userId" bson:"_id"`
	Username string `json:"username" bson:"username"`
	Email string `json:"email" bson:"email"`
	Upvotes []VoteDoc `json:"upvotes" bson:"upvotes"`
	Downvotes []VoteDoc `json:"downvotes" bson:"downvotes"`
}