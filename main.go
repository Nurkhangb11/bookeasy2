package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
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

	// Получаем порт из переменной окружения (Render использует PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000" // Порт по умолчанию, если переменная окружения не задана
	}

	// Запуск сервера на 0.0.0.0 и полученном порту
	fmt.Printf("Сервер запущен на http://0.0.0.0:%s\n", port)
	err = http.ListenAndServe("0.0.0.0:"+port, nil) // Слушаем на всех интерфейсах
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

// Функция подключения к PostgreSQL
func connectToDatabase() (*sql.DB, error) {
	// Получаем строку подключения из переменной окружения
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	// Подключаемся к базе данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Проверяем, что подключение установлено
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	fmt.Println("Connected to the database successfully!")
	return db, nil
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

		var isConfirmed bool
		err = db.QueryRow(`SELECT password, is_confirmed FROM users WHERE email = $1`, user.Email).Scan(&storedPassword, &isConfirmed)
		if err != nil {
			http.Error(w, `{"status":"fail","message":"Пользователь не найден"}`, http.StatusUnauthorized)
			return
		}

		if !isConfirmed {
			http.Error(w, `{"status":"fail","message":"Подтвердите email перед входом"}`, http.StatusForbidden)
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

	// Проверяем, что метод GET
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

// Обработчик для добавления сообщения через хэлпдеск
func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var requestData struct {
			Message string `json:"message"` // Используем "message", чтобы соответствовать клиентскому запросу
		}

		// Декодируем данные из запроса
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil || requestData.Message == "" {
			log.Println("Некорректные данные формы:", err)
			http.Error(w, `{"status":"fail","message":"Некорректные данные формы"}`, http.StatusBadRequest)
			return
		}

		// Сохраняем сообщение в таблицу `messages`
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

	if r.Method == http.MethodGet {
		// Изменяем запрос, чтобы использовать таблицу `messages`
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

		// Отправляем данные в формате JSON
		json.NewEncoder(w).Encode(messages)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

// Обработчик для очистки всех сообщений
func handleClearMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		_, err := db.Exec("DELETE FROM support_messages")
		if err != nil {
			http.Error(w, `{"status":"error","message":"Ошибка очистки данных"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"status":  "success",
			"message": "Сообщения успешно очищены",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

type Car struct {
	Model    string
	Price    int
	Rating   float64
	Category string
	Brand    string
}

var cars = []Car{
	{"Toyota Corolla", 50, 4.5, "Sedan", "Toyota"},
	{"Ford Explorer", 80, 4.0, "SUV", "Ford"},
	{"Tesla Model 3", 120, 5.0, "Electric", "Tesla"},
	{"Honda Civic", 40, 4.2, "Sedan", "Honda"},
	{"BMW XM", 200, 5.0, "SUV", "BMW"},
	{"Cadillac Escalade", 150, 4.8, "SUV", "Cadillac"},
	{"Rolls Royce Cullinan", 5000, 5.0, "SUV", "Rolls Royce"},
	{"Mercedes G63", 300, 4.9, "SUV", "Mercedes"},
	{"Mercedes GLE53", 150, 4.5, "SUV", "Mercedes"},
	{"GMC SLT", 100, 4.0, "SUV", "GMC"},
	{"Porsche Macan", 300, 4.7, "SUV", "Porsche"},
	{"Nissan Patrol", 100, 4.2, "SUV", "Nissan"},
	{"BMW M4 Competition", 200, 4.8, "Sedan", "BMW"},
	{"Audi RS3", 220, 4.6, "Sedan", "Audi"},
	{"Audi RS5", 270, 4.7, "Sedan", "Audi"},
	{"Audi S8", 300, 4.9, "Sedan", "Audi"},
	{"BMW 730LI", 290, 4.6, "Sedan", "BMW"},
	{"Mercedes EQE 350", 120, 4.5, "Electric", "Mercedes"},
	{"Tesla Model 3", 120, 5.0, "Electric", "Tesla"},
	{"Porsche 718", 4718, 4.9, "Sports", "Porsche"},
	{"Porsche 911 Turbo S", 9000, 5.0, "Sports", "Porsche"},
	{"Ferrari F8 Tributo", 9999, 5.0, "Sports", "Ferrari"},
	{"Audi R8", 2000, 4.8, "Sports", "Audi"},
	{"Audi RS6", 300, 4.7, "Sports", "Audi"},
	{"Mercedes V250", 2500, 4.6, "Van", "Mercedes"},
}

const carsPerPage = 3

// Функция для главной страницы
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		log.Println("Ошибка загрузки шаблона:", err)
		return
	}
	tmpl.Execute(w, nil)
}

// Функция для фильтрации, сортировки и пагинации автомобилей
func carsHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	brand := r.URL.Query().Get("brand")
	sortBy := r.URL.Query().Get("sort")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	// Фильтрация автомобилей
	filteredCars := []Car{}
	for _, car := range cars {
		if (category == "" || car.Category == category) && (brand == "" || car.Brand == brand) {
			filteredCars = append(filteredCars, car)
		}
	}

	// Сортировка автомобилей
	switch sortBy {
	case "price":
		sortCarsByPrice(filteredCars)
	case "rating":
		sortCarsByRating(filteredCars)
	}

	// Пагинация
	startIndex := (page - 1) * carsPerPage
	endIndex := startIndex + carsPerPage
	if endIndex > len(filteredCars) {
		endIndex = len(filteredCars)
	}

	// Передаем отфильтрованные и отсортированные данные
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		log.Println("Ошибка загрузки шаблона:", err)
		return
	}
	tmpl.Execute(w, struct {
		Cars        []Car
		TotalPages  int
		CurrentPage int
	}{
		Cars:        filteredCars[startIndex:endIndex],
		TotalPages:  (len(filteredCars) + carsPerPage - 1) / carsPerPage, // Общее количество страниц
		CurrentPage: page,
	})
}

// Функция сортировки автомобилей по цене
func sortCarsByPrice(cars []Car) {
	sort.SliceStable(cars, func(i, j int) bool {
		return cars[i].Price < cars[j].Price
	})
}

// Функция сортировки автомобилей по рейтингу
func sortCarsByRating(cars []Car) {
	sort.SliceStable(cars, func(i, j int) bool {
		return cars[i].Rating > cars[j].Rating
	})
}

// Функция для обработки статических файлов (CSS, изображения и т. д.)
func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static"+r.URL.Path)
}

// Функция для запуска сервера
func startServer() {
	r := mux.NewRouter() // создаем новый маршрутизатор

	// Регистрация обработчиков маршрутов
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/cars", carsHandler).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Логирование запуска сервера
	fmt.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", r)) // запуск сервера с маршрутизатором
}
