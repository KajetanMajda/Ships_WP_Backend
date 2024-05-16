package main

import (
	gameStart "backend/package/gameStart"
)

func main() {
	gameStart.StartServer()

	//client := httpclient.New()
	//url := "http://go-pjatk-server.fly.dev/api"
	// wpbot := true
	// resp, err := httpclient.SendRequest(client, url, wpbot)
	// if err != nil {
	// 	log.Fatalf("Błąd podczas wysyłania żądania: %v", err)
	// }
	// defer resp.Body.Close()
	// Odczytaj token z nagłówków odpowiedzi
	//token := resp.Header.Get("x-auth-token")
	//fmt.Println("Otrzymany token:", token)
	// Użyj tokenu do wykonania żądania GET
	// body, err := httpclient.SendGetRequestWithToken(client, url+"/game", token)
	// if err != nil {
	// 	log.Fatalf("Błąd podczas wysyłania żądania: %v", err)
	// }
	//fmt.Println("Odpowiedź:", string(body))
	// var data map[string]interface{}
	// err = json.Unmarshal([]byte(body), &data)
	// if err != nil {
	// 	log.Fatalf("Błąd podczas deserializacji: %v", err)
	// }
	//fmt.Println("Nick:", data["nick"])
	//fmt.Println("Game status:", data["game_status"])
	// board, err := httpclient.GetGameBoard(client, url, token)
	// if err != nil {
	// 	log.Fatalf("Błąd podczas pobierania planszy gry: %v", err)
	// }
	//fmt.Println("Plansza gry:", board)
	//g := game.New(board)
	// Wyświetl planszę gry
	//g.Board.Display()
	// Sprawdź status gry co 2 sekundy i zatrzymaj bedzie w trakcie
	// for {
	// 	gameStatus, err := httpclient.PrintGameStatus(client, url, token)
	// 	if err != nil {
	// 		log.Fatalf("Błąd podczas pobierania statusu gry: %v", err)
	// 	}
	// 	time.Sleep(2 * time.Second)
	// 	if gameStatus == "game_in_progress" {
	// 		break
	// 	}
	// }
	//Wyswietl nazwe przeciwnika
	//httpclient.ShowOponentName(client, url, token)
	// Wyswietl strzaly przeciwnika na początku gry
	// shots, err := httpclient.GetOpponentShots(client, url, token)
	// if err != nil {
	// 	log.Fatalf("Błąd podczas pobierania strzałów przeciwnika: %v", err)
	// }
	// if len(shots) > 0 {
	// 	fmt.Println("Przeciwnik strzelał. Strzały: ", shots)
	// 	for _, shot := range shots {
	// 		if game.Contains(board, shot) {
	// 			fmt.Println("Przeciwnik trafił w: ", shot)
	// 			//g.Board.Set(lightgui.Left, shot, lightgui.Hit) // Zaznacz trafienie przeciwnika na twojej planszy
	// 		} else {
	// 			fmt.Println("Przeciwnik nie trafił.")
	// 			//g.Board.Set(lightgui.Left, shot, lightgui.Miss) // Zaznacz pudło przeciwnika na twojej planszy
	// 		}
	// 	}
	// } else {
	// 	fmt.Println("Przeciwnik jeszcze nie strzelał.")
	// }
	// Wyświetl planszę gry po strzale przeciwnika
	//g.Board.Display()
	// Sprawdź status gry co 2 sekundy i zatrzymaj bedzie w trakcie
	// for {
	// 	shouldFire, err := httpclient.PrintShouldFire(client, url, token)
	// 	if err != nil {
	// 		log.Fatalf("Błąd podczas pobierania statusu gry: %v", err)
	// 	}
	// 	time.Sleep(2 * time.Second)
	// 	if shouldFire {
	// 		break
	// 	}
	// }
	// Wykonaj strzał
	//g.PlayGame(client, url, token, board)

	//fmt.Println(lobby.GetLobbyData(client, url)) //niby lista jest pusta bo nie ma zadnego oczekujacego gracza

}
