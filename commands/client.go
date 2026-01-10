package commands

import (
	"fmt"

	"github.com/amatsagu/tempest"
)

// A Client is a basic wrapper around a Tempest HTTP client.
type Client struct {
	*tempest.HTTPClient

	guildID tempest.Snowflake
}

var GlobalClient *Client

// NewClient creates a new Client wrapping the given Tempest HTTP client and guild ID.
func NewClient(httpClient *tempest.HTTPClient, guildID tempest.Snowflake) *Client {
	GlobalClient = &Client{
		HTTPClient: httpClient,
		guildID:    guildID,
	}
	return GlobalClient
}

// RegisterDefaultCommands registers all default commands to the Tempest HTTP client.
func (c *Client) RegisterDefaultCommands() error {
	for _, cmd := range defaultCommands {
		if err := cmd.Register(c.HTTPClient); err != nil {
			return fmt.Errorf("failed to register command %s: %w", cmd.Name, err)
		}
	}

	return c.SyncCommandsWithDiscord([]tempest.Snowflake{c.guildID}, nil, false)
}
