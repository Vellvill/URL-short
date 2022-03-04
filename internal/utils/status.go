package utils

import (
	"fmt"
	"net/http"
)

func Check(s chan string, url string) {
	resp, err := http.Get(url)
	if err != nil {
		s <- fmt.Sprintf("Ошибка соединения. %s\n", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		s <- fmt.Sprintf("Ошибка, http-status: %s\n", resp.StatusCode)
	}
	s <- fmt.Sprintf("Онлайн. http-статус: %d\n", resp.StatusCode)
}
