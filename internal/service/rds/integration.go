// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package rds

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rds_integration", name="Integration")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @Testing(tagsTest=false)
// @Testing(preIdentityVersion="v5.100.0")
func newIntegrationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &integrationResource{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
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
	framework.ResourceWithModel[integrationResourceModel]
	framework.WithImportByIdentity
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
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
			"integration_identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"integration_name": schema.StringAttribute{
				Required: true,
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
				Update: true,
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

	var input rds.CreateIntegrationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateIntegration(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating RDS Integration (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.IntegrationARN = fwflex.StringToFramework(ctx, output.IntegrationArn)
	data.IntegrationIdentifier = integrationIDFromARN(data.IntegrationARN)
	data.ID = data.IntegrationARN // Deprecated, for backward compatibility

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

	// Import support.
	data.IntegrationARN = data.ID
	data.IntegrationIdentifier = integrationIDFromARN(data.IntegrationARN)

	conn := r.Meta().RDSClient(ctx)

	output, err := findIntegrationByARN(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
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

func (r *integrationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state integrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)
	id := plan.ID.ValueString()

	if !plan.DataFilter.Equal(state.DataFilter) ||
		!plan.IntegrationName.Equal(state.IntegrationName) {
		input := rds.ModifyIntegrationInput{
			IntegrationIdentifier: aws.String(id),
		}

		if !plan.DataFilter.Equal(state.DataFilter) {
			input.DataFilter = plan.DataFilter.ValueStringPointer()
		}

		if !plan.IntegrationName.Equal(state.IntegrationName) {
			input.IntegrationName = plan.IntegrationName.ValueStringPointer()
		}

		if _, err := conn.ModifyIntegration(ctx, &input); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating RDS Integration (%s)", id), err.Error())
			return
		}

		if _, err := waitIntegrationUpdated(ctx, conn, plan.ID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Integration (%s) update", id), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *integrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data integrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	input := rds.DeleteIntegrationInput{
		IntegrationIdentifier: fwflex.StringFromFramework(ctx, data.ID),
	}

	_, err := conn.DeleteIntegration(ctx, &input)
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
				LastError: err,
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

func statusIntegration(conn *rds.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findIntegrationByARN(ctx, conn, arn)

		if retry.NotFound(err) {
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
		Refresh: statusIntegration(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Integration); ok {
		retry.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.Errors, integrationError)...))

		return output, err
	}

	return nil, err
}

func waitIntegrationUpdated(ctx context.Context, conn *rds.Client, arn string, timeout time.Duration) (*awstypes.Integration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{integrationStatusModifying, integrationStatusSyncing},
		Target:  []string{integrationStatusActive},
		Refresh: statusIntegration(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Integration); ok {
		retry.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.Errors, integrationError)...))

		return output, err
	}

	return nil, err
}

func waitIntegrationDeleted(ctx context.Context, conn *rds.Client, arn string, timeout time.Duration) (*awstypes.Integration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{integrationStatusDeleting, integrationStatusActive},
		Target:  []string{},
		Refresh: statusIntegration(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Integration); ok {
		retry.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.Errors, integrationError)...))

		return output, err
	}

	return nil, err
}

func integrationError(v awstypes.IntegrationError) error {
	return fmt.Errorf("%s: %s", aws.ToString(v.ErrorCode), aws.ToString(v.ErrorMessage))
}

type integrationResourceModel struct {
	framework.WithRegionModel
	AdditionalEncryptionContext fwtypes.MapOfString `tfsdk:"additional_encryption_context"`
	DataFilter                  types.String        `tfsdk:"data_filter"`
	ID                          types.String        `tfsdk:"id"`
	IntegrationARN              types.String        `tfsdk:"arn"`
	IntegrationIdentifier       types.String        `tfsdk:"integration_identifier"`
	IntegrationName             types.String        `tfsdk:"integration_name"`
	KMSKeyID                    types.String        `tfsdk:"kms_key_id"`
	SourceARN                   fwtypes.ARN         `tfsdk:"source_arn"`
	Tags                        tftags.Map          `tfsdk:"tags"`
	TagsAll                     tftags.Map          `tfsdk:"tags_all"`
	TargetARN                   fwtypes.ARN         `tfsdk:"target_arn"`
	Timeouts                    timeouts.Value      `tfsdk:"timeouts"`
}

// integrationIDFromARN extracts the integration identifier from
// the integration ARN.
// arn:${Partition}:rds:${Region}:${Account}:integration:${IntegrationIdentifier}
//
// A null value is returned if:
// - The argument is null or unknown
// - The argument value string is not a valid ARN
// - The resource section of the ARN is in an unexpected format
func integrationIDFromARN(integrationARN types.String) types.String {
	if integrationARN.IsNull() || integrationARN.IsUnknown() {
		return types.StringNull()
	}

	parsed, err := arn.Parse(integrationARN.ValueString())
	if err != nil {
		return types.StringNull()
	}

	parts := strings.Split(parsed.Resource, ":")
	if len(parts) != 2 {
		return types.StringNull()
	}

	return types.StringValue(parts[1])
}
