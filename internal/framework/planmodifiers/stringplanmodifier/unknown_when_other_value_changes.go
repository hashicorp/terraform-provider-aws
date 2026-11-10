// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringplanmodifier

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/internal"
)

func UnknownWhenOtherValueChanges(path path.Path) planmodifier.String {
	return internal.UnknownWhenOtherValueChanges(path)
}
