// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_msk_single_scram_secret_association", name="Single SCRAM Secret Association")
func newSingleSCRAMSecretAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &singleSCRAMSecretAssociationResource{}

	return r, nil
}

type singleSCRAMSecretAssociationResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *singleSCRAMSecretAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"secret_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *singleSCRAMSecretAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data singleSCRAMSecretAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	parts := []string{
		data.ClusterARN.ValueString(),
		data.SecretARN.ValueString(),
	}
	id, err := flex.FlattenResourceId(parts, singleSCRAMSecretAssociationResourceIDPartCount, false)

	if err != nil {
		response.Diagnostics.AddError("creating MSK Single SCRAM Secret Association resource ID", err.Error())

		return
	}

	conn := r.Meta().KafkaClient(ctx)

	if err := associateSRAMSecrets(ctx, conn, data.ClusterARN.ValueString(), []string{data.SecretARN.ValueString()}); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating MSK Single SCRAM Secret Association (%s)", id), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, id)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *singleSCRAMSecretAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data singleSCRAMSecretAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().KafkaClient(ctx)

	err := findSingleSCRAMSecretAssociationByTwoPartKey(ctx, conn, data.ClusterARN.ValueString(), data.SecretARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading MSK Single SCRAM Secret Association (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *singleSCRAMSecretAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data singleSCRAMSecretAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().KafkaClient(ctx)

	err := disassociateSRAMSecrets(ctx, conn, data.ClusterARN.ValueString(), []string{data.SecretARN.ValueString()})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting MSK Single SCRAM Secret Association (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

type singleSCRAMSecretAssociationResourceModel struct {
	ClusterARN fwtypes.ARN  `tfsdk:"cluster_arn"`
	ID         types.String `tfsdk:"id"`
	SecretARN  fwtypes.ARN  `tfsdk:"secret_arn"`
}

const (
	singleSCRAMSecretAssociationResourceIDPartCount = 2
)

func (data *singleSCRAMSecretAssociationResourceModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, singleSCRAMSecretAssociationResourceIDPartCount, false)

	if err != nil {
		return err
	}

	data.ClusterARN = fwtypes.ARNValue(parts[0])
	data.SecretARN = fwtypes.ARNValue(parts[1])

	return nil
}

func findSingleSCRAMSecretAssociationByTwoPartKey(ctx context.Context, conn *kafka.Client, clusterARN, secretARN string) error {
	output, err := findSCRAMSecretsByClusterARN(ctx, conn, clusterARN)

	if err != nil {
		return err
	}

	if !slices.Contains(output, secretARN) {
		return tfresource.NewEmptyResultError(nil)
	}

	return nil
}
