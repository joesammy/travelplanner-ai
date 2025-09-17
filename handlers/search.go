package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"travel-search/models"
)

// A small struct to cache token in memory
var amadeusToken string
var tokenExpiry time.Time

func getAmadeusToken() (string, error) {
	// Return cached token if still valid
	if amadeusToken != "" && time.Now().Before(tokenExpiry) {
		return amadeusToken, nil
	}

	key := os.Getenv("AMADEUS_API_KEY")
	secret := os.Getenv("AMADEUS_API_SECRET")

	data := []byte(fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", key, secret))
	req, _ := http.NewRequest("POST", "https://test.api.amadeus.com/v1/security/oauth2/token", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	token := result["access_token"].(string)
	expiresIn := int(result["expires_in"].(float64))

	amadeusToken = token
	tokenExpiry = time.Now().Add(time.Duration(expiresIn-60) * time.Second) // refresh a minute early

	return token, nil
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := getAmadeusToken()
	if err != nil {
		log.Println("Error getting token:", err)
		http.Error(w, "Auth failed", http.StatusInternalServerError)
		return
	}

	// Build request URL
	url := fmt.Sprintf(
		"https://test.api.amadeus.com/v2/shopping/flight-offers?originLocationCode=%s&destinationLocationCode=%s&departureDate=%s&adults=1&currencyCode=USD&max=3",
		req.From, req.To, req.Date,
	)

	httpReq, _ := http.NewRequest("GET", url, nil)
	httpReq.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Println("API call failed:", err)
		http.Error(w, "Failed to fetch flights", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// ⬇️ NEW CODE STARTS HERE
	var amadeusResp map[string]interface{}
	if err := json.Unmarshal(body, &amadeusResp); err != nil {
		http.Error(w, "Failed to parse API response", http.StatusInternalServerError)
		return
	}

	var results []models.FlightOption
	for _, d := range amadeusResp["data"].([]interface{}) {
		offer := d.(map[string]interface{})

		// Price
		priceStr := offer["price"].(map[string]interface{})["total"].(string)
		var price float64
		fmt.Sscanf(priceStr, "%f", &price)

		// Duration (from first itinerary)
		itineraries := offer["itineraries"].([]interface{})
		firstItinerary := itineraries[0].(map[string]interface{})
		duration := firstItinerary["duration"].(string)

		// Carrier code (from validatingAirlineCodes[0])
		carrierCodes := offer["validatingAirlineCodes"].([]interface{})
		carrier := carrierCodes[0].(string)

		results = append(results, models.FlightOption{
			Carrier:  carrier,
			Price:    price,
			Duration: duration,
		})
	}

	// Send simplified JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
	// ⬆️ NEW CODE ENDS HERE
}
