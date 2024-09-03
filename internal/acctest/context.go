// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"testing"

	baselogging "github.com/hashicorp/aws-sdk-go-base/v2/logging"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
	helperlogging "github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

func Context(t *testing.T) context.Context {
	t.Helper()

	helperlogging.SetOutput(t)

	ctx := context.Background()
	ctx = tfsdklog.RegisterTestSink(ctx, t)
	ctx = logger(ctx, t, "acctest")
	ctx = awsSDKLogger(ctx)

	return ctx
}

func logger(ctx context.Context, t *testing.T, name string) context.Context {
	t.Helper()

	ctx = tfsdklog.NewRootProviderLogger(ctx,
		tfsdklog.WithLevelFromEnv("TF_LOG"),
		tfsdklog.WithLogName(name),
		tfsdklog.WithoutLocation(),
	)
	ctx = testNameContext(ctx, t.Name())

	return ctx
}

// testNameContext adds the current test name to loggers.
func testNameContext(ctx context.Context, testName string) context.Context {
	ctx = tflog.SetField(ctx, "test_name", testName)

	return ctx
}

func awsSDKLogger(ctx context.Context) context.Context { // nosemgrep:ci.aws-in-func-name
	ctx, logger := baselogging.NewTfLogger(ctx)
	ctx = baselogging.RegisterLogger(ctx, logger)

	return ctx
}
