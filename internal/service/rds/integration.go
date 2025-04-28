// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rds_integration", name="Integration")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newIntegrationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &integrationResource{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	integrationStatusActive         = "active"
	integrationStatusCreating       = "creating"
	integrationStatusDeleting       = "deleting"
	integrationStatusFailed         = "failed"
	integrationStatusModifying      = "modifying"
	integrationStatusNeedsAttention = "needs_attention"
	integrationStatusSyncing        = "syncing"
)

type integrationResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[integrationResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *integrationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"additional_encryption_context": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"data_filter": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"integration_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrTargetARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *integrationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data integrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	name := data.IntegrationName.ValueString()
	input := &rds.CreateIntegrationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateIntegration(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating RDS Integration (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.IntegrationARN = fwflex.StringToFramework(ctx, output.IntegrationArn)
	data.setID()

	integration, err := waitIntegrationCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Integration (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.KMSKeyID = fwflex.StringToFramework(ctx, integration.KMSKeyId)
	data.DataFilter = fwflex.StringToFramework(ctx, integration.DataFilter)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *integrationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data integrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().RDSClient(ctx)

	output, err := findIntegrationByARN(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading RDS Integration (%s)", data.ID.ValueString()), err.Error())

		return
	}

	prevAdditionalEncryptionContext := data.AdditionalEncryptionContext

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Null vs. empty map handling.
	if prevAdditionalEncryptionContext.IsNull() && !data.AdditionalEncryptionContext.IsNull() && len(data.AdditionalEncryptionContext.Elements()) == 0 {
		data.AdditionalEncryptionContext = prevAdditionalEncryptionContext
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *integrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data integrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	_, err := conn.DeleteIntegration(ctx, &rds.DeleteIntegrationInput{
		IntegrationIdentifier: fwflex.StringFromFramework(ctx, data.ID),
	})

	if errs.IsA[*awstypes.IntegrationNotFoundFault](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting RDS Integration (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitIntegrationDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Integration (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func findIntegrationByARN(ctx context.Context, conn *rds.Client, arn string) (*awstypes.Integration, error) {
	input := &rds.DescribeIntegrationsInput{
		IntegrationIdentifier: aws.String(arn),
	}

	return findIntegration(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Integration]())
}

func findIntegration(ctx context.Context, conn *rds.Client, input *rds.DescribeIntegrationsInput, filter tfslices.Predicate[*awstypes.Integration]) (*awstypes.Integration, error) {
	output, err := findIntegrations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIntegrations(ctx context.Context, conn *rds.Client, input *rds.DescribeIntegrationsInput, filter tfslices.Predicate[*awstypes.Integration]) ([]awstypes.Integration, error) {
	var output []awstypes.Integration

	pages := rds.NewDescribeIntegrationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.IntegrationNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Integrations {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusIntegration(ctx context.Context, conn *rds.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findIntegrationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitIntegrationCreated(ctx context.Context, conn *rds.Client, arn string, timeout time.Duration) (*awstypes.Integration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{integrationStatusCreating, integrationStatusModifying},
		Target:  []string{integrationStatusActive},
		Refresh: statusIntegration(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Integration); ok {
		tfresource.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.Errors, integrationError)...))

		return output, err
	}

	return nil, err
}

func waitIntegrationDeleted(ctx context.Context, conn *rds.Client, arn string, timeout time.Duration) (*awstypes.Integration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{integrationStatusDeleting, integrationStatusActive},
		Target:  []string{},
		Refresh: statusIntegration(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Integration); ok {
		tfresource.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.Errors, integrationError)...))

		return output, err
	}

	return nil, err
}

func integrationError(v awstypes.IntegrationError) error {
	return fmt.Errorf("%s: %s", aws.ToString(v.ErrorCode), aws.ToString(v.ErrorMessage))
}

type integrationResourceModel struct {
	AdditionalEncryptionContext fwtypes.MapValueOf[types.String] `tfsdk:"additional_encryption_context"`
	DataFilter                  types.String                     `tfsdk:"data_filter"`
	ID                          types.String                     `tfsdk:"id"`
	IntegrationARN              types.String                     `tfsdk:"arn"`
	IntegrationName             types.String                     `tfsdk:"integration_name"`
	KMSKeyID                    types.String                     `tfsdk:"kms_key_id"`
	SourceARN                   fwtypes.ARN                      `tfsdk:"source_arn"`
	Tags                        tftags.Map                       `tfsdk:"tags"`
	TagsAll                     tftags.Map                       `tfsdk:"tags_all"`
	TargetARN                   fwtypes.ARN                      `tfsdk:"target_arn"`
	Timeouts                    timeouts.Value                   `tfsdk:"timeouts"`
}

func (model *integrationResourceModel) InitFromID() error {
	model.IntegrationARN = model.ID

	return nil
}

func (model *integrationResourceModel) setID() {
	model.ID = model.IntegrationARN
}
