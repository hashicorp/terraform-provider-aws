// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package create

import (
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
)

// Name returns in order the name if non-empty, a prefix generated name if non-empty, or fully generated name prefixed with terraform-
func Name(name string, namePrefix string) string {
	return NameWithSuffix(name, namePrefix, "")
}

// NameWithSuffix returns in order the name if non-empty, a prefix generated name if non-empty, or fully generated name prefixed with "terraform-".
// In the latter two cases, any suffix is appended to the generated name
func NameWithSuffix(name string, namePrefix string, nameSuffix string) string {
	if name != "" {
		return name
	}

	if namePrefix != "" {
		return id.PrefixedUniqueId(namePrefix) + nameSuffix
	}

	return id.UniqueId() + nameSuffix
}

// hasResourceUniqueIDPlusAdditionalSuffix returns true if the string has the built-in unique ID suffix plus an additional suffix
func hasResourceUniqueIDPlusAdditionalSuffix(s string, additionalSuffix string) bool {
	re := regexache.MustCompile(fmt.Sprintf("[[:xdigit:]]{%d}%s$", id.UniqueIDSuffixLength, additionalSuffix))
	return re.MatchString(s)
}

// NamePrefixFromName returns a name prefix if the string matches prefix criteria
//
// The input to this function must be strictly the "name" and not any
// additional information such as a full Amazon Resource Name (ARN).
//
// An expected usage might be:
//
//	d.Set("name_prefix", create.NamePrefixFromName(d.Id()))
func NamePrefixFromName(name string) *string {
	return NamePrefixFromNameWithSuffix(name, "")
}

func NamePrefixFromNameWithSuffix(name, nameSuffix string) *string {
	if !hasResourceUniqueIDPlusAdditionalSuffix(name, nameSuffix) {
		return nil
	}

	namePrefixIndex := len(name) - id.UniqueIDSuffixLength - len(nameSuffix)

	if namePrefixIndex <= 0 {
		return nil
	}

	namePrefix := name[:namePrefixIndex]

	return &namePrefix
}
