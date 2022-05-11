package main

// Link - структура хранения наши ссылок.
type Link struct {
	ID          string `json:"id" bson:"id"`                     // Идентификатор ссылки
	ActiveLink  string `json:"active_link" bson:"active_link"`   // Текущая (активная ссылка)
	HistoryLink string `json:"history_link" bson:"history_link"` // Предыдущая ссылка
}
