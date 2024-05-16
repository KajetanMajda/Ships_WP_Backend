package lobby

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LobbyData struct {
	GameStatus string `json:"game_status"`
	Nick       string `json:"nick"`
}

func GetLobbyData(Client *http.Client, url string) ([]LobbyData, error) {
	resp, err := Client.Get(url + "/lobby")
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	var lobbyData []LobbyData
	err = json.Unmarshal(body, &lobbyData)
	if err != nil {
		fmt.Println("Error unmarshalling response body:", err)
		return nil, err
	}

	return lobbyData, nil
}
