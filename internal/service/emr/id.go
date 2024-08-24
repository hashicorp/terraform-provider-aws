// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
)

const identityIDPattern = `([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}`

var identityIDPatternRegexp = regexache.MustCompile(identityIDPattern)

func isIdentityID(identityIdOrName string) bool {
	return identityIDPatternRegexp.MatchString(identityIdOrName)
}

func readStudioSessionMapping(id string) (studioId, identityType, identityIdOrName string, err error) {
	idOrNameParts := strings.Split(id, ":")
	if len(idOrNameParts) == 3 {
		return idOrNameParts[0], idOrNameParts[1], idOrNameParts[2], nil
	}

	if isIdentityID(identityIdOrName) {
		err = fmt.Errorf("expected ID in format studio-id:identity-type:identity-id, received: %s", identityIdOrName)
	} else {
		err = fmt.Errorf("expected ID in format studio-id:identity-type:identity-name, received: %s", identityIdOrName)
	}

	return "", "", "", err
}
