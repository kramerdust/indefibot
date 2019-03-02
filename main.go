package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	ex "github.com/kramerdust/indefibot/exegete"

	"github.com/kramerdust/indefibot/bot"
	"gopkg.in/yaml.v2"
)

func main() {
	proxyFlag := flag.Bool("proxy", false, "Use proxy")
	flag.Parse()

	log.SetOutput(os.Stdout)

	config, err := loadYAMLConfig("app.yaml")
	if err != nil {
		panic(err)
	}

	var myBot *bot.Bot
	if *proxyFlag {
		myBot, err = bot.NewBotWithProxy(config, bot.NewWordMap())
	} else {
		myBot, err = bot.NewBot(config, bot.NewWordMap())
	}
	if err != nil {
		panic(err)
	}

	provider := ex.NewOxfExpositorProvider(config.Source.AppID, config.Source.AppKey)
	myBot.SetExpositorProvider(provider)

	go myBot.Start()

	fmt.Println("Press enter to stop")
	fmt.Scanln()

}

func loadYAMLConfig(filename string) (*bot.Config, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	ymlConfig := &bot.Config{}
	err = yaml.Unmarshal(bytes, ymlConfig)
	if err != nil {
		return nil, err
	}
	return ymlConfig, nil
}
