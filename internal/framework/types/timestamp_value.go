// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewTimestampNull() TimestampValue {
	return TimestampValue{
		StringValue: types.StringNull(),
	}
}

func NewTimestampUnknown() TimestampValue {
	return TimestampValue{
		StringValue: types.StringUnknown(),
	}
}

func NewTimestampValue(t time.Time) TimestampValue {
	return newTimestampValue(t.Format(time.RFC3339), t)
}

func NewTimestampValueString(s string) (TimestampValue, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return TimestampValue{}, err
	}
	return newTimestampValue(s, t), nil
}

func newTimestampValue(s string, t time.Time) TimestampValue {
	return TimestampValue{
		StringValue: types.StringValue(s),
		value:       t,
	}
}

var (
	_ basetypes.StringValuable = TimestampValue{}
)

type TimestampValue struct {
	basetypes.StringValue

	// value contains the parsed value, if not Null or Unknown.
	value time.Time
}

func (val TimestampValue) Type(_ context.Context) attr.Type {
	return TimestampType{}
}

func (val TimestampValue) Equal(other attr.Value) bool {
	o, ok := other.(TimestampValue)

	if !ok {
		return false
	}

	if val.StringValue.IsUnknown() {
		return o.StringValue.IsUnknown()
	}

	if val.StringValue.IsNull() {
		return o.StringValue.IsNull()
	}

	return val.value.Equal(o.value)
}

// ValueTimestamp returns the known time.Time value. If Timestamp is null or unknown, returns 0.
// To get the value as a string, use ValueString.
func (val TimestampValue) ValueTimestamp() time.Time {
	return val.value
}
