package game

import (
	httpclient "backend/package/httpclient"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
	gui "github.com/grupawp/warships-lightgui/v2"
)

type Game struct {
	Board *gui.Board
}

func New(board []string) *Game {

	guiBoard := gui.New(gui.NewConfig())

	// Przekaż dane planszy do biblioteki warships-lightgui
	err := guiBoard.Import(board)
	if err != nil {
		log.Fatalf("Błąd podczas importowania planszy gry: %v", err)
	}
	return &Game{
		Board: guiBoard,
	}
}

func (g *Game) PlayGame(client *http.Client, url string, token string, board []string) {

	// Inicjalizacja planszy gry
	cfg := gui.NewConfig()
	cfg.HitChar = 'X'
	cfg.HitColor = color.FgRed

	for {
		// Pobierz i wydrukuj czas gry
		timer, err := httpclient.PrintGameTime(client, url, token)
		if err != nil {
			log.Fatalf("Błąd podczas pobierania statusu gry: %v", err)
		}

		if timer == 0 {
			fmt.Println("Koniec czasu na strzał")
			httpclient.AbandomGame(client, url, token)
			return
		}

		fmt.Print("Podaj koordynat (np. C8): ")

		coordCh := make(chan string)
		emptyCh := make(chan bool)

		go func() {
			var coord string
			fmt.Scanln(&coord)

			if len(coord) < 1 {
				emptyCh <- true
				return
			}

			coordCh <- coord
		}()

		select {
		case coord := <-coordCh:
			// Przekształć koordynat na wielkie litery
			coord = strings.ToUpper(strings.TrimSpace(coord))

			// Wyślij ogień do podanych koordynatów
			result, err := httpclient.SendFire(client, url, token, coord)
			if err != nil {
				log.Fatalf("Błąd podczas strzelania: %v", err)
			}

			// Jeśli wynik to "miss", zakończ pętlę
			if result == "hit" {
				g.Board.Set(gui.Right, coord, gui.Hit)
				g.Board.Display()
				// Zaznacz trafienie na planszy przeciwnika
			}

			// Jeśli wynik to "miss", zakończ pętlę
			if result != "hit" {

				g.Board.Set(gui.Right, coord, gui.Miss) // Zaznacz pudło na planszy przeciwnika
				g.Board.Display()

				//Wyswietl strzaly przeciwnika
				shots, err := httpclient.GetOpponentShots(client, url, token)
				if err != nil {
					log.Fatalf("Błąd podczas pobierania strzałów przeciwnika: %v", err)
				}

				if len(shots) > 0 {
					fmt.Println("Przeciwnik strzelał. Strzały: ", shots)
					for _, shot := range shots {
						if Contains(board, shot) {
							//fmt.Println("Przeciwnik trafił w: ", shot)
							g.Board.Set(gui.Left, shot, gui.Hit)
							g.Board.Display()
						} else {
							//fmt.Println("Przeciwnik nie trafił.")
							g.Board.Set(gui.Left, shot, gui.Miss)
							g.Board.Display()
						}
					}
				} else {
					fmt.Println("Przeciwnik jeszcze nie strzelał.")
				}

				for {
					shouldFire, err := httpclient.PrintShouldFire(client, url, token)
					if err != nil {
						log.Fatalf("Błąd podczas pobierania statusu gry: %v", err)
					}
					time.Sleep(2 * time.Second)

					if shouldFire {
						break
					}
				}
			}

		case <-emptyCh:
			// Jeśli użytkownik naciśnie Enter bez wprowadzania danych, natychmiast kontynuuj
			continue

		case <-time.After(10 * time.Second):
			// Jeśli użytkownik nie wprowadził danych w ciągu 10 sekund, kontynuuj
		}

	}
}

// contains sprawdza, czy slice zawiera określony element.
func Contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
