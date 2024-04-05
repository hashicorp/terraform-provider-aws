// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tf6serverlogging

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/internal/logging"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ServerCapabilities generates a TRACE "Announced server capabilities" log.
func ServerCapabilities(ctx context.Context, capabilities *tfprotov6.ServerCapabilities) {
	responseFields := map[string]interface{}{
		logging.KeyServerCapabilityGetProviderSchemaOptional: false,
		logging.KeyServerCapabilityPlanDestroy:               false,
	}

	if capabilities != nil {
		responseFields[logging.KeyServerCapabilityGetProviderSchemaOptional] = capabilities.GetProviderSchemaOptional
		responseFields[logging.KeyServerCapabilityPlanDestroy] = capabilities.PlanDestroy
	}

	logging.ProtocolTrace(ctx, "Announced server capabilities", responseFields)
}
