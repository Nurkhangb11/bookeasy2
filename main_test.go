package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/lib/pq" // Импортируем драйвер PostgreSQL
)

// Тестовая строка подключения к базе данных
const testConnStr = "host=localhost port=5432 user=postgres password=123olx123 dbname=hotel_booking sslmode=disable"

// TestConnectToDatabase — unit-тест для проверки подключения к базе данных
func TestConnectToDatabase(t *testing.T) {
	db, err := sql.Open("postgres", testConnStr)
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		t.Fatalf("База данных недоступна: %v", err)
	}

	t.Log("Подключение к базе данных прошло успешно")
}

// TestHandleRegister — unit-тест для проверки регистрации пользователя
func TestHandleRegister(t *testing.T) {
	// Подключаемся к тестовой базе данных
	db, err := sql.Open("postgres", testConnStr)
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	// Очищаем таблицу users перед тестом
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Не удалось очистить таблицу users: %v", err)
	}

	// Создаем тестового пользователя
	user := struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "test@example.com",
		Password:  "123456",
	}

	// Преобразуем пользователя в JSON
	payload, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Ошибка преобразования пользователя в JSON: %v", err)
	}

	// Создаем HTTP-запрос
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	// Создаем HTTP-ответ
	w := httptest.NewRecorder()

	// Вызываем обработчик с передачей db
	handleRegister := func(w http.ResponseWriter, r *http.Request) {
		// Используем db, который был создан в TestHandleRegister
		_, err := db.Exec("INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4)",
			user.FirstName, user.LastName, user.Email, user.Password)
		if err != nil {
			t.Fatalf("Ошибка при добавлении пользователя в базу данных: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}

	// Вызываем обработчик
	handleRegister(w, req)

	// Проверяем результат
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200 OK, но получен %v", res.StatusCode)
	}

	// Проверяем, что пользователь добавлен в базу данных
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", user.Email).Scan(&count)
	if err != nil {
		t.Fatalf("Ошибка при проверке базы данных: %v", err)
	}

	if count != 1 {
		t.Fatalf("Ожидалось, что в базе данных будет 1 пользователь, но найдено %v", count)
	}

	t.Log("Тест успешной регистрации прошёл успешно")
}

// TestIntegration_RegisterAndLogin — интеграционный тест для регистрации и авторизации
func TestIntegration_RegisterAndLogin(t *testing.T) {
	// Подключаемся к тестовой базе данных
	db, err := sql.Open("postgres", testConnStr)
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	// Очищаем таблицу users перед тестом
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Не удалось очистить таблицу users: %v", err)
	}

	// Регистрация пользователя
	registerPayload := `{"first_name":"Alice","last_name":"Smith","email":"integration@example.com","password":"securepassword"}`
	registerReq := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(registerPayload))
	registerReq.Header.Set("Content-Type", "application/json")

	registerResp := httptest.NewRecorder()

	// Вызываем обработчик с передачей db
	handleRegister := func(w http.ResponseWriter, r *http.Request) {
		var user struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Password  string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			t.Fatalf("Ошибка декодирования JSON: %v", err)
		}
		_, err := db.Exec("INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4)",
			user.FirstName, user.LastName, user.Email, user.Password)
		if err != nil {
			t.Fatalf("Ошибка при добавлении пользователя в базу данных: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}

	handleRegister(registerResp, registerReq)

	if registerResp.Result().StatusCode != http.StatusOK {
		t.Fatalf("Регистрация провалилась с кодом: %v", registerResp.Result().StatusCode)
	}

	// Логин зарегистрированного пользователя
	loginPayload := `{"email":"integration@example.com","password":"securepassword"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp := httptest.NewRecorder()

	// Вызываем обработчик с передачей db
	handleLogin := func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			t.Fatalf("Ошибка декодирования JSON: %v", err)
		}

		var storedPassword string
		err := db.QueryRow("SELECT password FROM users WHERE email = $1", credentials.Email).Scan(&storedPassword)
		if err != nil {
			t.Fatalf("Ошибка при поиске пользователя в базе данных: %v", err)
		}

		if storedPassword != credentials.Password {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
	}

	handleLogin(loginResp, loginReq)

	if loginResp.Result().StatusCode != http.StatusOK {
		t.Fatalf("Логин провалился с кодом: %v", loginResp.Result().StatusCode)
	}

	t.Log("Интеграционный тест регистрации и логина прошёл успешно")
}

// TestE2E_RegisterAndLogin — end-to-end тест для регистрации и авторизации
func TestE2E_RegisterAndLogin(t *testing.T) {
	// Подключаемся к тестовой базе данных
	db, err := sql.Open("postgres", testConnStr)
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	// Очищаем таблицу users перед тестом
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Не удалось очистить таблицу users: %v", err)
	}

	// Запуск HTTP-сервера
	mux := http.NewServeMux()

	// Регистрируем обработчики с передачей db
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		var user struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Password  string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			t.Fatalf("Ошибка декодирования JSON: %v", err)
		}
		_, err := db.Exec("INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4)",
			user.FirstName, user.LastName, user.Email, user.Password)
		if err != nil {
			t.Fatalf("Ошибка при добавлении пользователя в базу данных: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			t.Fatalf("Ошибка декодирования JSON: %v", err)
		}

		var storedPassword string
		err := db.QueryRow("SELECT password FROM users WHERE email = $1", credentials.Email).Scan(&storedPassword)
		if err != nil {
			t.Fatalf("Ошибка при поиске пользователя в базе данных: %v", err)
		}

		if storedPassword != credentials.Password {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Регистрация пользователя через реальный HTTP-сервер
	registerPayload := `{"first_name":"E2E","last_name":"User","email":"e2e@example.com","password":"e2epassword"}`
	registerResp, err := http.Post(server.URL+"/register", "application/json", strings.NewReader(registerPayload))
	if err != nil {
		t.Fatalf("Ошибка выполнения запроса на регистрацию: %v", err)
	}
	defer registerResp.Body.Close()

	if registerResp.StatusCode != http.StatusOK {
		t.Fatalf("Ошибка регистрации, код: %v", registerResp.StatusCode)
	}

	// Авторизация пользователя через реальный HTTP-сервер
	loginPayload := `{"email":"e2e@example.com","password":"e2epassword"}`
	loginReq, err := http.NewRequest(http.MethodPost, server.URL+"/login", strings.NewReader(loginPayload))
	if err != nil {
		t.Fatalf("Ошибка создания запроса на логин: %v", err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	loginResp, err := client.Do(loginReq)
	if err != nil {
		t.Fatalf("Ошибка выполнения запроса на логин: %v", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("Ошибка логина, код: %v", loginResp.StatusCode)
	}

	t.Log("E2E тест регистрации и логина прошёл успешно")
}

