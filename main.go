package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
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

var Hand []Card

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
	if strings.HasPrefix(m.Content, "??Draw") {
		var howMany = 1

		leng := len(m.Content)

		fmt.Println("Length = " + strconv.Itoa(leng))

		if len(m.Content) > 6 {
			howMany, _ = strconv.Atoi(string(m.Content[6]))
		}

		fmt.Println("users deck: ", userDeck.DeckID)
		if userDeck.DeckID == "" {
			return
		}

		for i := 0; i < howMany; i++ {
			suit, value := GetUserCards(userDeck.DeckID)
			card := Card{
				Suit:  suit,
				Value: value,
			}

			Hand = append(Hand, card)
		}

		_, err := s.ChannelMessageSend(m.ChannelID, "Your have drawn "+strconv.Itoa(len(Hand))+" cards")
		if err != nil {
			fmt.Println("error ", err)
		}
	}
	if m.Content == "??PM" {
		PM(s, m)
	}
	if strings.HasPrefix(m.Content, "??FinalHand") {

		handList := HandOrder(StupidFuncNeedToRemove(m.Content))
		output := ChooseHand(handList)
		_, err := s.ChannelMessageSend(m.ChannelID, output)
		if err != nil {
			fmt.Println("error ", err)
		}
	}
}

func StupidFuncNeedToRemove(original string) string {
	var re = regexp.MustCompile(`(?i)(\d)`)

	var stringBuilder strings.Builder
	for i, match := range re.FindAllString(original, -1) {
		stringBuilder.WriteString(match + ",")
		fmt.Println(match, "found at index", i)
	}
	fmt.Println("final built string " + stringBuilder.String())
	return stringBuilder.String()
}

var ChosenHand []Card

func ChooseHand(cards []int) string {
	if cards != nil {
		for _, element := range cards {
			fmt.Println("Element: " + strconv.Itoa(element))
			ChosenHand = append(ChosenHand, Hand[element])
		}
		Hand = nil
	}

	var stringBuilder strings.Builder
	for _, element := range ChosenHand {
		stringBuilder.WriteString("You revealed: " + element.Suit + " " + element.Value + "\n")
	}
	ChosenHand = nil
	return stringBuilder.String()
}

func HandOrder(cardsDelimitedList string) []int {
	fmt.Println("Just endered HandOrder. CardsDelimitedList is: " + cardsDelimitedList)
	cardsDelimitedList = cardsDelimitedList[:len(cardsDelimitedList)-1]
	fmt.Println("Just fucked with string. CardsDelimitedList is: " + cardsDelimitedList)
	inputSliced := strings.Split(cardsDelimitedList, ",")
	converted := make([]int, len(inputSliced))
	for index, value := range inputSliced {
		fmt.Println(len(strings.TrimSpace(value)))
		if len(strings.TrimSpace(value)) > 0 {
			fmt.Println("Index: " + strconv.Itoa(index) + " Value: " + value)
			output, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				fmt.Println("Theres a fucking error, mate: " + err.Error())
			}
			converted[index] = output
		}
	}
	return converted
}

func PM(s *discordgo.Session, m *discordgo.MessageCreate) {
	var stringBuilder strings.Builder
	for index, element := range Hand {
		stringBuilder.WriteString("Your card is " + element.Suit + " " + element.Value + ". Card reference number = " + strconv.Itoa(index) + "\n")
	}
	// We create the private channel with the user who sent the message.
	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		// If an error occurred, we failed to create the channel.
		//
		// Some common causes are:
		// 1. We don't share a server with the user (not possible here).
		// 2. We opened enough DM channels quickly enough for Discord to
		//    label us as abusing the endpoint, blocking us from opening
		//    new ones.
		fmt.Println("error creating channel:", err)
		s.ChannelMessageSend(
			m.ChannelID,
			"Something went wrong while sending the DM!",
		)
		return
	}
	// Then we send the message through the channel we created.
	_, err = s.ChannelMessageSend(channel.ID, stringBuilder.String())
	if err != nil {
		// If an error occurred, we failed to send the message.
		//
		// It may occur either when we do not share a server with the
		// user (highly unlikely as we just received a message) or
		// the user disabled DM in their settings (more likely).
		fmt.Println("error sending DM message:", err)
		s.ChannelMessageSend(
			m.ChannelID,
			"Failed to send you a DM. "+
				"Did you disable DM in your privacy settings?",
		)
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
