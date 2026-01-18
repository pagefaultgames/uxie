package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/oranguru/commands"
	"github.com/pagefaultgames/oranguru/db"
	"github.com/pagefaultgames/oranguru/utils"
)

const (
	// os.Getenv("DISCORD_BOT_TOKEN")
	DISCORD_BOT_TOKEN = "MTQ2MDExODAzMjE0MjY5NjUxMA.Gqitss.QCDPINoSHlYuuKpWcd_HNBimSDGEQnuA-b9ZYE"
	// os.Getenv("DISCORD_PUBLIC_KEY")
	DISCORD_PUBLIC_KEY = "c4151319f2cd857e73153949f58dd05151cd4363f42201051c7c004ece35787d"

	DISCORD_GUILD_ID = "1460127335649902592"

	ADDRESS = "localhost:8080"
)

func main() {
	slog.Info("Creating new Tempest client...")
	httpClient := tempest.NewHTTPClient(tempest.HTTPClientOptions{
		BaseClientOptions: tempest.BaseClientOptions{
			Token: DISCORD_BOT_TOKEN,
		},
		Trace: true,
		PublicKey: DISCORD_PUBLIC_KEY,
	})

	guildID, err := tempest.StringToSnowflake(DISCORD_GUILD_ID)
	if err != nil {
		utils.ErrorAttrs(
			"Failed to parse guild ID "+DISCORD_GUILD_ID,
			slog.Any("error", err),
		)
		panic(fmt.Sprintf("failed to parse guild ID env variable: %v\n", err))
	}

	client := commands.NewClient(httpClient)
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

	if err := client.RegisterDefaultCommands(guildID); err != nil {
		utils.ErrorAttrs("Failed to register default commands", slog.Any("error", err))
		panic(fmt.Sprintf("failed to register default commands: %v\n", err))
	}

	serveHTTP(client)
}

func serveHTTP(client *commands.Client) {
	http.HandleFunc("POST /discord/callback", client.DiscordRequestHandler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	server := &http.Server{Addr: ADDRESS}
	slog.Info("Serving application at: " + ADDRESS + "/discord/callback")

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			utils.ErrorAttrs("error during HTTP serving", slog.Any("error", err))
		}
		stop()
	}()

	<-ctx.Done()
	stop()
	slog.Info("Termination signal received, shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		utils.ErrorAttrs("error shutting down server", slog.Any("error", err))
	}

	slog.Info("Server shut down successfully")
}
