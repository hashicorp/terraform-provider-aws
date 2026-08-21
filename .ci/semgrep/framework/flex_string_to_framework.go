// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

type output struct {
	Name   *string
	Status *string
}

func testStringValueToFrameworkWithToString1(ctx context.Context, o output) types.String {
	// ruleid: string-value-to-framework-with-aws-to-string
	return fwflex.StringValueToFramework(ctx, aws.ToString(o.Name))
}

func testStringValueToFrameworkWithToString2(ctx context.Context, o output) types.String {
	// ruleid: string-value-to-framework-with-aws-to-string
	return fwflex.StringValueToFramework(ctx, aws.ToString(o.Status))
}

func testStringValueToFrameworkWithToStringVariable(ctx context.Context, ptr *string) types.String {
	// ruleid: string-value-to-framework-with-aws-to-string
	return fwflex.StringValueToFramework(ctx, aws.ToString(ptr))
}

func testStringToFrameworkOK1(ctx context.Context, o output) types.String {
	// ok: string-value-to-framework-with-aws-to-string
	return fwflex.StringToFramework(ctx, o.Name)
}

func testStringToFrameworkOK2(ctx context.Context, ptr *string) types.String {
	// ok: string-value-to-framework-with-aws-to-string
	return fwflex.StringToFramework(ctx, ptr)
}

func testStringValueToFrameworkOK(ctx context.Context) types.String {
	// ok: string-value-to-framework-with-aws-to-string
	return fwflex.StringValueToFramework(ctx, "some plain string")
}
