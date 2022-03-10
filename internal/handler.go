package service

import (
	"NewOne/internal/models"
	"NewOne/internal/utils"
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

	f := models.NewModelURL(0, keys[0], "", 0, "")
	if err := utils.Check(f); err != nil {

	}

	err := i.repo.AddLink(context.TODO(), f)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("your short url = %s", f.Shorturl)))
}

func (i *Implementation) RedirectToUrl(w http.ResponseWriter, r *http.Request) {
	shorturl := strings.Trim(r.URL.Path, "/")
	link, err := i.repo.GetLink(context.TODO(), shorturl)
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, link, 301) //
}

func (i *Implementation) CheckStats(w http.ResponseWriter, r *http.Request) {
	shorturl := strings.Trim(strings.Trim(r.URL.Path, "/getstats"), "/")
	stats := models.NewModelURL(0, "", shorturl, 0, "")
	i.repo.GetStats(context.TODO(), stats)
	w.Write([]byte(fmt.Sprintf("Ссылка:%s\nКоличество переходов:%d\n", stats.Longurl, stats.Numbersofredirect)))
}

func (i *Implementation) CheckStatus(w http.ResponseWriter, r *http.Request) {
	array, err := i.repo.FindAll(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range array {
		w.Write([]byte(fmt.Sprintf("ID: %d, Long url: %s, Short url: %s, Numbers of redirect: %d, Status: %s\n\n", v.ID, v.Longurl, v.Shorturl, v.Numbersofredirect, v.Status)))
	}
}
