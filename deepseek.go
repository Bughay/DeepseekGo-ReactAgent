package deepseekgoreactagent

import "time"


type ChatTemplate struct {

}

func deepseekOneshot(message string) any {
	err := godotenv.Load()
	apikey := os.Getenv("DEEPSEEKAPIKEY")
	if err != nil {
		log.Fatal("Error loading .env file")

	}
	jsonData, err := json.Marshal(message) 
	client := &http.Client(Timeout: 30 * time.Second)
	
}