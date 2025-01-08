// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"bytes"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/crypto/ssh"
)

// SuppressEquivalentOpenSSHPublicKeyDiffs returns whether two OpenSSH public key format strings represent the same key.
// Any key comment is ignored when comparing values.
func SuppressEquivalentOpenSSHPublicKeyDiffs(k, old, new string, d *schema.ResourceData) bool {
	oldKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(old))
	if err != nil {
		return false
	}

	newKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(new))

	if err != nil {
		return false
	}

	return oldKey.Type() == newKey.Type() && bytes.Equal(oldKey.Marshal(), newKey.Marshal())
}
