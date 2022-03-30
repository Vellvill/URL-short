package service

import (
	"UrlShort/internal/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func (i *Implementation) AddNewUrl(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["url"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("need to add full url"))
		return
	}

	f := models.NewModelURL(0, keys[0], "", "")

	err := i.repo.AddLink(context.TODO(), f)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(http.StatusOK) //
	w.Write([]byte(fmt.Sprintf("your short url = %s", f.Shorturl)))
}

func (i *Implementation) RedirectToUrl(w http.ResponseWriter, r *http.Request) {
	shorturl := strings.Trim(r.URL.Path, "/")
	link, err := i.repo.GetLink(context.TODO(), shorturl)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, link, 301) //
}
