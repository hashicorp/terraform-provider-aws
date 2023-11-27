// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func TimestampNull() Timestamp {
	return Timestamp{StringValue: types.StringNull()}
}

func TimestampUnknown() Timestamp {
	return Timestamp{StringValue: types.StringUnknown()}
}

func TimestampValue(value string) Timestamp {
	return Timestamp{
		StringValue: basetypes.NewStringValue(value),
		value:       errs.Must(time.Parse(time.RFC3339, value)),
	}
}

var (
	_ basetypes.StringValuable = (*Timestamp)(nil)
)

type Timestamp struct {
	basetypes.StringValue
	value time.Time
}

func (v Timestamp) Equal(o attr.Value) bool {
	other, ok := o.(Timestamp)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v Timestamp) Type(_ context.Context) attr.Type {
	return TimestampType
}

func (v Timestamp) ValueTimestamp() time.Time {
	return v.value
}
