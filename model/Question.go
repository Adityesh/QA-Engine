package model

import (
	"time"
	
)

type Answer struct {
	UserID string `json:"userid" bson:"userid`
	Username string `json:"username" bson:"username"`
	Answer string `json:"answer" bson:"answer"`
	ISSelected bool `json:"isselected" bson:"isselected"`
	DatePosted time.Time `json:"dateposted" bson:"dateposted"`
	Votes int `json:"votes" bson:"votes"`
	Email string `json:"email" bson:"email"`
}

type Question struct {
	Username string `json:"username" bson:"username"`
	Title string `json:"title" bson:"title"`
	Content string `json:"content" bson:"content"`
	Answers []Answer `json:"answers" bson:"answers"`
	SelectedAnswer Answer `json:"selectedanswer" bson:"selectedanswer"`
	Votes int `json:"votes" bson:"votes"`
}

