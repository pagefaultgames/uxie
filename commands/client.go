package commands

import (
	"fmt"

	"github.com/amatsagu/tempest"
)

type Client struct {
	*tempest.HTTPClient

	guildID tempest.Snowflake
}

var GlobalClient *Client

// NewClient creates a new Command Client wrapping the given Tempest HTTP client and guild ID.

func NewClient(httpClient *tempest.HTTPClient, guildID tempest.Snowflake) *Client {
	if GlobalClient != nil {
		return GlobalClient
	}
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
	return nil
}
