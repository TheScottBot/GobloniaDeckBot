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
var Token = ""
var BotPrefix = "??"

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

func GetUserDeckPointer(theDeck *Deck) {
	url := "https://www.deckofcardsapi.com/api/deck/new/shuffle/?deck_count=1"

	err := GetJson(url, theDeck)
	if err != nil {
		fmt.Printf("error getting deck: %s\n", err.Error())
	} else {
		fmt.Printf("Deck id: %s\n", theDeck.DeckID)
	}
}

func GetUserCardsPointer(theDeck *Deck) (string, string) {
	url := "https://www.deckofcardsapi.com/api/deck/" + theDeck.DeckID + "/draw/?count=1"

	err := GetJson(url, theDeck)
	if err != nil {
		fmt.Printf("error getting deck: %s\n", err.Error())
	}

	return theDeck.Cards[0].Suit, theDeck.Cards[0].Value
}

var Hand []Card

func messageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == BotId {
		return
	}

	if message.Content == BotPrefix+"ping" {
		_, err := session.ChannelMessageSend(message.ChannelID, "pong")
		if err != nil {
			fmt.Println("error ", err)
		}
	}
	HandleNewDeck(message, session)
	HandleDraw(message, session)
	HandlePM(session, message)
	HandleFinalHand(message, session)
}

func HandleFinalHand(message *discordgo.MessageCreate, session *discordgo.Session) {
	if strings.HasPrefix(message.Content, BotPrefix+"FinalHand") {
		if _, exists := UsersDecks[message.Author.ID]; exists {
			handList := HandOrder(StupidFuncNeedToRemove(message.Content))
			output := ChooseHand(handList, message.Author.ID)
			_, err := session.ChannelMessageSend(message.ChannelID, output)
			if err != nil {
				fmt.Println("error ", err)
			}
		}
	}
}

func HandleDraw(message *discordgo.MessageCreate, session *discordgo.Session) {
	if strings.HasPrefix(message.Content, BotPrefix+"Draw") {
		if _, exists := UsersDecks[message.Author.ID]; exists {
			var howMany = 1

			if len(message.Content) > 6 {
				howMany, _ = strconv.Atoi(string(message.Content[6]))
			}

			if UsersDecks[message.Author.ID].DeckID == "" {
				return
			}

			for i := 0; i < howMany; i++ {
				suit, value := GetUserCardsPointer(UsersDecks[message.Author.ID])
				card := Card{
					Suit:  suit,
					Value: value,
				}

				UsersHandP[message.Author.ID] = append(UsersHandP[message.Author.ID], card)
			}

			_, err := session.ChannelMessageSend(message.ChannelID, "Your have drawn "+strconv.Itoa(len(UsersHandP[message.Author.ID]))+" cards")
			if err != nil {
				fmt.Println("error ", err)
			}
		} else {
			fmt.Println("You don't have a deck to draw from. Use ??NewDeck first")
		}
	}
}

var UsersHandP = make(map[string][]Card)
var UsersDecks = make(map[string]*Deck)

func HandleNewDeck(message *discordgo.MessageCreate, session *discordgo.Session) {
	if message.Content == BotPrefix+"NewDeck" {
		if _, exists := UsersDecks[message.Author.ID]; exists {
			GetUserDeckPointer(UsersDecks[message.Author.ID])

			// GetUserDeck()
			_, err := session.ChannelMessageSend(message.ChannelID, "Your new deck is ready, it's ID is: "+UsersDecks[message.Author.ID].DeckID)
			if err != nil {
				fmt.Println("error ", err)
			}
		} else {
			newDeck := Deck{}
			UsersDecks[message.Author.ID] = &newDeck

			GetUserDeckPointer(UsersDecks[message.Author.ID])

			// GetUserDeck()
			_, err := session.ChannelMessageSend(message.ChannelID, "Your new deck is ready, it's ID is: "+UsersDecks[message.Author.ID].DeckID)
			if err != nil {
				fmt.Println("error ", err)
			}
		}
	}
}

func StupidFuncNeedToRemove(original string) string {
	var re = regexp.MustCompile(`(?i)(\d)`)

	var stringBuilder strings.Builder
	for _, match := range re.FindAllString(original, -1) {
		stringBuilder.WriteString(match + ",")
	}
	return stringBuilder.String()
}

var ChosenHand []Card

func ChooseHand(cards []int, userID string) string {
	if cards != nil {
		for _, element := range cards {
			ChosenHand = append(ChosenHand, UsersHandP[userID][element])
		}
		UsersHandP[userID] = nil
	}

	var stringBuilder strings.Builder
	for _, element := range ChosenHand {
		stringBuilder.WriteString("You revealed: " + element.Suit + " " + element.Value + "\n")
	}
	ChosenHand = nil
	return stringBuilder.String()
}

func HandOrder(cardsDelimitedList string) []int {
	cardsDelimitedList = cardsDelimitedList[:len(cardsDelimitedList)-1]
	inputSliced := strings.Split(cardsDelimitedList, ",")
	converted := make([]int, len(inputSliced))
	for index, value := range inputSliced {
		if len(strings.TrimSpace(value)) > 0 {
			output, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				fmt.Println("Theres an error, mate: " + err.Error())
			}
			converted[index] = output
		}
	}
	return converted
}

func HandlePM(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Content == BotPrefix+"PM" {
		var stringBuilder strings.Builder
		for index, element := range UsersHandP[message.Author.ID] {
			stringBuilder.WriteString("Your card is " + element.Suit + " " + element.Value + ". Card reference number = " + strconv.Itoa(index) + "\n")
		}
		// We create the private channel with the user who sent the message.
		channel, err := session.UserChannelCreate(message.Author.ID)
		if err != nil {
			// If an error occurred, we failed to create the channel.
			//
			// Some common causes are:
			// 1. We don't share a server with the user (not possible here).
			// 2. We opened enough DM channels quickly enough for Discord to
			//    label us as abusing the endpoint, blocking us from opening
			//    new ones.
			fmt.Println("error creating channel:", err)
			session.ChannelMessageSend(
				message.ChannelID,
				"Something went wrong while sending the DM!",
			)
			return
		}
		// Then we send the message through the channel we created.
		_, err = session.ChannelMessageSend(channel.ID, stringBuilder.String())
		if err != nil {
			// If an error occurred, we failed to send the message.
			//
			// It may occur either when we do not share a server with the
			// user (highly unlikely as we just received a message) or
			// the user disabled DM in their settings (more likely).
			fmt.Println("error sending DM message:", err)
			session.ChannelMessageSend(
				message.ChannelID,
				"Failed to send you a DM. "+
					"Did you disable DM in your privacy settings?",
			)
		}
	}
}

func main() {
	client = &http.Client{Timeout: 10 * time.Second}

	// https://www.deckofcardsapi.com/api/deck/<<deck_id>>/draw/?count=2
	Start()
}
