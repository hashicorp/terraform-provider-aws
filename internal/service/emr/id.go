package emr

import (
	"fmt"
	"strings"
)

func readStudioSessionMapping(id string) (studioId, identityType, identityId, identityName string, err error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 4 {
		return "", "", "", "", fmt.Errorf("expected ID in format studio-id:identity-type:identity-id:identity-name, received: %s", id)
	}
	return idParts[0], idParts[1], idParts[2], idParts[3], nil
}
