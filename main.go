package main

import (
	"context"
	"fmt"
	"github.com/PullRequestInc/go-gpt3"
	"log"
	_ "os"
)

func getCommand(word string) {

	apiKey := "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"
	if apiKey == "" {
		log.Fatalln("Missing API KEY")
	}

	ctx := context.Background()
	client := gpt3.NewClient(apiKey)

	resp, err := client.CompletionWithEngine(ctx, "text-davinci-001", gpt3.CompletionRequest{
		Prompt:           []string{word},
		MaxTokens:        gpt3.IntPtr(100),
		Temperature:      gpt3.Float32Ptr(0),
		FrequencyPenalty: float32(0.2),
		PresencePenalty:  float32(0),
		TopP:             gpt3.Float32Ptr(1),
	})
	if err != nil {
		log.Fatalln(err)
		fmt.Println("Hmm, something ins't right.")
	}
	//fmt.Println(resp)
	fmt.Println(resp.Choices[0].Text)
}

func main() {
	getCommand("install go using brew")
}
