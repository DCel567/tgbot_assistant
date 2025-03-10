package main

import (
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Message struct {
	Message string
}

type MetaMessage struct {
	Text     string
	TagCloud map[string]bool
}

type Context struct {
	Messages []MetaMessage
	Chat     *tgbotapi.Chat
	File     string
	Updated  bool
}

func handleInput(update tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, mss *[]MetaMessage, done chan bool) {
	messagesToShow := 3

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
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Это бот для заметок. "+
					"Заметки пока правда одни на все чаты и пишу их я руками, но это будет исправлено.")
				bot.Send(msg)
				continue
			}

			if update.Message.Text == "/help" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Создавать заметки пока нельзя. "+
					"Введите одно или несколько ключевых слов для поиска нужной заметки с окончанием слов можно ошибиться. "+
					"Чем в большее количество слов вы попадете, тем больше шанс найти нужную заметку")
				bot.Send(msg)
				continue
			}

			if len(update.Message.Text) > 5 && update.Message.Text[:4] == "/add" {

				text := update.Message.Text[5:]
				if len(getTagCloud(text)) == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Заметка не создана, слишком неинформативно")
					bot.Send(msg)
					continue
				}

				jsonMutex := sync.Mutex{}
				jsonMutex.Lock()
				defer jsonMutex.Unlock()

				AddMessageToJSON(text)

				var err error
				*mss, err = ScanMessages(cfg.MessagesFile)
				if err != nil {
					fmt.Println(err)
					continue
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Заметка создана")
				bot.Send(msg)
				continue
			}

			if responds, ok := findMessage(update.Message.Text, *mss, messagesToShow); ok {
				for _, r := range responds {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, r)
					bot.Send(msg)
				}
				continue
			}
		}
	}
}

func updateMessages(mss *[]MetaMessage) {
	ticker := time.NewTicker(time.Minute * 10)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("10 mins scan")
		jsonMutex := sync.Mutex{}
		jsonMutex.Lock()

		var err error
		*mss, err = ScanMessages(cfg.MessagesFile)
		if err != nil {
			fmt.Println(err)
			continue
		}
		continue
	}
}
