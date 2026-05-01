package main
import ("encoding/json"; "fmt"; "log"; "net/http"; "time")

const MyCarnet = "202308227"

type HealthResponse struct {
	Status string `json:"status"`; Message string `json:"message"`; Timestamp time.Time `json:"timestamp"`; VM string `json:"VM"`; Carnet string `json:"carnet"`
}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(HealthResponse{Status: "UP", Message: "API2 is Ready", Timestamp: time.Now(), VM: "VM1", Carnet: MyCarnet})
	})
	fmt.Println("API 2 corriendo en puerto 8002...")
	log.Fatal(http.ListenAndServe(":8002", nil))
}
