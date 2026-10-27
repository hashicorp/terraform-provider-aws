// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/continuum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/continuum/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_continuum_collector_association", name="Collector Association")
// @IdentityAttribute("environment_id")
// @IdentityAttribute("collector_arn")
// @ImportIDHandler("collectorAssociationImportID", setIDAttribute=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(identityTest=false)
// @Testing(tagsTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/continuum/types;types.AssociatedCollector")
func newCollectorAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &collectorAssociationResource{}, nil
}

type collectorAssociationResource struct {
	framework.ResourceWithModel[collectorAssociationResourceModel]
	framework.WithImportByIdentity
}

func (r *collectorAssociationResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"collector_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"collector_data_access_role_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the IAM role that Continuum assumes to read data from the collector.",
			},
			"collector_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttributeDeprecatedNoReplacement(),
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *collectorAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan collectorAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	output, err := associateCollector(ctx, conn, &plan)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.EnvironmentID.String())
		return
	}
	if output.AssociatedCollector != nil {
		plan.CollectorID = types.StringPointerValue(output.AssociatedCollector.CollectorId)
	}
	plan.ID = types.StringValue(plan.id())

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *collectorAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state collectorAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	environmentID, collectorARN := state.EnvironmentID.ValueString(), state.CollectorARN.ValueString()
	associated, err := findAssociatedCollector(ctx, conn, environmentID, collectorARN)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, environmentID)
		return
	}
	state.CollectorID = types.StringPointerValue(associated.CollectorId)
	state.ID = types.StringValue(state.id())

	environment, err := findEnvironmentByID(ctx, conn, environmentID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, environmentID)
		return
	}
	if code := environment.Configuration.CodeConfiguration; code != nil && code.CollectorDetails != nil && code.CollectorDetails.CollectorDataAccessRoleArn != nil {
		state.CollectorDataAccessRoleARN = fwtypes.ARNValue(aws.ToString(code.CollectorDetails.CollectorDataAccessRoleArn))
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *collectorAssociationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan collectorAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	// AssociateCollector is idempotent and re-authorizes with the supplied data access role.
	output, err := associateCollector(ctx, conn, &plan)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.EnvironmentID.String())
		return
	}
	if output.AssociatedCollector != nil {
		plan.CollectorID = types.StringPointerValue(output.AssociatedCollector.CollectorId)
	}
	plan.ID = types.StringValue(plan.id())

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *collectorAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state collectorAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	input := continuum.DisassociateCollectorInput{
		CollectorArn:  state.CollectorARN.ValueStringPointer(),
		EnvironmentId: state.EnvironmentID.ValueStringPointer(),
	}
	_, err := conn.DisassociateCollector(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.EnvironmentID.String())
		return
	}
}

func associateCollector(ctx context.Context, conn *continuum.Client, data *collectorAssociationResourceModel) (*continuum.AssociateCollectorOutput, error) {
	input := continuum.AssociateCollectorInput{
		ClientToken:                aws.String(create.UniqueId(ctx)),
		CollectorArn:               data.CollectorARN.ValueStringPointer(),
		CollectorDataAccessRoleArn: data.CollectorDataAccessRoleARN.ValueStringPointer(),
		EnvironmentId:              data.EnvironmentID.ValueStringPointer(),
	}

	output, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
		func(ctx context.Context) (*continuum.AssociateCollectorOutput, error) {
			input.ClientToken = aws.String(create.UniqueId(ctx))
			return conn.AssociateCollector(ctx, &input)
		},
		isRoleNotAssumable,
	)
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, errors.New("empty output")
	}

	return output, nil
}

func findAssociatedCollector(ctx context.Context, conn *continuum.Client, environmentID, collectorARN string) (*awstypes.AssociatedCollector, error) {
	input := continuum.ListAssociatedCollectorsInput{
		EnvironmentId: aws.String(environmentID),
	}

	pages := continuum.NewListAssociatedCollectorsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		if err != nil {
			return nil, err
		}

		for _, v := range page.AssociatedCollectors {
			if aws.ToString(v.CollectorArn) == collectorARN {
				return &v, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError()
}

type collectorAssociationResourceModel struct {
	framework.WithRegionModel
	CollectorARN               fwtypes.ARN  `tfsdk:"collector_arn"`
	CollectorDataAccessRoleARN fwtypes.ARN  `tfsdk:"collector_data_access_role_arn"`
	CollectorID                types.String `tfsdk:"collector_id"`
	EnvironmentID              types.String `tfsdk:"environment_id"`
	ID                         types.String `tfsdk:"id"`
}

func (m *collectorAssociationResourceModel) id() string {
	return m.EnvironmentID.ValueString() + intflex.ResourceIdSeparator + m.CollectorARN.ValueString()
}

var (
	_ inttypes.ImportIDParser           = collectorAssociationImportID{}
	_ inttypes.FrameworkImportIDCreator = collectorAssociationImportID{}
)

type collectorAssociationImportID struct{}

func (collectorAssociationImportID) Parse(id string) (string, map[string]any, error) {
	environmentID, collectorARN, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <environment-id>%s<collector-arn>", id, intflex.ResourceIdSeparator)
	}

	result := map[string]any{
		"environment_id": environmentID,
		"collector_arn":  collectorARN,
	}

	return id, result, nil
}

func (collectorAssociationImportID) Create(ctx context.Context, state tfsdk.State) string {
	var environmentID, collectorARN types.String
	state.GetAttribute(ctx, path.Root("environment_id"), &environmentID)
	state.GetAttribute(ctx, path.Root("collector_arn"), &collectorARN)

	return environmentID.ValueString() + intflex.ResourceIdSeparator + collectorARN.ValueString()
}
