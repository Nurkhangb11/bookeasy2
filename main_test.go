package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGenerateToken — unit тест для функции generateToken
func TestGenerateToken(t *testing.T) {
	token1 := generateToken()
	time.Sleep(1 * time.Millisecond)
	token2 := generateToken()

	if token1 == token2 {
		t.Errorf("Токены должны быть уникальными, но они одинаковы: %s и %s", token1, token2)
	}
}

// TestHandleRegister — интеграционный тест для маршрута /register
func TestHandleRegister(t *testing.T) {
	user := User{
		FirstName: "Test",
		LastName:  "User",
		Email:     "testuser@example.com",
		Password:  "password123",
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Ошибка маршалинга данных: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handleRegister)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, но получили %v", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}
	if response["status"] != "success" {
		t.Errorf("Ожидался статус 'success', но получили %v", response["status"])
	}
}

// TestCompleteRegistrationFlow — end-to-end тест для полного потока регистрации и логина
func TestCompleteRegistrationFlow(t *testing.T) {
	user := User{
		FirstName: "Test",
		LastName:  "User",
		Email:     "testuser@example.com",
		Password:  "password123",
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Ошибка маршалинга данных: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRegister)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, но получили %v", rr.Code)
	}

	loginData := map[string]string{
		"email":    "testuser@example.com",
		"password": "password123",
	}
	loginBody, _ := json.Marshal(loginData)

	reqLogin := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	rrLogin := httptest.NewRecorder()
	handlerLogin := http.HandlerFunc(handleLogin)
	handlerLogin.ServeHTTP(rrLogin, reqLogin)

	if rrLogin.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, но получили %v", rrLogin.Code)
	}
}
