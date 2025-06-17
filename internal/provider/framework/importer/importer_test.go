// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type importerFunc func(ctx context.Context, client importer.AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse)

func importByID(ctx context.Context, f importerFunc, client importer.AWSClient, resourceSchema schema.Schema, id string, identitySchema *identityschema.Schema, identitySpec inttypes.Identity) resource.ImportStateResponse {
	var identity *tfsdk.ResourceIdentity
	if identitySchema != nil {
		identity = emtpyIdentityFromSchema(ctx, identitySchema)
	}

	request := resource.ImportStateRequest{
		ID:       id,
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(ctx, resourceSchema),
		Identity: identity,
	}
	f(ctx, client, request, &identitySpec, &response)

	return response
}

func importByIDWithState(ctx context.Context, f importerFunc, client importer.AWSClient, resourceSchema schema.Schema, id string, stateAttrs map[string]string, identitySchema *identityschema.Schema, identitySpec inttypes.Identity) resource.ImportStateResponse {
	var identity *tfsdk.ResourceIdentity
	if identitySchema != nil {
		identity = emtpyIdentityFromSchema(ctx, identitySchema)
	}

	request := resource.ImportStateRequest{
		ID:       id,
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    stateFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
	}
	f(ctx, client, request, &identitySpec, &response)

	return response
}

func importByIdentity(ctx context.Context, f importerFunc, client importer.AWSClient, resourceSchema schema.Schema, identity *tfsdk.ResourceIdentity, identitySpec inttypes.Identity) resource.ImportStateResponse {
	request := resource.ImportStateRequest{
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(ctx, resourceSchema),
		Identity: identity,
	}
	f(ctx, client, request, &identitySpec, &response)

	return response
}

func ptr[T any](v T) *T {
	return &v
}
