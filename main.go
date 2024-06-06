package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/robfig/cron/v3"
)

var referenceHeight int
var notified bool = false

func readEnvironmentInteger(name string) int {
	parameter, err := strconv.Atoi(os.Getenv(name))
	if err != nil {
		log.Fatal(err, ". Failed to read environment variable: ", name)
	}
	return parameter
}

var MAX_DROP_HEIGHT = readEnvironmentInteger("MAX_DROP_HEIGHT")
var SECRET_PLAYERS = readEnvironmentInteger("SECRET_PLAYERS")
var PUSHME_ID = os.Getenv("PUSHME_ID")
var MYNOTIFIER_API_KEY = os.Getenv("MYNOTIFIER_API_KEY")

type LeaderboardEntry struct {
	Rank   int     `json:"rank"`
	Name   string  `json:"name"`
	Height float32 `json:"height"`
}

type LiveHeightsEntry struct {
	Rank   int     `json:"rank"`
	Name   string  `json:"display_name"`
	Height float32 `json:"height"`
}

type NotifyPayload struct {
	TriggerID string `json:"triggerId"`
	Text      string `json:"text"`
}

func getReferenceHeight() {
	resp, err := http.Get("https://dips-plus-plus.xk.io/leaderboard/global")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data []LeaderboardEntry
	json.Unmarshal(body, &data)
	log.Printf("Leaderboard: %v...", data[:10])
	referenceHeight = int(data[SECRET_PLAYERS].Height) - MAX_DROP_HEIGHT
	log.Printf("Waiting for players on height %d", referenceHeight)

}

func checkHeight() {
	resp, err := http.Get("https://dips-plus-plus.xk.io/live_heights/global")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data []LiveHeightsEntry
	json.Unmarshal(body, &data)

	nr_players_to_display := 10
	if len(data) < nr_players_to_display {
		nr_players_to_display = len(data)
	}

	var drivingInfo = fmt.Sprintf("%d:", len(data))
	// log.Printf("%d driving: %v...", len(data), data[:nr_players_to_display])
	for index, player := range data {
		drivingInfo += fmt.Sprintf(" %d %s %d", player.Rank, player.Name, int(player.Height))
		if index == nr_players_to_display-1 {
			if len(data) > nr_players_to_display {
				drivingInfo += "..."
			}
			break
		}
		drivingInfo += ","
	}

	if len(data) > 0 {
		if int(data[0].Height) >= referenceHeight {
			if !notified {
				log.Printf("%s's height: %f > %d", data[0].Name, data[0].Height, referenceHeight)
				notify(fmt.Sprintf("%v... (%d)", data[:nr_players_to_display], len(data)))
				notified = true
			}
		} else if notified {
			log.Printf("%s's height: %f < %d", data[0].Name, data[0].Height, referenceHeight)
			notify(fmt.Sprintf("%d: %v...", len(data), data[:nr_players_to_display]))
			notified = false
		}
	} else {
		drivingInfo += " No driving players found"
	}
	log.Print(drivingInfo)
}

func notify(infoText string) {
	payload := NotifyPayload{
		TriggerID: PUSHME_ID,
		Text:      infoText,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	http.Post("https://pushme.win/trigger", "application/json", bytes.NewBuffer(payloadBytes))
	http.PostForm("https://api.mynotifier.app", url.Values{
		"apiKey":  []string{MYNOTIFIER_API_KEY}, // This is your own private key
		"message": []string{infoText},           // Could be anything
		"type":    []string{"info"},             // info, error, warning or success
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Running with arguments: MAX_DROP_HEIGHT=%d, SECRET_PLAYERS=%d", MAX_DROP_HEIGHT, SECRET_PLAYERS)

	getReferenceHeight()
	checkHeight()

	c := cron.New()
	c.AddFunc("@daily", getReferenceHeight)
	c.AddFunc("@every 2m", checkHeight)
	c.Start()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.ListenAndServe(":8080", nil)
}
