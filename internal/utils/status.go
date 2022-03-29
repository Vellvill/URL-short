package utils

import (
	"NewOne/internal/models"
	"fmt"
	"log"
	"net/http"
)

func Check(s *models.Url) error {
	resp, err := http.Get(s.Longurl)
	if err != nil {
		return err
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
	log.Printf("STATUS SYSTEM: %s status: %s", s.Longurl, s.Status)
	return nil
}
