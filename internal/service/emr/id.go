// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"fmt"
	"strings"
)

func readStudioSessionMapping(id string) (studioId, identityType, identityId string, err error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("expected ID in format studio-id:identity-type:identity-id, received: %s", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}
