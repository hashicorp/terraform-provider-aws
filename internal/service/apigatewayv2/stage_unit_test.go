// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"encoding/json"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/aws/smithy-go/ptr"
	"testing"
)

type jsonSerializableRouteSettings struct {
	DataTraceEnabled       bool     `json:"data_trace_enabled"`
	DetailedMetricsEnabled bool     `json:"detailed_metrics_enabled"`
	LoggingLevel           string   `json:"logging_level"`
	ThrottlingBurstLimit   *int32   `json:"throttling_burst_limit"`
	ThrottlingRateLimit    *float64 `json:"throttling_rate_limit"`
}

func assertRouteSettingsEqualTo(t *testing.T, expected jsonSerializableRouteSettings, actual any) {
	serialized, err := json.Marshal(actual)

	if err != nil {
		t.Fatal("unexpected JSON serialization error")

		return
	}

	var deSerialized jsonSerializableRouteSettings

	err = json.Unmarshal(serialized, &deSerialized)

	if err != nil {
		t.Fatal("unexpected JSON deserialization error")

		return
	}

	if !(expected.DataTraceEnabled == deSerialized.DataTraceEnabled &&
		expected.DetailedMetricsEnabled == deSerialized.DetailedMetricsEnabled &&
		expected.LoggingLevel == deSerialized.LoggingLevel &&
		((expected.ThrottlingBurstLimit == nil &&
			deSerialized.ThrottlingBurstLimit == nil) ||
			*(expected.ThrottlingBurstLimit) == *(deSerialized.ThrottlingBurstLimit)) &&
		((expected.ThrottlingRateLimit == nil &&
			deSerialized.ThrottlingRateLimit == nil) ||
			*(expected.ThrottlingRateLimit) == *(deSerialized.ThrottlingRateLimit))) {
		t.Fatal("expected/actual route settings differ")
	}
}

func TestAPIGatewayV2Stage_flattenDefaultRouteSettings(t *testing.T) {
	t.Parallel()

	t.Run("default settings given, when nothing is set", func(t *testing.T) {
		var routeSettings awstypes.RouteSettings

		flattened := flattenRouteSettings(map[string]awstypes.RouteSettings{
			"example": routeSettings,
		})

		if len(flattened) != 1 {
			t.Fatal("expected route settings to contain one item")
		}

		assertRouteSettingsEqualTo(
			t,
			jsonSerializableRouteSettings{
				DataTraceEnabled:       false,
				DetailedMetricsEnabled: false,
				LoggingLevel:           "",
				ThrottlingBurstLimit:   nil,
				ThrottlingRateLimit:    nil,
			},
			flattened[0],
		)
	})

	t.Run("integer throttling is forwarded, when given", func(t *testing.T) {
		flattened := flattenRouteSettings(map[string]awstypes.RouteSettings{
			"example": awstypes.RouteSettings{
				ThrottlingRateLimit:  ptr.Float64(123),
				ThrottlingBurstLimit: ptr.Int32(456),
			},
		})

		if len(flattened) != 1 {
			t.Fatal("expected route settings to contain one item")
		}

		assertRouteSettingsEqualTo(
			t,
			jsonSerializableRouteSettings{
				DataTraceEnabled:       false,
				DetailedMetricsEnabled: false,
				LoggingLevel:           "",
				ThrottlingBurstLimit:   ptr.Int32(456),
				ThrottlingRateLimit:    ptr.Float64(123),
			},
			flattened[0],
		)
	})
}
