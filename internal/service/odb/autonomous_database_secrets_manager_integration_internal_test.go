// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestAutonomousDatabaseSecretsManagerIntegrationSchema(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	rawResource, err := newResourceAutonomousDatabaseSecretsManagerIntegration(ctx)
	if err != nil {
		t.Fatalf("creating resource: %s", err)
	}

	response := resource.SchemaResponse{}
	rawResource.Schema(ctx, resource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("creating resource schema: %v", response.Diagnostics)
	}
	if diagnostics := response.Schema.ValidateImplementation(ctx); diagnostics.HasError() {
		t.Fatalf("validating resource schema: %v", diagnostics)
	}
}

func TestFlattenAutonomousDatabaseSecretsManagerIntegration(t *testing.T) {
	t.Parallel()

	model := autonomousDatabaseSecretsManagerIntegrationResourceModel{}
	role := &odbtypes.OciIamRole{
		IamRoleArn:   aws.String("arn:aws:iam::123456789012:role/ADBSSecretManagerServiceRole-123456789012"),
		Status:       odbtypes.OciIamRoleStatusAvailable,
		StatusReason: aws.String("available"),
	}

	flattenAutonomousDatabaseSecretsManagerIntegration(role, &model)

	if got, want := model.RoleARN.ValueString(), aws.ToString(role.IamRoleArn); got != want {
		t.Fatalf("role_arn = %q, want %q", got, want)
	}
	if got, want := model.Status.ValueString(), string(odbtypes.OciIamRoleStatusAvailable); got != want {
		t.Fatalf("status = %q, want %q", got, want)
	}
	if got, want := model.StatusReason.ValueString(), aws.ToString(role.StatusReason); got != want {
		t.Fatalf("status_reason = %q, want %q", got, want)
	}
}
