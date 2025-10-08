// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

// TimeFromFramework converts a Framework RFC3339 value to a time pointer.
// A null or unknonwn RFC3339 is converted to a nil time pointer.
func TimeFromFramework(ctx context.Context, v timetypes.RFC3339) *time.Time {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	// Because both sides are structs we don't use AutoFlEx.
	return aws.Time(fwdiag.Must(v.ValueRFC3339Time()))
}

// TimeFromFramework converts a Framework RFC3339 value to a time pointer.
func TimeToFramework(ctx context.Context, v *time.Time) timetypes.RFC3339 {
	return timetypes.NewRFC3339TimePointerValue(v)
}
