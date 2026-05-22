// SPDX-FileCopyrightText: 2026 Pagefault Games
// SPDX-FileContributor: Bertie690
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/amatsagu/tempest"
	"github.com/pagefaultgames/uxie/commands"
	"github.com/pagefaultgames/uxie/db"
	"github.com/pagefaultgames/uxie/utils"
)

var (
	DISCORD_BOT_TOKEN  = os.Getenv("DISCORD_BOT_TOKEN")
	DISCORD_PUBLIC_KEY = os.Getenv("DISCORD_PUBLIC_KEY")
	DISCORD_GUILD_ID   = os.Getenv("DISCORD_GUILD_ID")
	ADDRESS            = os.Getenv("ADDRESS")
	DB_DSN             = os.Getenv("DB_DSN")
)

func main() {
	slog.Info("Creating new Tempest client...")

	httpClient := tempest.NewHTTPClient(tempest.HTTPClientOptions{
		BaseClientOptions: tempest.BaseClientOptions{
			Token: DISCORD_BOT_TOKEN,
			PreCommandHook: func(cmd tempest.Command, ctx *tempest.CommandInteraction) bool {
				slog.Info("Received command interaction",
					slog.String("username", ctx.BaseUser().Username),
					slog.Uint64("ID", uint64(ctx.ID)),
					slog.String("commandName", ctx.Data.Name),
				)
				return true
			},
		},
		Trace:     true,
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	if err := db.Open(ctx, DB_DSN); err != nil {
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

	http.HandleFunc("POST /discord/callback", client.DiscordRequestHandler)

	server := &http.Server{Addr: ADDRESS}
	slog.Info("Serving application at: " + ADDRESS + "/discord/callback")

	go func() {
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
			stop()
			return
		}

		if err == nil {
			panic("invariant of listenAndServe violated: returned nil error")
		}
		if strings.Contains(err.Error(), "address already in use") {
			utils.ErrorAttrs(
				"TCP address already in use. Make sure no other instance of the server is running and that the port is free.",
				slog.Any("error", err),
			)
		} else {
			utils.ErrorAttrs("error during HTTP serving", slog.Any("error", err))
		}
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
