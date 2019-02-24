package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	ex "github.com/kramerdust/indefibot/exegete"

	"github.com/kramerdust/indefibot/bot"
	"gopkg.in/yaml.v2"
)

func main() {
	proxyFlag := flag.Bool("proxy", false, "Use proxy")
	flag.Parse()

	config, err := loadYAMLConfig("app.yaml")
	if err != nil {
		panic(err)
	}

	var myBot *bot.Bot
	if *proxyFlag {
		myBot, err = bot.NewBotWithProxy(config)
	} else {
		myBot, err = bot.NewBot(config)
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
