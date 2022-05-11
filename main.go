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

// uri - строка константа для подключения к MongoDB
const uri = "mongodb://localhost:27017"

func main() {
	// 1. Подключени к базе данных
	// 1.1 Здесь с помощью функции Connect из пакета mongo, производится подключение к базе MongoDB.
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	// 1.2.1 Если подключение не удалось, то переманная err не будет nil
	// 1.2.2 Если подключение успешно, тогда переманная err будет nil
	if err != nil {
		panic(err)
	}
	// 1.3 Здесь мы получаем доступ к нашей базе данных redirect.
	db := client.Database("redirect")

	// 1.4 Здесь мы получаем объект ответственный за коллекцию links в базе данных redirect.
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

	// 4. Открытие локального сервера на порте 8080
	http.ListenAndServe(":8080", linksResource.Routers())
	//
}

// loadfile - читает файл json и сохраняет данные в бд.
func loadfile(linksCol *mongo.Collection) {
	// 1. Читаем данные из файла links.json в переменную data (слайс байтов).
	data, err := os.ReadFile("links.json")
	/*
		1.1 Если чтение не получилось(файл не корректный или его не существует) тогда err не будет равна nil.
			Если чтение успешное, тогда err будет nil.
	*/
	if err != nil {
		panic(err)
	}

	// 2. Слайс байтов переводим в слайс структур ([]Link).
	var links []Link
	// 2.1 Функция Unmarshal из пакета json используется для конвертации слайса байтов в объект (структуру).
	err = json.Unmarshal(data, &links)
	/*
		2.2 Если конвертация не успешна, то err не будет равна nil.
			  Если конвертация успешна, то err будет равна nil.
	*/
	if err != nil {
		panic(err)
	}
	// 3. Проходим по слайсу структур и по ключевому значению сравниваем их с данным в бд
	for _, link := range links {
		// 3.1 Запрашиваем из базы по полям active_link и history_link.
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
		// 3.2 Если не удалось считать данные с базы данных (оборвалось соединение), то err не будет равна nil.
		if err != nil {
			// 3.2.1 Если произошла ошибка то лучшим вариантом будет продолжить выполнение следующих элементов из слайса.
			continue
		}
		// 3.3 Если в бд нет одинаковых данных, то записываем, если они есть то пропускаем.
		if cur.RemainingBatchLength() == 0 {
			// 3.3.1 Создание нового ID
			link.ID = primitive.NewObjectID().Hex()
			// 3.3.2 Сохраняем в нашу бд данные
			linksCol.InsertOne(context.TODO(), link)
		}
	}
}
