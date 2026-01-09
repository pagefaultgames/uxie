package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/commands"
	"github.com/pagefaultgames/oranguru/db"
	"github.com/pagefaultgames/oranguru/utils"
)

func main() {
	slog.Info("Creating new Tempest client...")
	httpClient := tempest.NewHTTPClient(tempest.HTTPClientOptions{
		BaseClientOptions: tempest.BaseClientOptions{
			Token: os.Getenv("DISCORD_BOT_TOKEN"),
		},
		PublicKey: os.Getenv("DISCORD_PUBLIC_KEY"),
	})

	addr := os.Getenv("LISTENING_ADDRESS")
	guildID, err := tempest.StringToSnowflake(os.Getenv("DISCORD_GUILD_ID"))
	if err != nil {
		utils.ErrorAttrs(
			"Failed to parse guild ID env var",
			slog.String("env var", os.Getenv("DISCORD_GUILD_ID")),
			slog.String("error", err.Error()),
		)
		panic(fmt.Sprintf("failed to parse guild ID env variable: %v\n", err))
	}

	client := commands.NewClient(httpClient, guildID)
	if err := db.Open(); err != nil {
		utils.ErrorAttrs("Failed to open database", slog.Any("error", err))
		panic(fmt.Sprintf("failed to open database: %v\n", err))
	}

	defer func() {
		if err := db.Close(); err != nil {
			utils.ErrorAttrs("Failed to close database", slog.Any("error", err))
			panic(err)
		}
	}()

	if err := client.RegisterDefaultCommands(); err != nil {
		utils.ErrorAttrs("Failed to register default commands", slog.String("error", err.Error()))
		panic(fmt.Sprintf("failed to register default commands: %v\n", err))
	}

	http.HandleFunc("POST /discord/callback", httpClient.DiscordRequestHandler)

	slog.Info("Serving application at: " + addr + "/discord/callback\n")

	if err = http.ListenAndServe(addr, nil); err != nil {
		utils.ErrorAttrs("the shit has hit the fan", slog.Any("error", err))
	}
}
