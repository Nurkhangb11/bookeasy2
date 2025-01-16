package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"

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

// Структура для регистрации пользователя
type User struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
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
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/profile", handleProfile)
	http.HandleFunc("/send-support-message", handleSendSupportMessage)

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

func handleRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil || user.FirstName == "" || user.LastName == "" || user.Email == "" || user.Password == "" {
			http.Error(w, `{"status":"fail","message":"Некорректные данные формы"}`, http.StatusBadRequest)
			return
		}

		// Сохраняем пользователя в базе данных без хеширования пароля
		_, err = db.Exec(`INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4)`, user.FirstName, user.LastName, user.Email, user.Password)
		if err != nil {
			log.Printf("Ошибка SQL: %v", err)
			http.Error(w, `{"status":"error","message":"Ошибка сохранения данных в базе"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"status":  "success",
			"message": "Пользователь успешно зарегистрирован",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

// Структура для данных запроса
type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil || user.Email == "" || user.Password == "" {
			http.Error(w, `{"status":"fail","message":"Некорректные данные формы"}`, http.StatusBadRequest)
			return
		}

		// Проверка пользователя в базе данных без хеширования пароля
		var storedPassword string
		err = db.QueryRow(`SELECT password FROM users WHERE email = $1`, user.Email).Scan(&storedPassword)
		if err != nil {
			http.Error(w, `{"status":"fail","message":"Пользователь не найден"}`, http.StatusUnauthorized)
			return
		}

		// Сравнение пароля с сохраненным паролем
		if storedPassword != user.Password {
			http.Error(w, `{"status":"fail","message":"Неверный пароль"}`, http.StatusUnauthorized)
			return
		}

		// Успешный логин
		response := map[string]string{
			"status":  "success",
			"message": "Вы успешно вошли",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

// Функция для получения данных пользователя
func handleProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Проверяем, что метод POST
	if r.Method == http.MethodGet {
		// Получаем email из заголовка запроса или из сессии (зависит от реализации аутентификации)
		// Для простоты примем, что email передается в заголовке запроса
		email := r.URL.Query().Get("email")
		if email == "" {
			http.Error(w, `{"status":"fail","message":"Email не указан"}`, http.StatusBadRequest)
			return
		}

		// Получаем данные пользователя из базы данных
		var user User
		err := db.QueryRow(`SELECT first_name, last_name, email FROM users WHERE email = $1`, email).Scan(&user.FirstName, &user.LastName, &user.Email)
		if err != nil {
			http.Error(w, `{"status":"fail","message":"Пользователь не найден"}`, http.StatusNotFound)
			return
		}

		// Отправляем данные пользователя в ответ
		json.NewEncoder(w).Encode(user)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

// Обработчик для отправки сообщений
func handleSendSupportMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseMultipartForm(10 << 20) // Устанавливаем лимит на размер формы (10 MB)

		// Получаем данные из формы
		email := r.FormValue("email")
		message := r.FormValue("message")
		attachment, _, err := r.FormFile("attachment")

		if err != nil && attachment != nil {
			http.Error(w, "Failed to read attachment", http.StatusInternalServerError)
			return
		}

		// Если файл прикреплен, сохраняем его
		if attachment != nil {
			file, err := ioutil.TempFile("./uploads", "attachment-*.jpg")
			if err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			_, err = io.Copy(file, attachment)
			if err != nil {
				http.Error(w, "Failed to copy attachment", http.StatusInternalServerError)
				return
			}
		}

		// Создаем письмо
		subject := "Support Request"
		body := fmt.Sprintf("Email: %s\nMessage: %s", email, message)

		// Настройка SMTP
		smtpHost := "smtp.mail.ru"
		smtpPort := "587"
		from := "bookeasy_help@mail.ru"
		password := "L1sFSSHs1qax2Cy3ssxN"
		to := []string{"erme.shoinov@bk.ru"}

		auth := smtp.PlainAuth("erme.shoinov@bk.ru", from, password, smtpHost)
		msg := []byte("To: \r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" + body + "\r\n")

		// Отправка письма
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
		if err != nil {
			http.Error(w, "Failed to send email", http.StatusInternalServerError)
			return
		}

		// Ответ на запрос
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "success", "message": "Message sent successfully"}`)
		return
	}

	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}
