package main

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"os"
)

/*
1. Открыть соединение с базой данных MongoDB
2. Прочитать файл json
3. Создать ресурс (handler) для нашего сервера
4. Открыть сервер, в этом тебе поможет пакет net/http и функция ListAndServe
*/

const uri = "mongodb://localhost:27017"

func main() {
	// 1. Подключени к базе данных
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	db := client.Database("redirect")
	linksCol := db.Collection("links")
	//

	// 2. Чтение файла json и сохранение в базу данных в коллекцию links
	loadfile(linksCol)
	//

	// 3. Создать ресурс (handler) для нашего сервера
	linksResource := LinksResource{
		col: linksCol,
	}
	//

	// 3. Открытие локального сервера на порте 8080
	http.ListenAndServe(":8080", linksResource.Routers())
	//
}

// loadfile - читает файл json и сохраняет данные в бд.
func loadfile(linksCol *mongo.Collection) {
	// Читаем данные из файла links.json в переменную data (слайс байтов)
	data, err := os.ReadFile("links.json")
	if err != nil {
		return
	}
	// Слайс байтов переводим в слайс структур ([]Link)
	var links []Link
	err = json.Unmarshal(data, &links)
	if err != nil {
		return
	}
	//Проходим по слайсу структур и по ключевому значению сравниваем их с данным в бд
	for _, link := range links {
		cur, err := linksCol.Find(context.TODO(), bson.D{
			{
				Key:   "active_link",
				Value: link.ActiveLink,
			},
			{
				Key:   "history_link",
				Value: link.HistoryLink,
			},
		})
		if err != nil {
			continue
		}
		// Если в бд нет одинаковых данных, то записываем. если они есть то пропускаем.
		if cur.RemainingBatchLength() == 0 {
			// Создание нового ID
			link.ID = primitive.NewObjectID().Hex()
			// Сохраняем в нашу бд данные
			linksCol.InsertOne(context.TODO(), link)
		}
	}
}
