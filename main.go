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

type BotConfig struct {
	TokenFile     string `json:"token_file"`
	StopWordsFile string `json:"stop_words_file"`
	MessagesFile  string `json:"messages_file"`
}

func CreateConfig() *BotConfig {
	return &BotConfig{
		TokenFile:     "token.txt",
		StopWordsFile: "stopwords-ru.json",
		MessagesFile:  "messages.json",
	}
}

var stopWords map[string]bool

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
	token, err := os.ReadFile(cfg.TokenFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	bot, err := tgbotapi.NewBotAPI(string(token))
	if err != nil {
		fmt.Println(err)
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	allMessages, err := ScanMessages(cfg.MessagesFile)
	if err != nil {
		fmt.Println(err)
		return
	}

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

	for i := 0; i < numThreads; i++ {
		go func() {
			defer threads.Done()
			handleInput(updates, bot, &allMessages, stopBot)
		}()
	}

	go updateMessages(&allMessages)

	time.Sleep(time.Hour * 10)

	// stopBot <- true
}
