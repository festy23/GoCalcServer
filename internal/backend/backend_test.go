package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Структура теста, который должен проходить корректно

type TestDataSuccess struct {
	testReq Request
	testRes Response
}

// Структура теста, который должен вызвать ошибку

type TestDataFail struct {
	testReq Request
	testRes ErrorResponse
}

func TestBackend(t *testing.T) {
	testsSuccess := []TestDataSuccess{
		{
			testReq: Request{Expression: "2+2*2"},
			testRes: Response{Result: "6.00"},
		},
		{
			testReq: Request{Expression: "5*3-1"},
			testRes: Response{Result: "14.00"},
		},
		{
			testReq: Request{Expression: "7*(8+1)"},
			testRes: Response{Result: "63.00"},
		},
		{
			testReq: Request{Expression: "100*(100+100)-100/100"},
			testRes: Response{Result: "19999.00"},
		},
		{
			testReq: Request{Expression: "100*(99+100/32-12)*(2331013-12)+3242"},
			testRes: Response{Result: "21008149754.50"},
		},
		{
			testReq: Request{Expression: "88*88*(2*2-3-2*10.12)/12*24924-241-2/11"},
			testRes: Response{Result: "-309461942.30"},
		},
		{
			testReq: Request{Expression: "1/169*1488-1488+3131-69*1488/148*88"},
			testRes: Response{Result: "-59396.41"},
		},
		{
			testReq: Request{Expression: "14888*1/88*13/88888/148881301413+3141234/12341+41342-1324*(1243-1/134134)"},
			testRes: Response{Result: "-1604135.45"},
		},
	}

	//Тесты, которые должны вызвать ошибку
	testsFail := []TestDataFail{
		{
			testReq: Request{Expression: "2+2a"},
			testRes: ErrorResponse{Error: "Expression is not valid"},
		},
		{
			testReq: Request{Expression: "2//2"},
			testRes: ErrorResponse{Error: "Expression is not valid"},
		},
		{
			testReq: Request{Expression: ""},
			testRes: ErrorResponse{Error: "Expression is not valid"},
		},
		{
			testReq: Request{Expression: 0},
			testRes: ErrorResponse{Error: "Expression is not valid"},
		},
		{
			testReq: Request{Expression: 414.1},
			testRes: ErrorResponse{Error: "Expression is not valid"},
		},
		{
			testReq: Request{Expression: -321},
			testRes: ErrorResponse{Error: "Expression is not valid"},
		},
	}

	// Создаем тестовый сервер для каждого теста
	handler := http.HandlerFunc(calculateHandler)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Запускаем сервер в отдельной горутине, чтобы можно было отправить запросы
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Даем серверу время на запуск
	time.Sleep(1 * time.Second) // Подождать 1 секунду, чтобы сервер был готов

	// Тестируем успешные запросы
	for _, testData := range testsSuccess {
		t.Run("Success", func(t *testing.T) {
			reqBody, err := json.Marshal(testData.testReq)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			// Отправляем POST-запрос на тестовый сервер
			req, err := http.NewRequest("POST", ts.URL+"/api/v1/calculate", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Получаем ответ от сервера
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			// Проверяем статус-код
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
			}

			var actual Response
			if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if actual.Result != testData.testRes.Result {
				t.Errorf("Expected result %v, got %v", testData.testRes.Result, actual.Result)
			}
		})
	}

	// Тестируем неудачные запросы
	for _, testData := range testsFail {
		t.Run(fmt.Sprintf("Failure: %v", testData.testReq.Expression), func(t *testing.T) {
			reqBody, err := json.Marshal(testData.testReq)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			// Отправляем POST-запрос на тестовый сервер
			req, err := http.NewRequest("POST", ts.URL+"/api/v1/calculate", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Получаем ответ от сервера
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnprocessableEntity {
				t.Errorf("Expected status code %d, got %d", http.StatusUnprocessableEntity, resp.StatusCode)
			}

			var actual ErrorResponse
			if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			if actual.Error != testData.testRes.Error {
				t.Errorf("Expected error %v, got %v", testData.testRes.Error, actual.Error)
			}
		})
	}
}
