// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

import (
	"context"

	baselogging "github.com/hashicorp/aws-sdk-go-base/v2/logging"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	HTTPKeyRequestBody  = "http.request.body"
	HTTPKeyResponseBody = "http.response.body"
	KeyResourceId       = "tf_aws.resource_attribute." + "id"
)

func ResourceAttributeKey(name string) string {
	return "tf_aws.resource_attribute." + name
}

// MaskSensitiveValuesByKey masks sensitive values using tflog
func MaskSensitiveValuesByKey(ctx context.Context, keys ...string) context.Context {
	l := baselogging.RetrieveLogger(ctx)

	if _, ok := l.(baselogging.NullLogger); ok {
		return ctx
	}

	for _, v := range keys {
		ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, v)
	}

	return ctx
}
