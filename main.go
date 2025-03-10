package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var cfg *BotConfig
var stopWords map[string]bool

type BotConfig struct {
	TokenFile     string `json:"token_file"`
	StopWordsFile string `json:"stop_words_file"`
	MessagesFile  string `json:"messages_file"`
}

func CreateConfig() *BotConfig {
	return &BotConfig{
		TokenFile:     "token.txt",
		StopWordsFile: "stopwords-ru.json",
		MessagesFile:  "chats/messages.json",
	}
}

func CreateEngine(cfg *BotConfig, msgs *[]MetaMessage) (*tgbotapi.BotAPI, error) {
	token, err := os.ReadFile(cfg.TokenFile)
	if err != nil {
		return nil, err
	}
	bot, err := tgbotapi.NewBotAPI(string(token))
	if err != nil {
		return nil, err
	}

	*msgs, err = ScanMessages(cfg.MessagesFile)
	if err != nil {
		return nil, err
	}

	return bot, nil
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

func main() {
	cfg = CreateConfig()

	allMessages := make([]MetaMessage, 0)

	bot, err := CreateEngine(cfg, &allMessages)
	if err != nil {
		fmt.Println(err)
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	err = readStopWords(cfg.StopWordsFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("bot started")

	stopBot := make(chan bool, 1)

	numThreads := 10
	threads := sync.WaitGroup{}
	threads.Add(numThreads)

	go updateMessages(&allMessages)

	for range numThreads {
		go func() {
			defer threads.Done()
			handleInput(updates, bot, &allMessages, stopBot)
		}()
	}

	time.Sleep(time.Hour * 10)
}
