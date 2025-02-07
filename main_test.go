package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

// Мок для отправки email
type MockEmailSender struct{}

func (m *MockEmailSender) SendEmail(to []string, subject, body string) error {
	// Ничего не делаем, просто возвращаем успех
	return nil
}

// Тест для handleRegister с использованием sqlmock и мока для отправки email
func TestHandleRegister(t *testing.T) {
	// Создаем мок базы данных
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка создания мока: %v", err)
	}
	defer mockDB.Close()

	// Подменяем глобальную переменную db на мок
	db = mockDB

	// Подменяем глобальную переменную emailSender на мок
	emailSender = &MockEmailSender{}

	// Ожидаем, что будет выполнен запрос INSERT
	mock.ExpectExec("INSERT INTO users").
		WithArgs("John", "Doe", "john.doe@example.com", "password123", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Подготовка тестовых данных
	payload := `{"first_name":"John","last_name":"Doe","email":"john.doe@example.com","password":"password123"}`
	req, err := http.NewRequest("POST", "/register", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}

	// Создание ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRegister)

	// Вызов обработчика
	handler.ServeHTTP(rr, req)

	// Проверка статуса ответа
	assert.Equal(t, http.StatusOK, rr.Code, "Ожидался статус 200")

	// Проверка тела ответа
	expected := `{"status":"success","message":"Пользователь успешно зарегистрирован. Проверьте email для подтверждения."}`
	assert.JSONEq(t, expected, rr.Body.String(), "Неверный ответ сервера")

	// Проверка, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Невыполненные ожидания: %v", err)
	}

	// Вывод сообщения о том, что запрос был успешно обработан
	t.Log("Запрос на регистрацию пользователя успешно обработан")
}

// Тест для end-to-end сценария регистрации пользователя
func TestUserRegistrationE2E(t *testing.T) {
	// Настройка Selenium WebDriver
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Path: "", // Укажите путь к ChromeDriver, если он не в PATH
	}
	caps.AddChrome(chromeCaps)

	// Запуск WebDriver
	wd, err := selenium.NewRemote(caps, "")
	if err != nil {
		t.Fatalf("Ошибка запуска WebDriver: %v", err)
	}
	defer wd.Quit()

	// Открываем страницу регистрации
	if err := wd.Get("http://localhost:8080/register"); err != nil {
		t.Fatalf("Ошибка открытия страницы регистрации: %v", err)
	}

	// Заполняем форму регистрации
	firstNameInput, err := wd.FindElement(selenium.ByID, "first_name")
	if err != nil {
		t.Fatalf("Ошибка поиска поля first_name: %v", err)
	}
	firstNameInput.SendKeys("John")

	lastNameInput, err := wd.FindElement(selenium.ByID, "last_name")
	if err != nil {
		t.Fatalf("Ошибка поиска поля last_name: %v", err)
	}
	lastNameInput.SendKeys("Doe")

	emailInput, err := wd.FindElement(selenium.ByID, "email")
	if err != nil {
		t.Fatalf("Ошибка поиска поля email: %v", err)
	}
	emailInput.SendKeys("john.doe@example.com")

	passwordInput, err := wd.FindElement(selenium.ByID, "password")
	if err != nil {
		t.Fatalf("Ошибка поиска поля password: %v", err)
	}
	passwordInput.SendKeys("password123")

	// Нажимаем кнопку регистрации
	registerButton, err := wd.FindElement(selenium.ByID, "register-button")
	if err != nil {
		t.Fatalf("Ошибка поиска кнопки регистрации: %v", err)
	}
	registerButton.Click()

	// Проверяем, что регистрация прошла успешно
	successMessage, err := wd.FindElement(selenium.ByID, "success-message")
	if err != nil {
		t.Fatalf("Ошибка поиска сообщения об успешной регистрации: %v", err)
	}

	text, err := successMessage.Text()
	if err != nil {
		t.Fatalf("Ошибка получения текста сообщения: %v", err)
	}

	expected := "Пользователь успешно зарегистрирован. Проверьте email для подтверждения."
	if text != expected {
		t.Errorf("Ожидалось сообщение '%s', получено '%s'", expected, text)
	}

	// Вывод сообщения о том, что E2E тест завершен
	t.Log("End-to-End тест регистрации пользователя успешно завершен")
}
