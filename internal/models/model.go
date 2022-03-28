package models

type Url struct {
	ID       int    `json:"id"`
	Longurl  string `json:"longurl"`
	Shorturl string `json:"shorturl"`
	Status   string `json:"status"`
}

func NewModelURL(ID int, Longurl, Shorturl string, Status string) *Url {
	return &Url{ID: ID, Longurl: Longurl, Shorturl: Shorturl, Status: Status}
}
