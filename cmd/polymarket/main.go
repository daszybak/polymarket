package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/daszybak/polymarket/internal/polymarket/gamma"
	"github.com/daszybak/polymarket/internal/polymarket/websocket"
	"go.yaml.in/yaml/v4"
)

type config struct {
	PolyMarket struct {
		WebsocketURL string `yaml:"ws_url"`
		GammaURL     string `yaml:"gamma_url"`
		ClobURL      string `yaml:"clob_url"`
		Events       []struct {
			Slug string `yaml:"slug"`
		} `yaml:"events"`
	} `yaml:"polymarket"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// TODO The config path should be a parameter (argument) of the main function.
	rawConfig, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Couldn't read config: %v", err)
	}

	cfg := &config{}
	if err = yaml.Unmarshal(rawConfig, cfg); err != nil {
		log.Fatalf("Couldn't parse config: %v", err)
	}

	gammaClient := gamma.New(cfg.PolyMarket.GammaURL)

	markets := make([]*gamma.Market, 0)
	for _, m := range cfg.PolyMarket.Events {
		event := &gamma.Event{}
		event, err = gammaClient.GetEventBySlug(m.Slug)
		if err != nil {
			log.Printf("Error fetching event %s: %v", m.Slug, err)
			continue
		}
		markets = append(markets, event.Markets...)
	}

	rawMarkets, _ := json.Marshal(markets)
	log.Printf("markets: %s", rawMarkets)

	ws, err := websocket.New(ctx, cfg.PolyMarket.WebsocketURL+"/market")
	if err != nil {
		log.Fatalf("Couldn't open websocket connection: %v", err)
	}
	defer ws.Close(ctx)

	log.Println("Connected successfully")

	tokenIDs := make([]string, 0, len(markets))
	for _, m := range markets {
		tokenIDs = append(tokenIDs, m.ClobTokenIDs...)
	}

	if err := ws.SubscribeMarket(ctx, tokenIDs, true, nil); err != nil {
		log.Fatalf("Couldn't send subscription: %v", err)
	}

	for {
		msg, err := ws.ReadMessage(ctx)
		if err != nil {
			log.Fatalf("Couldn't read message: %v", err)
		}
		log.Printf("message: %s", msg)
	}
}
