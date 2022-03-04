package models

type Url struct {
	ID                int    `json:"id"`
	Longurl           string `json:"longurl"`
	Shorturl          string `json:"shorturl"`
	Numbersofredirect int    `json:"numbersofredirect"`
}

func NewModelURL(ID int, Longurl, Shorturl string, Numbersofredirect int) *Url {
	return &Url{ID: ID, Longurl: Longurl, Shorturl: Shorturl, Numbersofredirect: Numbersofredirect}
}
