// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Snapshot Copy Configuration")
func newResourceSnapshotCopyConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSnapshotCopyConfiguration{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameSnapshotCopyConfiguration = "Snapshot Copy Configuration"
)

type resourceSnapshotCopyConfiguration struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceSnapshotCopyConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_redshiftserverless_snapshot_copy_configuration"
}

func (r *resourceSnapshotCopyConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"destination_kms_key_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"destination_region": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": framework.IDAttribute(),
			"namespace_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"snapshot_retention_period": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
			},
		},
	}
}

func (r *resourceSnapshotCopyConfiguration) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data snapshotCopyConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftServerlessClient(ctx)

	input := &redshiftserverless.CreateSnapshotCopyConfigurationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateSnapshotCopyConfiguration(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Redshift Serverless Snapshot Copy Configuration", err.Error())
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.SnapshotCopyConfiguration.SnapshotCopyConfigurationId)
	data.ARN = fwflex.StringToFramework(ctx, output.SnapshotCopyConfiguration.SnapshotCopyConfigurationArn)
	data.DestinationKmsKeyId = fwflex.StringToFramework(ctx, output.SnapshotCopyConfiguration.DestinationKmsKeyId)
	data.SnapshotRetentionPeriod = fwflex.Int32ToFramework(ctx, output.SnapshotCopyConfiguration.SnapshotRetentionPeriod)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *resourceSnapshotCopyConfiguration) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data snapshotCopyConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftServerlessClient(ctx)

	output, err := findSnapshotCopyConfigurationByID(ctx, conn, data.NamespaceName.ValueString(), data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Redshift Serverless Snapshot Copy Configuration (%s)", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceSnapshotCopyConfiguration) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old snapshotCopyConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftServerlessClient(ctx)

	input := &redshiftserverless.UpdateSnapshotCopyConfigurationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
	if response.Diagnostics.HasError() {
		return
	}
	input.SnapshotCopyConfigurationId = aws.String(new.ID.ValueString())

	_, err := conn.UpdateSnapshotCopyConfiguration(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Redshift Serverless Snapshot Copy Configuration (%s)", new.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceSnapshotCopyConfiguration) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data snapshotCopyConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftServerlessClient(ctx)

	_, err := conn.DeleteSnapshotCopyConfiguration(ctx, &redshiftserverless.DeleteSnapshotCopyConfigurationInput{
		SnapshotCopyConfigurationId: aws.String(data.ID.ValueString()),
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Redshift Serverless Snapshot Copy Configuration (%s)", data.ID.ValueString()), err.Error())
		return
	}
}

func findSnapshotCopyConfigurationByID(ctx context.Context, conn *redshiftserverless.Client, namespaceName string, ID string) (*awstypes.SnapshotCopyConfiguration, error) {
	input := &redshiftserverless.ListSnapshotCopyConfigurationsInput{
		NamespaceName: aws.String(namespaceName),
		MaxResults:    aws.Int32(100),
	}

	var configuration *awstypes.SnapshotCopyConfiguration

	for {
		output, err := conn.ListSnapshotCopyConfigurations(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, item := range output.SnapshotCopyConfigurations {
			if aws.ToString(item.SnapshotCopyConfigurationId) == ID {
				cpy := item
				configuration = &cpy
				break
			}
		}

		if (configuration != nil) || output.NextToken == nil {
			break
		}

		input.NextToken = output.NextToken
	}

	if configuration == nil {
		return nil, &retry.NotFoundError{
			LastError:   nil,
			LastRequest: input,
		}
	}

	return configuration, nil
}

type snapshotCopyConfigurationResourceModel struct {
	ARN                     types.String `tfsdk:"arn"`
	DestinationRegion       types.String `tfsdk:"destination_region"`
	DestinationKmsKeyId     types.String `tfsdk:"destination_kms_key_id"`
	ID                      types.String `tfsdk:"id"`
	NamespaceName           types.String `tfsdk:"namespace_name"`
	SnapshotRetentionPeriod types.Int64  `tfsdk:"snapshot_retention_period"`
}
