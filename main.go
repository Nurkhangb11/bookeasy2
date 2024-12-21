package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

// Структура для данных запроса
type RequestData struct {
	Message string `json:"message"`
}

// Структура для ответа сервера
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Структура для получения данных из базы
type Message struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

var db *sql.DB

func main() {
	// Подключение к базе данных PostgreSQL
	var err error
	db, err = connectToDatabase()
	if err != nil {
		fmt.Println("Ошибка подключения к базе данных:", err)
		return
	}
	defer db.Close()

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		fmt.Println("База данных недоступна:", err)
		return
	}

	// Статические файлы из папки "static"
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Обработчики для операций CRUD
	http.HandleFunc("/contact", handleContact)
	http.HandleFunc("/messages", handleSelectMessages)
	http.HandleFunc("/messages/", handleSelectMessageByID)

	// Запуск сервера
	fmt.Println("Сервер запущен на http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}

// Функция подключения к PostgreSQL
func connectToDatabase() (*sql.DB, error) {
	connStr := "host=localhost port=5432 user=postgres password=password dbname=hotel_booking sslmode=disable"
	return sql.Open("postgres", connStr)
}

// Обработчик для добавления сообщения
func handleContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var requestData RequestData
		err := json.NewDecoder(r.Body).Decode(&requestData)

		if err != nil || requestData.Message == "" {
			http.Error(w, `{"status":"fail","message":"Некорректные данные формы"}`, http.StatusBadRequest)
			return
		}

		_, err = db.Exec("INSERT INTO messages (content) VALUES ($1)", requestData.Message)
		if err != nil {
			http.Error(w, `{"status":"error","message":"Ошибка сохранения данных"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"status":  "success",
			"message": "Данные успешно сохранены",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

// Обработчик для SELECT (все сообщения)
func handleSelectMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		rows, err := db.Query("SELECT id, content FROM messages")
		if err != nil {
			http.Error(w, `{"status":"error","message":"Ошибка получения данных"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var messages []Message
		for rows.Next() {
			var msg Message
			if err := rows.Scan(&msg.ID, &msg.Content); err != nil {
				http.Error(w, `{"status":"error","message":"Ошибка обработки данных"}`, http.StatusInternalServerError)
				return
			}
			messages = append(messages, msg)
		}

		json.NewEncoder(w).Encode(messages)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

// Обработчик для SELECT BY ID
func handleSelectMessageByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		idStr := r.URL.Path[len("/messages/"):] // Получаем ID из URL
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, `{"status":"fail","message":"Некорректный ID"}`, http.StatusBadRequest)
			return
		}

		var msg Message
		err = db.QueryRow("SELECT id, content FROM messages WHERE id = $1", id).Scan(&msg.ID, &msg.Content)
		if err == sql.ErrNoRows {
			http.Error(w, `{"status":"fail","message":"Сообщение не найдено"}`, http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, `{"status":"error","message":"Ошибка получения данных"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(msg)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}
