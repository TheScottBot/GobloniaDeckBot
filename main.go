package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Deck struct {
	Success   bool   `json:"success"`
	DeckID    string `json:"deck_id"`
	Cards     []Card `json:"cards"`
	Shuffuled bool   `json:"shuffled"`
	Remaining int    `json:"remaining"`
}

type Card struct {
	Code  string `json:"code"`
	Image string `json:"image"`
	Value string `json:"value"`
	Suit  string `json:"suit"`
}

var client *http.Client
var deck Deck
var Token = ""
var BotPrefix = "?"

func GetCards(deckID string) {
	url := "https://www.deckofcardsapi.com/api/deck/" + deckID + "/draw/?count=1"

	err := GetJson(url, &deck)
	if err != nil {
		fmt.Printf("error getting deck: %s\n", err.Error())
	} else {
		card := deck.Cards[0].Value + " " + deck.Cards[0].Suit
		fmt.Printf("Deck id: %s\n", card)
	}
}

func GetDeck() {
	url := "https://www.deckofcardsapi.com/api/deck/new/shuffle/?deck_count=1"

	err := GetJson(url, &deck)
	if err != nil {
		fmt.Printf("error getting deck: %s\n", err.Error())
	} else {
		fmt.Printf("Deck id: %s\n", deck.DeckID)
	}
}

func GetJson(url string, target interface{}) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

var BotId string

func Start() {
	goBot, err := discordgo.New("Bot " + Token)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	u, err := goBot.User("@me")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	BotId = u.ID

	goBot.AddHandler(messageHandler)

	err = goBot.Open()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	goBot.Close()
}

var userDeck Deck
var userCard Card

func GetUserDeck() {
	url := "https://www.deckofcardsapi.com/api/deck/new/shuffle/?deck_count=1"

	err := GetJson(url, &userDeck)
	if err != nil {
		fmt.Printf("error getting deck: %s\n", err.Error())
	} else {
		fmt.Printf("Deck id: %s\n", userDeck.DeckID)
	}
}

func GetUserCards(deckID string) (string, string) {
	fmt.Println("INSIDE GetUserCards")
	url := "https://www.deckofcardsapi.com/api/deck/" + deckID + "/draw/?count=1"

	err := GetJson(url, &userDeck)
	if err != nil {
		fmt.Printf("error getting deck: %s\n", err.Error())
	}

	fmt.Printf("User's card is: %s\n", userCard)
	return userDeck.Cards[0].Suit, userDeck.Cards[0].Value
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotId {
		fmt.Println("author: ", m.Author.ID, " \n BotId: ", BotId, "\nmessge: ", m.Content)
		return
	}

	if m.Content == "??ping" {
		_, err := s.ChannelMessageSend(m.ChannelID, "pong")
		if err != nil {
			fmt.Println("error ", err)
		}
	}
	if m.Content == "??NewDeck" {
		GetUserDeck()
		_, err := s.ChannelMessageSend(m.ChannelID, "Your new deck is ready, it's ID is: "+userDeck.DeckID)
		if err != nil {
			fmt.Println("error ", err)
		}
	}
	if m.Content == "??Draw" {
		fmt.Println("users deck: ", userDeck.DeckID)
		if userDeck.DeckID == "" {
			return
		}
		suit, value := GetUserCards(userDeck.DeckID)
		_, err := s.ChannelMessageSend(m.ChannelID, "Your card is: "+suit+" "+value)
		if err != nil {
			fmt.Println("error ", err)
		}
	}
}

func main() {
	client = &http.Client{Timeout: 10 * time.Second}

	GetDeck()
	GetCards(deck.DeckID)
	GetCards(deck.DeckID)
	GetCards(deck.DeckID)
	println(deck.Remaining)
	// https://www.deckofcardsapi.com/api/deck/<<deck_id>>/draw/?count=2
	Start()
}
