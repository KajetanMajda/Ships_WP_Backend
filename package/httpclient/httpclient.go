package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type PostData struct {
	Wpbot bool `json:"wpbot"`
}

type BoardResponse struct {
	Board []string `json:"board"`
}

type GameStatus struct {
	ShouldFire bool `json:"shouldFire"`
	Timer      int  `json:"timer"`
}

type FireData struct {
	Coord string `json:"coord"`
}

type FireResponse struct {
	Result string `json:"result"`
}

func New() *http.Client {
	return &http.Client{Timeout: time.Second * 10}
}

func SendRequest(client *http.Client, url string, wpbot bool) (*http.Response, error) {
	data := PostData{Wpbot: wpbot}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url+"/game", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return client.Do(req)
}

func SendGetRequestWithToken(client *http.Client, url string, token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Auth-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

func GetGameBoard(client *http.Client, url string, token string) ([]string, error) {
	// Użyj funkcji SendGetRequestWithToken do wykonania żądania GET do /api/game/board
	board, err := SendGetRequestWithToken(client, url+"/game/board", token)
	if err != nil {
		return nil, err
	}

	boardResponse := &BoardResponse{}
	err = json.Unmarshal([]byte(board), boardResponse)
	if err != nil {
		return nil, err
	}

	return boardResponse.Board, nil
}

func PrintGameStatus(client *http.Client, url string, token string) (string, error) {

	req, err := http.NewRequest(http.MethodGet, url+"/game", nil)
	if err != nil {
		log.Fatalf("Błąd podczas tworzenia żądania: %v", err)
	}
	req.Header.Set("X-Auth-Token", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Błąd podczas pobierania statusu gry: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		log.Fatalf("Błąd podczas deserializacji statusu gry: %v", err)
	}

	if gameStatus, ok := data["game_status"].(string); ok {
		//fmt.Println("Game status:", gameStatus)
		return gameStatus, nil
	}

	return "D:", nil
}

func ShowOponentName(client *http.Client, url string, token string) {

	req, err := http.NewRequest(http.MethodGet, url+"/game", nil)
	if err != nil {
		log.Fatalf("Błąd podczas tworzenia żądania: %v", err)
	}
	req.Header.Set("X-Auth-Token", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Błąd podczas pobierania statusu gry: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		log.Fatalf("Błąd podczas deserializacji statusu gry: %v", err)
	}
	if opponent, ok := data["opponent"]; ok {
		fmt.Println("Oponent", opponent)
	} else {
		fmt.Println("Oponent name: not found")
	}

}

func PrintShouldFire(client *http.Client, url string, token string) (bool, error) {

	req, err := http.NewRequest(http.MethodGet, url+"/game", nil)
	if err != nil {
		log.Fatalf("Błąd podczas tworzenia żądania: %v", err)
	}
	req.Header.Set("X-Auth-Token", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Błąd podczas pobierania statusu gry: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		log.Fatalf("Błąd podczas deserializacji statusu gry: %v", err)
	}

	if shouldFire, ok := data["should_fire"].(bool); ok {
		//fmt.Println("ShouldFire: ", shouldFire)
		return shouldFire, nil
	}

	return false, nil
}

func PrintGameTime(client *http.Client, url string, token string) (float64, error) {

	req, err := http.NewRequest(http.MethodGet, url+"/game", nil)
	if err != nil {
		log.Fatalf("Błąd podczas tworzenia żądania: %v", err)
	}
	req.Header.Set("X-Auth-Token", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Błąd podczas pobierania czasu gry: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		log.Fatalf("Błąd podczas deserializacji czasu gry: %v", err)
	}

	if timer, ok := data["timer"].(float64); ok {

		fmt.Println("\nTimer: Time left -> ", timer, " ")
		//print("\033[1000D\033[A")
	} else {
		//fmt.Println("Timer: Run out of time")
		return 0, nil
	}

	return 60, nil
}

func AbandomGame(client *http.Client, url string, token string) error {
	req, err := http.NewRequest(http.MethodDelete, url+"/game/abandon", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func SendFire(client *http.Client, url string, token string, coord string) (string, error) {
	data := FireData{Coord: coord}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url+"/game/fire", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var fireResp FireResponse
	if err := json.NewDecoder(resp.Body).Decode(&fireResp); err != nil {
		return "", err
	}

	// Wydrukuj wynik
	fmt.Println("Wynik strzału:", fireResp.Result)

	return fireResp.Result, nil
}

func GetOpponentShots(client *http.Client, url string, token string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, url+"/game", nil)
	if err != nil {
		log.Fatalf("Błąd podczas tworzenia żądania: %v", err)
	}
	req.Header.Set("X-Auth-Token", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Błąd podczas pobierania statusu gry: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		log.Fatalf("Błąd podczas deserializacji tablicy strzałow przciwnika: %v", err)
	}

	if opp_shots, ok := data["opp_shots"].([]interface{}); ok {
		shots := make([]string, len(opp_shots))
		for i, v := range opp_shots {
			shots[i] = v.(string)
		}
		fmt.Println("Opp_shots: ", shots)
		return shots, nil
	} else {
		fmt.Println("Opp_shots not found")
	}

	return []string{}, nil
}
