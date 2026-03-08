// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package gamelift_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfgamelift "github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
)

func TestResourceContainerFleet_playerGatewaySchema(t *testing.T) {
	t.Parallel()

	r := tfgamelift.ResourceContainerFleet()

	playerGatewayMode, ok := r.Schema["player_gateway_mode"]
	if !ok {
		t.Fatal("schema missing player_gateway_mode")
	}

	if !playerGatewayMode.Optional {
		t.Error("player_gateway_mode Optional should be true")
	}
	if !playerGatewayMode.Computed {
		t.Error("player_gateway_mode Computed should be true")
	}
	if !playerGatewayMode.ForceNew {
		t.Error("player_gateway_mode ForceNew should be true")
	}

	locationAttributes, ok := r.Schema["location_attributes"]
	if !ok {
		t.Fatal("schema missing location_attributes")
	}

	locationAttributesResource, ok := locationAttributes.Elem.(*schema.Resource)
	if !ok {
		t.Fatal("location_attributes Elem should be *schema.Resource")
	}

	playerGatewayStatus, ok := locationAttributesResource.Schema["player_gateway_status"]
	if !ok {
		t.Fatal("location_attributes schema missing player_gateway_status")
	}

	if !playerGatewayStatus.Computed {
		t.Error("player_gateway_status Computed should be true")
	}
}
