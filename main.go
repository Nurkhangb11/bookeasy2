package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"time"

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
	http.HandleFunc("/send-chat-message", handleSendMessage)
	http.HandleFunc("/messages", handleSelectMessages)
	http.HandleFunc("/clear-messages", handleClearMessages)
	http.HandleFunc("/confirm", handleConfirm)

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

		// Генерация токена
		token := generateToken()

		// Сохранение пользователя с токеном
		_, err = db.Exec(`INSERT INTO users (first_name, last_name, email, password, confirmation_token) VALUES ($1, $2, $3, $4, $5)`,
			user.FirstName, user.LastName, user.Email, user.Password, token)
		if err != nil {
			log.Printf("Ошибка SQL: %v", err)
			http.Error(w, `{"status":"error","message":"Ошибка сохранения данных в базе"}`, http.StatusInternalServerError)
			return
		}

		// Отправка email
		err = sendConfirmationEmail(user.Email, token)
		if err != nil {
			log.Printf("Ошибка отправки email: %v", err)
			http.Error(w, `{"status":"error","message":"Ошибка отправки email"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"status":  "success",
			"message": "Пользователь успешно зарегистрирован. Проверьте email для подтверждения.",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

func generateToken() string {
	return fmt.Sprintf("%x", time.Now().UnixNano()) // Уникальный токен
}

func sendConfirmationEmail(email, token string) error {
	smtpHost := "smtp.mail.ru"
	smtpPort := "587"
	from := "bookeasy_help@mail.ru"
	password := "L1sFSSHs1qax2Cy3ssxN"
	to := []string{email}

	subject := "Подтверждение регистрации"
	body := fmt.Sprintf("Здравствуйте!\n\nПерейдите по ссылке для подтверждения регистрации:\nhttp://localhost:8080/confirm?token=%s", token)

	auth := smtp.PlainAuth("", from, password, smtpHost)
	msg := []byte("To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")

	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
}

func handleConfirm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, `{"status":"fail","message":"Токен не указан"}`, http.StatusBadRequest)
			return
		}

		// Проверка токена и активация пользователя
		result, err := db.Exec(`UPDATE users SET is_confirmed = TRUE, confirmation_token = NULL WHERE confirmation_token = $1`, token)
		if err != nil {
			log.Printf("Ошибка SQL: %v", err)
			http.Error(w, `{"status":"error","message":"Ошибка базы данных"}`, http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, `{"status":"fail","message":"Некорректный токен"}`, http.StatusBadRequest)
			return
		}

		response := map[string]string{
			"status":  "success",
			"message": "Регистрация подтверждена. Теперь вы можете войти.",
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method == http.MethodPost {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil || user.Email == "" || user.Password == "" {
			http.Error(w, `{"status":"fail","message":"Некорректные данные формы"}`, http.StatusBadRequest)
			return
		}

		var storedPassword string
		err = db.QueryRow(`SELECT password FROM users WHERE email = $1`, user.Email).Scan(&storedPassword)
		if err != nil {
			http.Error(w, `{"status":"fail","message":"Пользователь не найден"}`, http.StatusUnauthorized)
			return
		}

		if storedPassword != user.Password {
			http.Error(w, `{"status":"fail","message":"Неверный пароль"}`, http.StatusUnauthorized)
			return
		}

		var isConfirmed bool
		err = db.QueryRow(`SELECT is_confirmed FROM users WHERE email = $1`, user.Email).Scan(&isConfirmed)
		if err != nil || !isConfirmed {
			http.Error(w, `{"status":"fail","message":"Подтвердите email перед входом"}`, http.StatusForbidden)
			return
		}

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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method == http.MethodGet {
		// Получаем email из параметров URL
		email := r.URL.Query().Get("email")
		if email == "" {
			http.Error(w, `{"status":"fail","message":"Email не указан"}`, http.StatusBadRequest)
			return
		}

		// Получаем данные пользователя из базы данных
		var user User
		err := db.QueryRow(`SELECT first_name, last_name, email FROM users WHERE email = $1`, email).Scan(&user.FirstName, &user.LastName, &user.Email)
		if err != nil {
			log.Println("Ошибка при запросе к БД:", err)
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
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Разрешаем CORS
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodPost {
		// Ограничиваем размер запроса
		r.ParseMultipartForm(10 << 20)

		// Получаем данные формы
		email := r.FormValue("email")
		message := r.FormValue("message")

		var attachment multipart.File
		var err error

		if r.MultipartForm != nil && r.MultipartForm.File["attachment"] != nil {
			attachment, _, err = r.FormFile("attachment")
			if err != nil {
				log.Println("Ошибка чтения вложения:", err)
				http.Error(w, `{"status":"fail","message":"Failed to read attachment"}`, http.StatusBadRequest)
				return
			}
			defer attachment.Close()

			// Сохраняем файл на диск
			file, err := ioutil.TempFile("./uploads", "attachment-*.jpg")
			if err != nil {
				log.Println("Ошибка сохранения файла:", err)
				http.Error(w, `{"status":"fail","message":"Failed to save file"}`, http.StatusInternalServerError)
				return
			}
			defer file.Close()

			_, err = io.Copy(file, attachment)
			if err != nil {
				log.Println("Ошибка копирования файла:", err)
				http.Error(w, `{"status":"fail","message":"Failed to copy attachment"}`, http.StatusInternalServerError)
				return
			}
			log.Println("Файл сохранён:", file.Name())
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

		auth := smtp.PlainAuth("", from, password, smtpHost)
		msg := []byte("To: erme.shoinov@bk.ru\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" + body + "\r\n")

		// Отправка письма
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
		if err != nil {
			log.Println("Ошибка отправки письма:", err)
			http.Error(w, `{"status":"fail","message":"Failed to send email"}`, http.StatusInternalServerError)
			return
		}

		// Успешный ответ
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "success", "message": "Message sent successfully"}`)
		return
	}

	http.Error(w, `{"status":"fail","message":"Invalid request method"}`, http.StatusMethodNotAllowed)
}

