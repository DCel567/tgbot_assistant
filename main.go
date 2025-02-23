package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	russian "github.com/kljensen/snowball/russian"
)

var stopWords map[string]bool

type Message struct {
	Message string `json:"message"`
}

type MetaMessage struct {
	Text     string          `json:"text"`
	TagCloud map[string]bool `json:"tag_cloud"`
}

func getTagCloud(text string) []string {
	var tags []string

	for _, word := range strings.Split(text, " ") {
		word = russian.Stem(word, true)
		if !stopWords[word] {
			tags = append(tags, word)
		}
	}

	return tags
}

func findMessage(text string, mss []MetaMessage, n int) ([]string, bool) {
	tags := getTagCloud(text)
	bestMatch := make([]string, 0, n)
	maxScore := 0

	for _, mm := range mss {
		key_score := 0
		for _, tag := range tags {
			if mm.TagCloud[tag] {
				key_score++
			}
		}

		if key_score == maxScore && key_score > 0 && len(bestMatch) < n {
			bestMatch = append(bestMatch, mm.Text)
		}

		if key_score > maxScore {
			bestMatch = make([]string, 0, n)
			bestMatch = append(bestMatch, mm.Text)
			maxScore = key_score
		}

	}

	if maxScore > 0 {
		return bestMatch, true
	}

	return bestMatch, false
}

func handleInput(update tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, mss []MetaMessage, done chan bool) {
	for {
		select {
		case <-done:
			return
		case update := <-update:
			if update.Message == nil {
				continue
			}

			fmt.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)

			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я ваш новый Telegram-бот.")
				bot.Send(msg)
				continue
			}

			if update.Message.Text == "/help" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я могу помочь вам с основными вопросами! Напишите /start для начала.")
				bot.Send(msg)
				continue
			}

			if responds, ok := findMessage(update.Message.Text, mss, 1); ok {
				for _, r := range responds {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, r)
					bot.Send(msg)
				}
				continue
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "404")
				bot.Send(msg)
				continue
			}
		}
	}
}

func readStopWords(filepath string) error {
	stopWords = make(map[string]bool)
	text, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	var words []string
	err = json.Unmarshal(text, &words)
	if err != nil {
		return err
	}
	for _, word := range words {
		stopWords[word] = true
	}
	return nil
}

func readMessages(filepath string) ([]MetaMessage, error) {
	text, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var mss []Message
	err = json.Unmarshal(text, &mss)
	if err != nil {
		return nil, err
	}

	allMessages := make([]MetaMessage, len(mss))

	for i, m := range mss {
		tags := getTagCloud(m.Message)
		allMessages[i] = MetaMessage{Text: m.Message, TagCloud: make(map[string]bool)}
		for _, tag := range tags {
			allMessages[i].TagCloud[tag] = true
		}
	}

	return allMessages, nil
}

func main() {
	token, err := os.ReadFile("token.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	bot, err := tgbotapi.NewBotAPI(string(token))
	if err != nil {
		fmt.Println(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	filepath := "messages.json"
	allMessages, err := readMessages(filepath)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(allMessages)

	filepath = "stopwords-ru.json"
	err = readStopWords(filepath)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("bot started")

	handleInput(updates, bot, allMessages, make(chan bool))
}
