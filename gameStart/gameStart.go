package gamestart

/*Jak ma dzialac wyszukiwanie przeciwnika oraz jak ma dzialac czekanie na przeciwnika*/

import (
	"bytes"
	"context"
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

type FireRequest struct {
	Coord string `json:"coord"`
}

type FireResponse struct {
	Result string `json:"result"`
}

type GameDescResponse struct {
	Desc     string `json:"desc"`
	Nick     string `json:"nick"`
	OppDesc  string `json:"opp_desc"`
	Opponent string `json:"opponent"`
}

// Global variables to store responses
var gameStatusResponse GameStatusResponse
var boardResponse BoardResponse
var gameDescResponse GameDescResponse

// Global variable for the auth token
var authToken string

// Global variable to track if game description has been fetched
var isGameDescFetched bool

// Global variable to track if game has been abandoned
var isGameAbandoned bool

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

	gameStatusResponse.Timer = 60

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Wysyłanie jednego żądania GET w celu uzyskania statusu gry
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

		// Wysyłanie jednego żądania GET w celu uzyskania danych planszy
		boardResponse, err = SendGetBoardRequest(client, "https://go-pjatk-server.fly.dev/api/game/board", authToken)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Board:", boardResponse.Board)

		if gameStatusResponse.GameStatus == "game_in_progress" && !isGameDescFetched {
			// Pobieranie nicku i opisu
			gameDescResponse, err = GetNickAndDesc(client, "https://go-pjatk-server.fly.dev/api/game/desc", authToken)
			if err != nil {
				log.Fatal(err)
			}
			isGameDescFetched = true
			fmt.Println("Nick:", gameDescResponse.Nick)
			fmt.Println("Desc:", gameDescResponse.Desc)
			fmt.Println("Opponent:", gameDescResponse.Opponent)
			fmt.Println("OppDesc:", gameDescResponse.OppDesc)
		}

		// Porzucenie gry, gdy timer osiągnie zero
		if gameStatusResponse.Timer == 1 && !isGameAbandoned {
			err = SendAbandonRequest(client, "https://go-pjatk-server.fly.dev/api/game/abandon", authToken)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Game abandoned due to timer reaching zero")
			isGameAbandoned = true
			gameStatusResponse.GameStatus = "abandoned"
			break
		}

		if isGameAbandoned {
			break
		}

		if gameStatusResponse.GameStatus == "waiting" {
			sendWaiting(authToken)
		}

	}

}

// Funkcja do wysyłania danych do frontendu
func SendDataToFront(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if isGameAbandoned {
		gameStatusResponse.GameStatus = "abandoned"
	}
	json.NewEncoder(w).Encode(gameStatusResponse)
}

func SendBoardToFront(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boardResponse)
}

func SendDescToFront(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameDescResponse)
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

// Handle DELETE request to abandon game
func HandleAbandonGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	authToken := r.Header.Get("x-auth-token")
	if authToken == "" {
		http.Error(w, "Missing auth token", http.StatusUnauthorized)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, "https://go-pjatk-server.fly.dev/api/game/abandon", nil)
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("x-auth-token", authToken)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error sending request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to abandon game", resp.StatusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Game abandoned successfully"))
}

// Send DELETE request to abandon game
func SendAbandonRequest(client *http.Client, url string, authToken string) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-auth-token", authToken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to abandon game, status code: %d", resp.StatusCode)
	}

	return nil
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

// Funkcja do pobierania nicku i opisu z serwera
func GetNickAndDesc(client *http.Client, url string, authToken string) (GameDescResponse, error) {
	var gameDescResponse GameDescResponse
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return gameDescResponse, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-auth-token", authToken)

	resp, err := client.Do(req)
	if err != nil {
		return gameDescResponse, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&gameDescResponse)
	if err != nil {
		return gameDescResponse, err
	}

	return gameDescResponse, nil
}

func sendWaiting(authToken string) {
	url := "https://go-pjatk-server.fly.dev/api/game/refresh"

	// Tworzymy nowy obiekt żądania
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Ustawiamy nagłówek autoryzacyjny
	req.Header.Set("x-auth-token", authToken)

	// Wysyłamy żądanie
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
}

var srv *http.Server

// RestartServer function to restart the server
func RestartServer() {
	if srv != nil {
		fmt.Println("Shutting down server...")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalf("Server Shutdown Failed:%+v", err)
		}
	}
	fmt.Println("Starting new server instance...")
	StartServer()
}

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/gameData", SendDataToFront)
	mux.HandleFunc("/api/boardData", SendBoardToFront)
	mux.HandleFunc("/api/descData", SendDescToFront)
	mux.HandleFunc("/api/data", HandlePostData)
	mux.HandleFunc("/api/fire", HandleFireRequest)
	mux.HandleFunc("/api/abandon", HandleAbandonGame)
	handler := cors.Default().Handler(mux)

	srv = &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	fmt.Println("Starting server at port 8080")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}
}
