// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Import(ctx context.Context, rd *schema.ResourceData, meta any) error {
	identitySpec := identitySpec(ctx)

	switch {
	case identitySpec.IsARN:
		if identitySpec.IsGlobalResource {
			return GlobalARN(ctx, rd, identitySpec)
		} else {
			return RegionalARN(ctx, rd, identitySpec)
		}

	case identitySpec.IsSingleton:
		if identitySpec.IsGlobalResource {
			return GlobalSingleton(ctx, rd, identitySpec, meta.(AWSClient))
		} else {
			return RegionalSingleton(ctx, rd, identitySpec, meta.(AWSClient))
		}

	case identitySpec.IsCustomInherentRegion:
		// Not supported for Global resources. This is validated in validateResourceSchemas().
		return RegionalInherentRegion(ctx, rd, identitySpec)

	case identitySpec.IsSingleParameter:
		if identitySpec.IsGlobalResource {
			return GlobalSingleParameterized(ctx, rd, identitySpec, meta.(AWSClient))
		} else {
			return RegionalSingleParameterized(ctx, rd, identitySpec, meta.(AWSClient))
		}

	default:
		importSpec := importSpec(ctx)
		if identitySpec.IsGlobalResource {
			return GlobalMultipleParameterized(ctx, rd, identitySpec, importSpec, meta.(AWSClient))
		} else {
			return RegionalMultipleParameterized(ctx, rd, identitySpec, importSpec, meta.(AWSClient))
		}
	}
}
