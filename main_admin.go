package main

import (
	"encoding/json"
	"net/http"
)

type Car struct {
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Category string `json:"category"`
}

var cars []Car // Хранилище машин

func getCarsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cars)
}

func addCarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var car Car
	if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cars = append(cars, car)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(car)
}

func main() {
	http.HandleFunc("/api/cars", getCarsHandler)
	http.HandleFunc("/api/cars/add", addCarHandler)

	http.ListenAndServe(":8080", nil)
}
