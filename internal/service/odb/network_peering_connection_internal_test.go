// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNetworkPeeringConnectionOdbNetworkID(t *testing.T) {
	t.Parallel()

	const networkARN = "arn:aws:odb:us-east-1:123456789012:odb-network/odbnet_abcdefgh12"
	const networkID = "odbnet_abcdefgh12"

	tests := map[string]struct {
		current types.String
		want    types.String
	}{
		"configured ARN": {types.StringValue(networkARN), types.StringValue(networkARN)},
		"configured ID":  {types.StringValue(networkID), types.StringValue(networkID)},
		"imported":       {types.StringNull(), types.StringValue(networkID)},
		"unknown":        {types.StringUnknown(), types.StringValue(networkID)},
		"different ARN":  {types.StringValue("arn:aws:odb:us-east-1:123456789012:odb-network/odbnet-other"), types.StringValue(networkID)},
		"different ID":   {types.StringValue("odbnet-other"), types.StringValue(networkID)},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := networkPeeringConnectionOdbNetworkID(test.current, networkARN, networkID)
			if !got.Equal(test.want) {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestNetworkPeeringConnectionSameOdbNetwork(t *testing.T) {
	t.Parallel()

	const networkARN = "arn:aws:odb:us-east-1:123456789012:odb-network/odbnet_abcdefgh12"
	const networkID = "odbnet_abcdefgh12"

	tests := map[string]struct {
		current    types.String
		planned    types.String
		networkARN types.String
		want       bool
	}{
		"ID to ARN":               {types.StringValue(networkID), types.StringValue(networkARN), types.StringValue(networkARN), true},
		"ARN to ID":               {types.StringValue(networkARN), types.StringValue(networkID), types.StringValue(networkARN), true},
		"same network ID":         {types.StringValue(networkID), types.StringValue(networkID), types.StringValue(networkARN), true},
		"different network ID":    {types.StringValue(networkID), types.StringValue("odbnet_different"), types.StringValue(networkARN), false},
		"different network ARN":   {types.StringValue(networkID), types.StringValue("arn:aws:odb:us-east-1:123456789012:odb-network/odbnet_different"), types.StringValue(networkARN), false},
		"different account":       {types.StringValue(networkID), types.StringValue("arn:aws:odb:us-east-1:999999999999:odb-network/odbnet_abcdefgh12"), types.StringValue(networkARN), false},
		"different region":        {types.StringValue(networkID), types.StringValue("arn:aws:odb:us-west-2:123456789012:odb-network/odbnet_abcdefgh12"), types.StringValue(networkARN), false},
		"missing current":         {types.StringNull(), types.StringValue(networkARN), types.StringValue(networkARN), false},
		"unknown planned":         {types.StringValue(networkID), types.StringUnknown(), types.StringValue(networkARN), false},
		"missing network ARN":     {types.StringValue(networkID), types.StringValue(networkARN), types.StringNull(), false},
		"malformed network ARN":   {types.StringValue(networkID), types.StringValue(networkARN), types.StringValue("not-an-arn"), false},
		"different resource kind": {types.StringValue(networkID), types.StringValue(networkARN), types.StringValue("arn:aws:odb:us-east-1:123456789012:odb-peering/odbnet_abcdefgh12"), false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := networkPeeringConnectionSameOdbNetwork(test.current, test.planned, test.networkARN)
			if got != test.want {
				t.Errorf("got %t, want %t", got, test.want)
			}
		})
	}
}
