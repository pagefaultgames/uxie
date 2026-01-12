package commands

import (
	"fmt"
	"time"

	"github.com/amatsagu/tempest"
)

// A Client is a basic wrapper around a Tempest HTTP client.
type Client struct {
	*tempest.HTTPClient
}

var GlobalClient *Client

// NewClient creates a new Client wrapping the given Tempest HTTP client.
func NewClient(httpClient *tempest.HTTPClient) *Client {
	GlobalClient = &Client{
		HTTPClient: httpClient,
	}
	return GlobalClient
}

const commandRegistrationDelay = 500 * time.Millisecond

// RegisterDefaultCommands registers all default commands to the Tempest HTTP client
// and updates it them to the given guild.
func (c *Client) RegisterDefaultCommands(guildID tempest.Snowflake) error {
	for _, cmd := range defaultCommands {
		if err := cmd.Register(c.HTTPClient); err != nil {
			return fmt.Errorf("failed to register command %s: %w", cmd.Name, err)
		}
		time.Sleep(commandRegistrationDelay)
	}

	return c.SyncCommandsWithDiscord([]tempest.Snowflake{guildID}, nil, false)
}
