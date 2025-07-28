// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func RegisterLogger(ctx context.Context) context.Context {
	// tflog.WithLevelFromEnv() does not accommodate a custom default level,
	// so manage it ourselves here
	level := defaultLogLevel
	if l := hclog.LevelFromString(os.Getenv(envvar)); l != hclog.NoLevel {
		level = l
	}

	return registerLogger(ctx, level)
}

func registerTestingLogger(ctx context.Context) context.Context {
	return registerLogger(ctx, hclog.NoLevel)
}

func registerLogger(ctx context.Context, logLevel hclog.Level) context.Context {
	return tflog.NewSubsystem(ctx, subsystemName,
		tflog.WithLevel(logLevel),
		tflog.WithRootFields(),
	)
}
