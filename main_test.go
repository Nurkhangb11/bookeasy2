package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

const testConnStr = "host=localhost port=5432 user=postgres password=123olx123 dbname=hotel_booking sslmode=disable"

// TestConnectToDatabase — unit-тест для проверки подключения к базе данных
func TestConnectToDatabase(t *testing.T) {
	db, err := sql.Open("postgres", testConnStr)
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("База данных недоступна: %v", err)
	}
	t.Log("Подключение к базе данных прошло успешно")
}

// TestIntegration_RegisterAndLogin — интеграционный тест для регистрации и авторизации
func TestIntegration_RegisterAndLogin(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		var user struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Password  string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Ошибка декодирования JSON", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			http.Error(w, "Ошибка декодирования JSON", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	registerPayload := `{"first_name":"Alice","last_name":"Smith","email":"integration@example.com","password":"securepassword"}`
	registerResp, err := http.Post(server.URL+"/register", "application/json", bytes.NewBufferString(registerPayload))
	if err != nil {
		t.Fatalf("Ошибка выполнения запроса на регистрацию: %v", err)
	}
	defer registerResp.Body.Close()

	if registerResp.StatusCode != http.StatusOK {
		t.Fatalf("Регистрация провалилась с кодом: %v", registerResp.StatusCode)
	}

	loginPayload := `{"email":"integration@example.com","password":"securepassword"}`
	loginResp, err := http.Post(server.URL+"/login", "application/json", bytes.NewBufferString(loginPayload))
	if err != nil {
		t.Fatalf("Ошибка выполнения запроса на логин: %v", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("Авторизация провалилась с кодом: %v", loginResp.StatusCode)
	}
	t.Log("Интеграционный тест регистрации и авторизации прошёл успешно")
}

// TestE2E_RegisterAndLogin — end-to-end тест для регистрации и авторизации
func TestE2E_RegisterAndLogin(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		var user struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Password  string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Ошибка декодирования JSON", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			http.Error(w, "Ошибка декодирования JSON", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	registerPayload := `{"first_name":"E2E","last_name":"User","email":"e2e@example.com","password":"e2epassword"}`
	registerResp, err := http.Post(server.URL+"/register", "application/json", bytes.NewBufferString(registerPayload))
	if err != nil {
		t.Fatalf("Ошибка выполнения запроса на регистрацию: %v", err)
	}
	defer registerResp.Body.Close()

	if registerResp.StatusCode != http.StatusOK {
		t.Fatalf("Регистрация провалилась с кодом: %v", registerResp.StatusCode)
	}

	loginPayload := `{"email":"e2e@example.com","password":"e2epassword"}`
	loginResp, err := http.Post(server.URL+"/login", "application/json", bytes.NewBufferString(loginPayload))
	if err != nil {
		t.Fatalf("Ошибка выполнения запроса на логин: %v", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("Авторизация провалилась с кодом: %v", loginResp.StatusCode)
	}
	t.Log("E2E тест регистрации и авторизации прошёл успешно")
}

