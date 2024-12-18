package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethanhosier/clips/openai"

	"github.com/joho/godotenv"
)

const id = "JFh7vQEoqX4"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	oc := openai.NewOpenaiClient()

	transcription, err := oc.CreateTranscription(context.Background(), "input.mp4", openai.OpenaiAudioResponseFormatVerboseJSON)
	if err != nil {
		fmt.Printf(err.Error())
	}

	fmt.Printf("%+v\n", *transcription)

}
