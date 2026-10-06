// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func testNetworkPeeringConnectionARN(region, accountID, resource string) string {
	return arn.ARN{
		Partition: endpoints.AwsPartitionID,
		Service:   "odb",
		Region:    region,
		AccountID: accountID,
		Resource:  resource,
	}.String()
}

func TestNetworkPeeringConnectionNetworkID(t *testing.T) {
	t.Parallel()

	const networkID = "odbnet_abcdefgh12"
	networkARN := testNetworkPeeringConnectionARN(endpoints.UsEast1RegionID, "123456789012", "odb-network/"+networkID)

	tests := map[string]struct {
		current types.String
		want    types.String
	}{
		"configured ARN": {types.StringValue(networkARN), types.StringValue(networkARN)},
		"configured ID":  {types.StringValue(networkID), types.StringValue(networkID)},
		"imported":       {types.StringNull(), types.StringValue(networkID)},
		"unknown":        {types.StringUnknown(), types.StringValue(networkID)},
		"different ARN":  {types.StringValue(testNetworkPeeringConnectionARN(endpoints.UsEast1RegionID, "123456789012", "odb-network/odbnet-other")), types.StringValue(networkID)},
		"different ID":   {types.StringValue("odbnet-other"), types.StringValue(networkID)},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := networkPeeringConnectionNetworkID(test.current, networkARN, networkID)
			if !got.Equal(test.want) {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestNetworkPeeringConnectionSameNetwork(t *testing.T) {
	t.Parallel()

	const networkID = "odbnet_abcdefgh12"
	networkARN := testNetworkPeeringConnectionARN(endpoints.UsEast1RegionID, "123456789012", "odb-network/"+networkID)

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
		"different network ARN":   {types.StringValue(networkID), types.StringValue(testNetworkPeeringConnectionARN(endpoints.UsEast1RegionID, "123456789012", "odb-network/odbnet_different")), types.StringValue(networkARN), false},
		"different account":       {types.StringValue(networkID), types.StringValue(testNetworkPeeringConnectionARN(endpoints.UsEast1RegionID, "999999999999", "odb-network/"+networkID)), types.StringValue(networkARN), false},
		"different region":        {types.StringValue(networkID), types.StringValue(testNetworkPeeringConnectionARN(endpoints.UsWest2RegionID, "123456789012", "odb-network/"+networkID)), types.StringValue(networkARN), false},
		"missing current":         {types.StringNull(), types.StringValue(networkARN), types.StringValue(networkARN), false},
		"unknown planned":         {types.StringValue(networkID), types.StringUnknown(), types.StringValue(networkARN), false},
		"missing network ARN":     {types.StringValue(networkID), types.StringValue(networkARN), types.StringNull(), false},
		"malformed network ARN":   {types.StringValue(networkID), types.StringValue(networkARN), types.StringValue("not-an-arn"), false},
		"different resource kind": {types.StringValue(networkID), types.StringValue(networkARN), types.StringValue(testNetworkPeeringConnectionARN(endpoints.UsEast1RegionID, "123456789012", "odb-peering/"+networkID)), false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := networkPeeringConnectionSameNetwork(test.current, test.planned, test.networkARN)
			if got != test.want {
				t.Errorf("got %t, want %t", got, test.want)
			}
		})
	}
}
