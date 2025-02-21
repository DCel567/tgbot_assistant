package main

import (
	"encoding/json"
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Question struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type Questions struct {
	Questions []Question `json:"questions"`
}

func handleInput(update tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, qs map[string]string, done chan bool) {
	for {
		select {
		case <-done:
			return
		case update := <-update:
			if update.Message == nil { // Игнорируем не сообщения
				continue
			}

			// Выводим текст полученного сообщения
			fmt.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)

			// Ответ на команду /start
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я ваш новый Telegram-бот.")
				bot.Send(msg)
			}

			if qs[update.Message.Text] != "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, qs[update.Message.Text])
				bot.Send(msg)
			}

			// Ответ на команду /help
			if update.Message.Text == "/help" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я могу помочь вам с основными вопросами! Напишите /start для начала.")
				bot.Send(msg)
			}
		}
	}
}

func main() {
	token, err := os.ReadFile("token.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(token))
	bot, err := tgbotapi.NewBotAPI(string(token))
	if err != nil {
		fmt.Println(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	filepath := "questions.json"
	text, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println(err)
		return
	}

	qs := Questions{}
	err = json.Unmarshal(text, &qs)
	if err != nil {
		fmt.Println(err)
	}

	questions := make(map[string]string)

	for _, q := range qs.Questions {
		questions[q.Question] = q.Answer
	}

	handleInput(updates, bot, questions, make(chan bool))
}
