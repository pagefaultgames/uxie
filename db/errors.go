package db

import (
	"fmt"
)

// errNotFound is returned when a command cannot be located.
func errNotFound(name string) error {
	return fmt.Errorf("command %s not found", name)
}

func errImmutable(name string) error {
	return fmt.Errorf("command %s is immutable and cannot be modified or deleted", name)
}
