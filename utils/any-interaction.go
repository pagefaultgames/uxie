package utils

import (
	"github.com/amatsagu/tempest"
)

// anyInteraction is an interface that abstracts over both CommandInteraction and ModalInteraction
// (both of which can send follow up messages).
type anyInteraction interface {
	SendFollowUp(content tempest.ResponseMessageData, ephemeral bool) (tempest.Message, error)
	SendLinearFollowUp(content string, ephemeral bool) (tempest.Message, error)
	BaseUser() *tempest.User
	Responded() bool
}

// ensure both CommandInteraction and ModalInteraction satisfy the interface.
// This is removed from the compiled binary, so no efficiency is lost.
var (
	_ anyInteraction = (*tempest.CommandInteraction)(nil)
	_ anyInteraction = (*tempest.ModalInteraction)(nil)
)
