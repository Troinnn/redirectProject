package main

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type LinksResource struct {
	col *mongo.Collection
}

func (lr LinksResource) Routers() chi.Router {
	r := chi.NewRouter()
	r.Get("/admin/redirects", lr.AdminRedirects)
	r.Get("/admin/redirects/{id}", lr.AdminRedirect)
	r.Post("/admin/redirects", lr.AdminCreateRedirect)
	r.Patch("/admin/redirects/{id}", lr.AdminUpdateRedirect)
	r.Delete("/admin/redirects/{id}", lr.AdminDeleteRedirect)
	r.Get("/redirects", lr.UserRedirect)
	return r
}

func (lr LinksResource) AdminRedirects(w http.ResponseWriter, r *http.Request) {
	cur, err := lr.col.Find(context.TODO(), bson.D{})
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)
		return
	}
	var result []Link
	err = cur.All(context.TODO(), &result)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, result)
}
func (lr LinksResource) AdminRedirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	one := lr.col.FindOne(context.TODO(), bson.D{
		{
			Key:   "id",
			Value: id,
		},
	})
	if one == nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)
		return
	}

	var result Link
	err := one.Decode(&result)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)
		return
	}
	if result.ID == "" {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, nil)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, result)

}
func (lr LinksResource) AdminCreateRedirect(w http.ResponseWriter, r *http.Request) {
	var link Link
	err := json.NewDecoder(r.Body).Decode(&link)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, nil)
		return
	}
	cur, err := lr.col.Find(context.TODO(), bson.D{
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
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)
		return
	}
	// Если в бд нет одинаковых данных, то записываем. если они есть то пропускаем.
	if cur.RemainingBatchLength() != 0 {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)
		return
	}

	// Создание нового ID
	link.ID = primitive.NewObjectID().Hex()
	lr.col.InsertOne(context.TODO(), link)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, link)
}
func (lr LinksResource) AdminUpdateRedirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var link Link
	err := json.NewDecoder(r.Body).Decode(&link)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, nil)

		return
	}
	one := lr.col.FindOne(context.TODO(), bson.D{
		{
			Key:   "id",
			Value: id,
		},
	})
	if one == nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)

		return
	}
	var oldLink Link
	err = one.Decode(&oldLink)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)

		return
	}
	oldLink.HistoryLink = oldLink.ActiveLink
	oldLink.ActiveLink = link.ActiveLink

	_, err = lr.col.UpdateOne(context.TODO(), bson.D{
		{
			Key:   "id",
			Value: id,
		},
	},
		bson.D{
			{"$set", oldLink},
		})
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, oldLink)
}
func (lr LinksResource) AdminDeleteRedirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := lr.col.DeleteOne(context.TODO(), bson.D{
		{
			Key:   "id",
			Value: id,
		},
	})
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, nil)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, nil)
}

func (lr LinksResource) UserRedirect(w http.ResponseWriter, r *http.Request) {
	// localhost:8080/redirect?link=fuck - Query параметр
	queryLink := r.URL.Query().Get("link")
	one := lr.col.FindOne(context.TODO(), bson.D{
		{
			Key:   "active_link",
			Value: queryLink,
		},
	})
	var oneLink Link
	one.Decode(&oneLink)

	two := lr.col.FindOne(context.TODO(), bson.D{
		{
			Key:   "history_link",
			Value: queryLink,
		},
	})
	var twoLink Link
	two.Decode(&twoLink)

	if oneLink.ID != "" {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, nil)
		return
	}
	if twoLink.ID != "" {
		render.Status(r, http.StatusMovedPermanently)
		render.JSON(w, r, nil)
		return
	}

	render.Status(r, http.StatusNotFound)
	render.JSON(w, r, nil)
}
