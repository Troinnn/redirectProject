package main

type Link struct {
	ID          string `json:"id" bson:"id"`
	ActiveLink  string `json:"active_link" bson:"active_link"`
	HistoryLink string `json:"history_link" bson:"history_link"`
}
