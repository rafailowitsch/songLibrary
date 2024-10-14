package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// SongDetail структура для ответа на запрос
type SongDetail struct {
	Name        string `json:"name"`
	Group       string `json:"group"`
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

// Генерация случайного текста
func generateRandomText() string {
	texts := []string{
		"Ooh baby, don't you know I suffer?",
		"Ooh baby, can you hear me moan?",
		"You caught me under false pretenses.",
		"How long before you let me go?",
	}
	return texts[rand.Intn(len(texts))]
}

// Генерация случайной ссылки
func generateRandomLink() string {
	links := []string{
		"https://example.com/song1",
		"https://example.com/song2",
		"https://example.com/song3",
		"https://example.com/song4",
	}
	return links[rand.Intn(len(links))]
}

// Генерация случайной даты релиза
func generateRandomReleaseDate() string {
	now := time.Now()
	randomDays := rand.Intn(365 * 10) // Случайное количество дней за последние 10 лет
	randomDate := now.AddDate(0, 0, -randomDays)
	return randomDate.Format("02.01.2006")
}

// Обработчик для получения информации о песне
func songInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Получение параметров запроса
	group := r.URL.Query().Get("group")
	name := r.URL.Query().Get("song")

	// Проверка, что параметры присутствуют
	if group == "" || name == "" {
		log.Println("Missing group or name")
		http.Error(w, "Missing group or name", http.StatusBadRequest)
		return
	}

	// Вероятность 30% на возврат ошибки
	if rand.Float64() < 0.3 {
		log.Println("Bad request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Генерация случайных данных
	songDetail := SongDetail{
		Name:        name,
		Group:       group,
		ReleaseDate: generateRandomReleaseDate(),
		Text:        generateRandomText(),
		Link:        generateRandomLink(),
	}

	// Установка заголовков и возврат данных
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songDetail)
}

func main() {
	// Маршруты
	http.HandleFunc("/info", songInfoHandler)

	// Запуск сервера
	log.Println("Mock Music Info Server started on :8088")
	log.Fatal(http.ListenAndServe("localhost:8088", nil))
}
