package backend

import (
	"encoding/json"
	"fmt"
	"github.com/festy23/GoCalcServer/pkg/rpn"
	"io"
	"net/http"
	"reflect"
	"sync"
)

var POSTRequestwWasCorrect = false
var lastResult string
var mu sync.Mutex

// Request Структура для запроса
type Request struct {
	Expression interface{} `json:"expression"`
}

// Response Структура для ответа в случае корректного выполнения
type Response struct {
	Result string `json:"result"`
}

// Структура для ответа в случае ошибки
type ErrorResponse struct {
	Error string `json:"error"`
}

// calculateHandler Обработчик запросов, который в случае запрос не типа POST, возвращает ошибку
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Обрабатываем POST-запрос
		postCalculate(w, r)
	case http.MethodGet:
		// Обрабатываем GET-запрос
		getCalculate(w, r)
	default:
		http.Error(w, "Method is not supported", http.StatusMethodNotAllowed)
	}

}

func getCalculate(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	if len(lastResult) == 0 {
		fmt.Fprintf(w, "Weclome to the GoCalcServer. Make a post-requset to start work. You didn't make a POST-request")
	} else {
		fmt.Fprintf(w, "Last correct result: %s", lastResult)
	}
}

// Обработчик POST-запроса
func postCalculate(w http.ResponseWriter, r *http.Request) {
	var req Request           // входящий json
	var errResp ErrorResponse // json для ошибки
	var resp Response         // json при корректной работе

	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)

	if err != nil {
		//Код ошибки 500
		w.WriteHeader(http.StatusInternalServerError)
		errResp.Error = "Internal server error"
		err := json.NewEncoder(w).Encode(errResp)
		if err != nil {
			return
		}
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		//Код ошибки 400
		w.WriteHeader(http.StatusBadRequest)
		errResp.Error = "Internal server error"
		err := json.NewEncoder(w).Encode(errResp)
		if err != nil {
			return
		}
		return
	}

	//Проверяем тип expression
	if reflect.TypeOf(req.Expression).Kind() != reflect.String || req.Expression == nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		errResp.Error = "Expression is not valid"
		err := json.NewEncoder(w).Encode(errResp)
		if err != nil {
			return
		}
		return
	}
	result, err := rpn.Calc(req.Expression.(string))
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		errResp.Error = "Expression is not valid"
		err := json.NewEncoder(w).Encode(errResp)
		if err != nil {
			return
		}
		return
	}
	stringResult := fmt.Sprintf("%.2f", result)

	mu.Lock()
	resp.Result = stringResult
	lastResult = stringResult
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

// StartServer Запуск сервера
func StartServer() {
	http.HandleFunc("/api/v1/calculate", calculateHandler)
	fmt.Println("Server is running on localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error while running a server:", err)
	}
}

/*
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
*/
