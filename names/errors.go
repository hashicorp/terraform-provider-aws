package names

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

const (
	ErrActionReading           = "reading"
	ErrActionDeleting          = "deleting"
	ErrActionUpdating          = "updating"
	ErrActionCreating          = "creating"
	ErrActionCheckingExistence = "checking existence"
	ErrActionCheckingDestroyed = "checking destroyed"
)

func Error(service, action, resource, id string, gotError error) error {
	hf, err := FullHumanFriendly(service)

	if err != nil {
		return fmt.Errorf("finding human-friendly name for service (%s) while creating error (%s, %s, %s, %s): %w", service, action, resource, id, gotError, err)
	}

	if gotError == nil {
		return fmt.Errorf("%s %s %s (%s)", action, hf, resource, id)
	}

	return fmt.Errorf("%s %s %s (%s): %w", action, hf, resource, id, gotError)
}

func DiagError(service, action, resource, id string, gotError error) diag.Diagnostics {
	hf, err := FullHumanFriendly(service)

	if err != nil {
		return diag.Errorf("finding human-friendly name for service (%s) while creating error (%s, %s, %s, %s): %s", service, action, resource, id, gotError, err)
	}

	if gotError == nil {
		return diag.Errorf("%s %s %s (%s)", action, hf, resource, id)
	}

	return diag.Errorf("%s %s %s (%s): %s", action, hf, resource, id, gotError)
}

func WarnLog(service, action, resource, id string, gotError error) {
	hf, err := FullHumanFriendly(service)

	if err != nil {
		log.Printf("[ERROR] finding human-friendly name for service (%s) while logging warn (%s, %s, %s, %s): %s", service, action, resource, id, gotError, err)
	}

	if gotError == nil {
		log.Printf("[WARN] %s %s %s (%s)", action, hf, resource, id)
	}

	log.Printf("[WARN] %s %s %s (%s): %s", action, hf, resource, id, gotError)
}

func LogNotFoundRemoveState(service, action, resource, id string) {
	WarnLog(service, action, resource, id, errors.New("not found, removing from state"))
}
