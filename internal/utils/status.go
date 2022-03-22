package utils

import (
	"NewOne/internal/models"
	"fmt"
	"net/http"
)

func Check(s *models.Url) error {
	resp, err := http.Get(s.Longurl)
	if err != nil {
		s.Status = fmt.Sprintf("Error conntecton, %s", err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if resp.StatusCode != 200 {
		s.Status = fmt.Sprintf("Error conntecton, %s", err)
	}
	s.Status = fmt.Sprintf("Online")
	return err
}
