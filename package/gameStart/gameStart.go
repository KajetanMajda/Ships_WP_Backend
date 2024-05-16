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

type BotData struct {
	TargetNick string `json:"target_nick"`
	Wpbot      bool   `json:"wpbot"`
}

type GameStatusResponse struct {
	GameStatus string   `json:"game_status"`
	Nick       string   `json:"nick"`
	OppShots   []string `json:"opp_shots"`
	Opponent   string   `json:"opponent"`
	ShouldFire bool     `json:"should_fire"`
	Timer      int      `json:"timer"`
}

// Handle POST request with data from frontedn to start game (board, nick, desc, target_nick, wpbot)
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

	fmt.Println("Wpbot", data.Wpbot)
	fmt.Println("TargetNick", data.TargetNick)

	client := &http.Client{}
	_, authToken, err := SendRequest(client, "https://go-pjatk-server.fly.dev/api/game", data)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		gameStatusResponse, err = SendGetRequest(client, "https://go-pjatk-server.fly.dev/api/game", authToken)
		if err != nil {
			log.Fatal(err)
		}
	}
}

var gameStatusResponse GameStatusResponse

func SendDataToFront(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameStatusResponse)

	fmt.Println("GameStatus:", gameStatusResponse.GameStatus)
	fmt.Println("Nick:", gameStatusResponse.Nick)
	fmt.Println("Opponent:", gameStatusResponse.Opponent)
	fmt.Println("OppShots:", gameStatusResponse.OppShots)
	fmt.Println("ShouldFire:", gameStatusResponse.ShouldFire)
	fmt.Println("Timer:", gameStatusResponse.Timer)
}

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/gameData", SendDataToFront)
	mux.HandleFunc("/api/data", HandlePostData)
	handler := cors.Default().Handler(mux)

	fmt.Println("Starting server at port 8080")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}

}

// Send POST request with data to start game
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

// Send GET request with token to get game_status, nick, opp_shots, opponent, should_fire, timer
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
