package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)


const MyCarnet = "202308227"


type HealthResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	VM        string    `json:"VM"`
	Carnet    string    `json:"carnet"`
}

type CallResponse struct {
	ApiName    string `json:"apiname"`
	Message    string `json:"message"`
	Connection bool   `json:"connection"`
	Carnet     string `json:"carnet"`
}

func main() {
	//endpoint 1: Salud
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "UP",
			Message:   "API1 is Ready",
			Timestamp: time.Now(),
			VM:        "VM1",
			Carnet:    MyCarnet,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	//endpoint 2: Llamar a API 2 (puerto 8002)
	http.HandleFunc(fmt.Sprintf("/api1/%s/call-api2", MyCarnet), func(w http.ResponseWriter, r *http.Request) {
		callOtherAPI(w, "http://localhost:8002/health", "API2", "VM1")
	})

	//endpoint 3: Llamar a API 3 (IP de la VM2, puerto 8003)
	http.HandleFunc(fmt.Sprintf("/api1/%s/call-api3", MyCarnet), func(w http.ResponseWriter, r *http.Request) {
		callOtherAPI(w, "http://192.168.122.14:8003/health", "API3", "VM2")
	})

	fmt.Println("API 1 corriendo en puerto 8001...")
	log.Fatal(http.ListenAndServe(":8001", nil))
}

func callOtherAPI(w http.ResponseWriter, url string, targetAPI string, targetVM string) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	connectionSuccess := false
	message := fmt.Sprintf("ERROR: The %s located on the %s is not working", targetAPI, targetVM)

	if err == nil && resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		var healthResp HealthResponse
		json.Unmarshal(body, &healthResp)
		if healthResp.Status == "UP" {
			connectionSuccess = true
			message = fmt.Sprintf("The %s located on the %s is working", targetAPI, targetVM)
		}
		resp.Body.Close()
	}
	finalResponse := CallResponse{ApiName: targetAPI, Message: message, Connection: connectionSuccess, Carnet: MyCarnet}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finalResponse)
}