func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodPost {
		var requestData struct {
			Message string `json:"message"`
		}

		// Декодируем данные из запроса
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil || requestData.Message == "" {
			log.Println("Некорректные данные формы:", err)
			http.Error(w, `{"status":"fail","message":"Некорректные данные формы"}`, http.StatusBadRequest)
			return
		}

		// Сохраняем сообщение в базу данных
		_, err = db.Exec("INSERT INTO messages (content) VALUES ($1)", requestData.Message)
		if err != nil {
			log.Println("Ошибка сохранения данных в базу данных:", err)
			http.Error(w, `{"status":"error","message":"Ошибка сохранения данных"}`, http.StatusInternalServerError)
			return
		}

		// Успешный ответ
		response := map[string]string{
			"status":  "success",
			"message": "Сообщение успешно отправлено",
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Println("Ошибка кодирования ответа:", err)
			http.Error(w, `{"status":"error","message":"Ошибка формирования ответа"}`, http.StatusInternalServerError)
		}
		return
	}

	// Если метод не поддерживается
	log.Println("Метод не поддерживается:", r.Method)
	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

// Обработчик для SELECT (все сообщения)
func handleSelectMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodGet {
		rows, err := db.Query("SELECT id, content FROM messages")
		if err != nil {
			log.Println("Ошибка запроса к базе данных:", err)
			http.Error(w, `{"status":"error","message":"Ошибка получения данных"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var messages []struct {
			ID      int    `json:"id"`
			Content string `json:"content"`
		}

		for rows.Next() {
			var msg struct {
				ID      int    `json:"id"`
				Content string `json:"content"`
			}
			if err := rows.Scan(&msg.ID, &msg.Content); err != nil {
				log.Println("Ошибка обработки строки:", err)
				http.Error(w, `{"status":"error","message":"Ошибка обработки данных"}`, http.StatusInternalServerError)
				return
			}
			messages = append(messages, msg)
		}

		if err := rows.Err(); err != nil {
			log.Println("Ошибка итерации строк:", err)
			http.Error(w, `{"status":"error","message":"Ошибка обработки данных"}`, http.StatusInternalServerError)
			return
		}

		// Успешный ответ
		if err := json.NewEncoder(w).Encode(messages); err != nil {
			log.Println("Ошибка кодирования JSON:", err)
			http.Error(w, `{"status":"error","message":"Ошибка формирования ответа"}`, http.StatusInternalServerError)
		}
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

// Обработчик для очистки всех сообщений
func handleClearMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodPost {
		_, err := db.Exec("DELETE FROM support_messages")
		if err != nil {
			log.Println("Ошибка при очистке сообщений:", err)
			http.Error(w, `{"status":"error","message":"Ошибка очистки данных"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"status":  "success",
			"message": "Сообщения успешно очищены",
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Println("Ошибка формирования JSON-ответа:", err)
			http.Error(w, `{"status":"error","message":"Ошибка формирования ответа"}`, http.StatusInternalServerError)
		}
		return
	}

	log.Println("Метод не поддерживается:", r.Method)
	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}
