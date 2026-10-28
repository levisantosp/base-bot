package env

import (
	"log"

	"github.com/Oudwins/zog"
	"github.com/Oudwins/zog/zenv"
	"github.com/joho/godotenv"
)

type TEnv struct {
	BotToken string `zog:"BOT_TOKEN"`
	Prefix   string `zog:"PREFIX"`
}

var Data TEnv

func LoadEnv() {
	_ = godotenv.Overload()

	schema := zog.Struct(zog.Shape{
		"BotToken": zog.String().Required(),
		"Prefix":   zog.String().Required(),
	})

	err := schema.Parse(zenv.NewDataProvider(), &Data)
	if err != nil {
		log.Fatal(err)
	}
}
