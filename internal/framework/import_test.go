// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework_test

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type importStater interface {
	ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse)
}

func importByID(ctx context.Context, importer importStater, resourceSchema schema.Schema, id string, identitySchema identityschema.Schema) resource.ImportStateResponse {
	identity := emtpyIdentityFromSchema(ctx, identitySchema)

	request := resource.ImportStateRequest{
		ID:       id,
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(ctx, resourceSchema),
		Identity: identity,
	}
	importer.ImportState(ctx, request, &response)

	return response
}

func importByIDWithState(ctx context.Context, importer importStater, resourceSchema schema.Schema, id string, stateAttrs map[string]string, identitySchema identityschema.Schema) resource.ImportStateResponse {
	identity := emtpyIdentityFromSchema(ctx, identitySchema)

	request := resource.ImportStateRequest{
		ID:       id,
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    stateFromSchema(ctx, resourceSchema, stateAttrs),
		Identity: identity,
	}
	importer.ImportState(ctx, request, &response)

	return response
}

func importByIDNoIdentity(ctx context.Context, importer importStater, resourceSchema schema.Schema, id string) resource.ImportStateResponse {
	request := resource.ImportStateRequest{
		ID:       id,
		Identity: nil,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(ctx, resourceSchema),
		Identity: nil,
	}
	importer.ImportState(ctx, request, &response)

	return response
}

func importByIdentity(ctx context.Context, importer importStater, resourceSchema schema.Schema, identitySchema identityschema.Schema, identityAttrs map[string]string) resource.ImportStateResponse {
	identity := identityFromSchema(ctx, identitySchema, identityAttrs)

	request := resource.ImportStateRequest{
		Identity: identity,
	}
	response := resource.ImportStateResponse{
		State:    emtpyStateFromSchema(ctx, resourceSchema),
		Identity: identity,
	}
	importer.ImportState(ctx, request, &response)

	return response
}
