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

	serveHTTP(addr, client)
}

func serveHTTP(addr string, client *commands.Client) {
	http.HandleFunc("POST /discord/callback", client.DiscordRequestHandler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	server := &http.Server{Addr: addr}
	slog.Info("Serving application at: " + addr + "/discord/callback")

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
