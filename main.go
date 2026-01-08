package main

import (
	"log"
	"net/http"
	"os"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/bot"
)

func main() {
	log.Println("Creating new Tempest client...")
	client := tempest.NewHTTPClient(tempest.HTTPClientOptions{
		BaseClientOptions: tempest.BaseClientOptions{
			Token: os.Getenv("DISCORD_BOT_TOKEN"),
		},
		PublicKey: os.Getenv("DISCORD_PUBLIC_KEY"),
	})

	addr := os.Getenv("LISTENING_ADDRESS")
	guildID, err := tempest.StringToSnowflake(os.Getenv("DISCORD_GUILD_ID"))
	if err != nil {
		log.Fatal("failed to parse guild ID env variable:", err)
	}

	b, err := bot.NewBot(client, guildID)
	if err != nil {
		log.Fatal("failed to create bot:", err)
	}
	defer b.Close()

	if err := b.RegisterDefaultCommands(); err != nil {
		log.Fatal("failed to register default commands:", err)
	}

	http.HandleFunc("POST /discord/callback", client.DiscordRequestHandler)

	log.Printf("Serving application at: %s/discord/callback\n", addr)

	if err = http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("the shit has hit the fan", err)
	}
}
