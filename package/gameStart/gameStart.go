package gamestart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/rs/cors"
)

type Data struct {
	Coords     []string `json:"coords"`
	Nick       string   `json:"nick"`
	Desc       string   `json:"desc"`
	TargetNick string   `json:"target_nick"`
	Wpbot      bool     `json:"wpbot"`
}

type GameStatusResponse struct {
	GameStatus string   `json:"game_status"`
	Nick       string   `json:"nick"`
	OppShots   []string `json:"opp_shots"`
	Opponent   string   `json:"opponent"`
	ShouldFire bool     `json:"should_fire"`
	Timer      int      `json:"timer"`
}

type BoardResponse struct {
	Board []string `json:"board"`
}

// Struct for fire request
type FireRequest struct {
	Coord string `json:"coord"`
}

// Struct for fire response
type FireResponse struct {
	Result string `json:"result"`
}

// Global variables to store responses
var gameStatusResponse GameStatusResponse
var boardResponse BoardResponse

// Global variable for the auth token
var authToken string

// Handle POST request with data from frontend to start game (board, nick, desc, target_nick, wpbot)
func HandlePostData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Coords:", data.Coords)
	fmt.Println("Nick:", data.Nick)
	fmt.Println("Desc:", data.Desc)
	fmt.Println("Wpbot:", data.Wpbot)
	fmt.Println("TargetNick:", data.TargetNick)

	client := &http.Client{}
	_, authToken, err = SendRequest(client, "https://go-pjatk-server.fly.dev/api/game", data)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Send one GET request to get game status
		gameStatusResponse, err = SendGetRequest(client, "https://go-pjatk-server.fly.dev/api/game", authToken)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("GameStatus:", gameStatusResponse.GameStatus)
		fmt.Println("Nick:", gameStatusResponse.Nick)
		fmt.Println("Opponent:", gameStatusResponse.Opponent)
		fmt.Println("OppShots:", gameStatusResponse.OppShots)
		fmt.Println("ShouldFire:", gameStatusResponse.ShouldFire)
		fmt.Println("Timer:", gameStatusResponse.Timer)

		// Send one GET request to get the board data
		boardResponse, err = SendGetBoardRequest(client, "https://go-pjatk-server.fly.dev/api/game/board", authToken)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Board:", boardResponse.Board)
	}
}

func SendDataToFront(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameStatusResponse)
}

func SendBoardToFront(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boardResponse)
}

// Handle POST from Frontend with coord only
func HandleFireRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var fireRequest FireRequest
	err = json.Unmarshal(body, &fireRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Print the received coordinate to the terminal
	fmt.Println("Received coord:", fireRequest.Coord)

	client := &http.Client{}
	fireResponse, err := SendFireRequest(client, "https://go-pjatk-server.fly.dev/api/game/fire", fireRequest, authToken)
	if err != nil {
		http.Error(w, "Error sending fire request", http.StatusInternalServerError)
		return
	}

	// Print the response from the external server to the terminal
	fmt.Println("Fire response:", fireResponse)

	// Send the response back to the frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fireResponse)
}

// Send POST request with fire coord
func SendFireRequest(client *http.Client, url string, fireRequest FireRequest, authToken string) (FireResponse, error) {
	var fireResponse FireResponse

	jsonData, err := json.Marshal(fireRequest)
	if err != nil {
		return fireResponse, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fireResponse, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-auth-token", authToken)

	resp, err := client.Do(req)
	if err != nil {
		return fireResponse, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&fireResponse)
	if err != nil {
		return fireResponse, err
	}

	return fireResponse, nil
}

func SendRequest(client *http.Client, url string, data Data) (string, string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	authToken := resp.Header.Get("x-auth-token")
	fmt.Println("x-auth-token:", authToken)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	return string(body), authToken, nil
}

func SendGetRequest(client *http.Client, url string, authToken string) (GameStatusResponse, error) {
	var gameStatusResponse GameStatusResponse

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return gameStatusResponse, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-auth-token", authToken)

	resp, err := client.Do(req)
	if err != nil {
		return gameStatusResponse, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&gameStatusResponse)
	if err != nil {
		return gameStatusResponse, err
	}

	return gameStatusResponse, nil
}

func SendGetBoardRequest(client *http.Client, url string, authToken string) (BoardResponse, error) {
	var boardResponse BoardResponse

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return boardResponse, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-auth-token", authToken)

	resp, err := client.Do(req)
	if err != nil {
		return boardResponse, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&boardResponse)
	if err != nil {
		return boardResponse, err
	}

	return boardResponse, nil
}

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/gameData", SendDataToFront)
	mux.HandleFunc("/api/boardData", SendBoardToFront)
	mux.HandleFunc("/api/data", HandlePostData)
	mux.HandleFunc("/api/fire", HandleFireRequest) // New endpoint for fire requests
	handler := cors.Default().Handler(mux)

	fmt.Println("Starting server at port 8080")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
