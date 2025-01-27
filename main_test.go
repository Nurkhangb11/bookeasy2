package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func connectToDatabase() (*sql.DB, error) {
	connStr := "host=localhost port=5432 user=postgres password=123olx123 dbname=hotel_booking sslmode=disable"
	return sql.Open("postgres", connStr)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	db, err := connectToDatabase()
	if err != nil {
		http.Error(w, `{"status":"error","message":"Ошибка подключения к базе данных"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var user struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, `{"status":"error","message":"Ошибка чтения данных"}`, http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		http.Error(w, `{"status":"error","message":"Ошибка хеширования пароля"}`, http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`INSERT INTO users (first_name, last_name, email, password, is_confirmed) VALUES ($1, $2, $3, $4, $5)`,
		user.FirstName, user.LastName, user.Email, hashedPassword, false)
	if err != nil {
		http.Error(w, `{"status":"error","message":"Ошибка создания пользователя"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Пользователь успешно зарегистрирован"}`))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	db, err := connectToDatabase()
	if err != nil {
		http.Error(w, `{"status":"error","message":"Ошибка подключения к базе данных"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, `{"status":"error","message":"Ошибка чтения данных"}`, http.StatusBadRequest)
		return
	}

	var hashedPassword string
	err = db.QueryRow(`SELECT password FROM users WHERE email = $1`, creds.Email).Scan(&hashedPassword)
	if err != nil {
		http.Error(w, `{"status":"fail","message":"Неверный email или пароль"}`, http.StatusUnauthorized)
		return
	}

	if !checkPasswordHash(creds.Password, hashedPassword) {
		http.Error(w, `{"status":"fail","message":"Неверный email или пароль"}`, http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Успешный вход в систему"}`))
}



func TestConnectToDatabase(t *testing.T) {
	db, err := connectToDatabase()
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	// Проверяем подключение
	if err = db.Ping(); err != nil {
		t.Fatalf("База данных недоступна: %v", err)
	}

	t.Log("Подключение к базе данных прошло успешно")
}

func TestHandleRegister(t *testing.T) {
	db, err := connectToDatabase()
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()


	payload := `{"first_name":"John","last_name":"Doe","email":"test@example.com","password":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handleRegister(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200 OK, но получен %v", res.Status)
	}

	var response map[string]string
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}

	if response["status"] != "success" {
		t.Fatalf("Ожидался статус success, но получен %v", response["status"])
	}

	t.Log("Тест успешной регистрации прошёл успешно")
}

func TestHandleLogin(t *testing.T) {
	db, err := connectToDatabase()
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	hashedPassword, _ := hashPassword("123456")
	_, err = db.Exec(`INSERT INTO users (first_name, last_name, email, password, is_confirmed) VALUES ($1, $2, $3, $4, $5)`,
		"John", "Doe", "test@example.com", hashedPassword, true)
	if err != nil {
		t.Fatalf("Не удалось создать тестового пользователя: %v", err)
	}

	payload := `{"email":"test@example.com","password":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handleLogin(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200 OK, но получен %v", res.Status)
	}

	var response map[string]string
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}

	if response["status"] != "success" {
		t.Fatalf("Ожидался статус success, но получен %v", response["status"])
	}

	t.Log("Тест успешного логина прошёл успешно")
}
func TestIntegration_RegisterAndLogin(t *testing.T) {
	db, err := connectToDatabase()
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	_, _ = db.Exec(`DELETE FROM users WHERE email = $1`, "integration@example.com")

	registerPayload := `{"first_name":"Alice","last_name":"Smith","email":"integration@example.com","password":"securepassword"}`
	registerReq := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(registerPayload))
	registerReq.Header.Set("Content-Type", "application/json")

	registerResp := httptest.NewRecorder()
	handleRegister(registerResp, registerReq)

	if registerResp.Result().StatusCode != http.StatusOK {
		t.Fatalf("Регистрация провалилась с кодом: %v", registerResp.Result().StatusCode)
	}

	loginPayload := `{"email":"integration@example.com","password":"securepassword"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp := httptest.NewRecorder()
	handleLogin(loginResp, loginReq)

	if loginResp.Result().StatusCode != http.StatusOK {
		t.Fatalf("Логин провалился с кодом: %v", loginResp.Result().StatusCode)
	}

	t.Log("Интеграционный тест регистрации и логина прошёл успешно")
}
func TestE2E_RegisterAndLogin(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", handleRegister)
	mux.HandleFunc("/login", handleLogin)

	server := httptest.NewServer(mux)
	defer server.Close()

	registerPayload := `{"first_name":"E2E","last_name":"User","email":"e2e@example.com","password":"e2epassword"}`
	registerResp, err := http.Post(server.URL+"/register", "application/json", strings.NewReader(registerPayload))
	if err != nil {
		t.Fatalf("Ошибка выполнения запроса на регистрацию: %v", err)
	}
	defer registerResp.Body.Close()

	if registerResp.StatusCode != http.StatusOK {
		t.Fatalf("Ошибка регистрации, код: %v", registerResp.StatusCode)
	}

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
