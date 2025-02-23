package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	russian "github.com/kljensen/snowball/russian"
)

func ScanMessages(filepath string) ([]MetaMessage, error) {
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

func AddMessageToJSON(newMessage string) error {
	existingJSON, err := os.ReadFile(cfg.MessagesFile)
	if err != nil {
		return fmt.Errorf("failed to read existing JSON: %v", err)
	}

	var messages []Message
	err = json.Unmarshal(existingJSON, &messages)
	if err != nil {
		return fmt.Errorf("failed to unmarshal existing JSON: %v", err)
	}

	newMsg := Message{Message: newMessage}
	messages = append(messages, newMsg)

	updatedJSON, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated JSON: %v", err)
	}

	err = os.WriteFile(cfg.MessagesFile, updatedJSON, 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated JSON: %v", err)
	}

	return nil
}
