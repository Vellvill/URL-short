package utils

import (
	"NewOne/internal/models"
	"fmt"
	"net/http"
)

func Check(s *models.Url) {
	resp, err := http.Get(s.Longurl)
	if err != nil {
		s.Status = fmt.Sprintf("Error conntecton, %s", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		s.Status = fmt.Sprintf("Error conntecton, %s", err)
	}
	s.Status = fmt.Sprintf("Online")
}
