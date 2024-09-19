package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var a App

func TestMain(m *testing.M) {
	err := a.Initialise(DbUser, DbPassword, "test")
	if err != nil {
		log.Fatal("error occured while initialising database")
	}
	m.Run()
}

func createTable() {
	query := `CREATE TABLE IF NOT EXISTS products (
	id int NOT NULL AUTO_INCREMENT, 
	name varchar(255) NOT NULL, 
	quantity int, 
	price float(10,7), 
	PRIMARY KEY(id)
	);`
	_, err := a.Db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.Db.Exec("DELETE FROM products")
	a.Db.Exec("ALTER TABLE products AUTO_INCREMENT=1")
	log.Println("Table cleared")
}

func addProduct(name string, quantity int, price int) {
	query := fmt.Sprintf("INSERT INTO products (name, quantity, price) VALUES ('%v', %v, %v);", name, quantity, price)
	a.Db.Exec(query)
	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	log.Println("Product added")
	// }

}

func TestGetProduct(t *testing.T) {
	createTable()
	clearTable()
	addProduct("keyboard", 300, 120)

	request := httptest.NewRequest("GET", "/product/1", nil)
	response := sendRequest(request)

	checkStatusCode(t, http.StatusOK, response.Code)

	// Testing with non existent ID
	request = httptest.NewRequest("GET", "/product/10", nil)
	response = sendRequest(request)

	checkStatusCode(t, http.StatusNotFound, response.Code)

	// Testing with invalid ID format
	request = httptest.NewRequest("GET", "/product/testabc", nil)
	response = sendRequest(request)

	checkStatusCode(t, http.StatusBadRequest, response.Code)
}

func TestCreateProduct(t *testing.T) {
	clearTable()
	requestBody := []byte(`{"name": "Mouse", "quantity": 12, "price": 99}`)
	request := httptest.NewRequest("POST", "/product", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-type", "appication/json")

	response := sendRequest(request)

	checkStatusCode(t, http.StatusCreated, response.Code)

	// Checking the returned response
	var m map[string]interface{}

	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "Mouse" {
		t.Errorf("Expected name: %v, got : %v", "Mouse", m["name"])
	}

	if m["quantity"] != 12.0 {
		t.Errorf("Expected quantity: %v, got : %v", 1.0, m["quantity"])
	}
}

func TestUpdateProduct(t *testing.T) {

	clearTable()
	addProduct("Keyboard", 100, 199)
	addProduct("USB_128Gb", 400, 55)
	requestBody := []byte(`{"name": "Keyboard", "quantity": 200, "price": 199}`)
	request := httptest.NewRequest("PUT", "/product/1", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-type", "appication/json")

	response := sendRequest(request)

	checkStatusCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}

	json.Unmarshal(response.Body.Bytes(), &m)

	// Testing the returned response with updated data
	if m["name"] != "Keyboard" {
		t.Errorf("Expected name: %v, got : %v", "Keyboard", m["name"])
	}

	if m["quantity"] == 100.0 {
		t.Errorf("Expected quantity: %v, got : %v", 100, m["quantity"])
	}

	if m["price"] != 199.0 {
		t.Errorf("Expected price: %v, got : %v", 199, m["price"])
	}

	// Test with non existent ID
	request = httptest.NewRequest("PUT", "/product/10", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-type", "application/json")
	response = sendRequest(request)

	checkStatusCode(t, http.StatusNotFound, response.Code)

	// Test with Invalid ID format
	request = httptest.NewRequest("PUT", "/product/abc", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-type", "application/json")
	response = sendRequest(request)

	checkStatusCode(t, http.StatusBadRequest, response.Code)
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProduct("Headphone", 150, 250)
	addProduct("Ddr4 RAM 16Gb", 100, 25)

	request := httptest.NewRequest("DELETE", "/product/1", nil)
	response := sendRequest(request)

	checkStatusCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}

	json.Unmarshal(response.Body.Bytes(), &m)

	// Test response message body
	if m["result"] != "Deletion successful" {
		t.Errorf("Expected message : %v, got : %v", "Deletion successful", m["result"])
	}

	// Test with unknown ID

	request = httptest.NewRequest("DELETE", "/product/10", nil)
	response = sendRequest(request)

	checkStatusCode(t, http.StatusNotFound, response.Code)

	// Test with invalid ID format
	request = httptest.NewRequest("DELETE", "/product/abc", nil)
	response = sendRequest(request)

	checkStatusCode(t, http.StatusBadRequest, response.Code)

}

func sendRequest(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, request)
	return recorder
}

func checkStatusCode(t *testing.T, expectedCode int, actualCode int) {
	if expectedCode != actualCode {
		t.Errorf("Expected code: %v, got %v", expectedCode, actualCode)
		t.Log("Code not match")
	}

}
