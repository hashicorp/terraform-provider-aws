// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func RegisterLogger(ctx context.Context) context.Context {
	return tflog.NewSubsystem(ctx, subsystemName,
		tflog.WithLevelFromEnv(envvar),
		tflog.WithRootFields(),
	)
}
