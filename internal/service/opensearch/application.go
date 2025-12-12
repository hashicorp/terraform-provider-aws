// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_opensearch_application", name="Application")
// @Tags(identifierAttribute="arn")
func newResourceApplication(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceApplication{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameApplication = "Application"
)

type resourceApplication struct {
	framework.ResourceWithModel[resourceApplicationModel]
	framework.WithTimeouts
}

func (r *resourceApplication) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 30),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z][a-z0-9\-]+$`),
						"name must start with a lowercase letter and contain only lowercase letters, numbers, and hyphens",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"app_config": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[appConfigModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtMost(200),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKey: schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									"opensearchDashboards.dashboardAdmin.users",
									"opensearchDashboards.dashboardAdmin.groups",
								),
							},
						},
						names.AttrValue: schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 4096),
							},
						},
					},
				},
			},
			"data_source": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[dataSourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"data_source_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(20, 2048),
							},
						},
						"data_source_description": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(1000),
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^([a-zA-Z0-9_])*[\\a-zA-Z0-9_@#%*+=:?./!\s-]*$`),
									"data source description contains invalid characters",
								),
							},
						},
					},
				},
			},
			"iam_identity_center_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[iamIdentityCenterOptionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrEnabled: schema.BoolAttribute{
							Optional: true,
						},
						"iam_identity_center_instance_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(20, 2048),
							},
						},
						"iam_role_for_identity_center_application_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(20, 2048),
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^arn:(aws|aws\-cn|aws\-us\-gov|aws\-iso|aws\-iso\-b):iam::[0-9]+:role\/.*$`),
									"must be a valid IAM role ARN",
								),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().OpenSearchClient(ctx)

	var plan resourceApplicationModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input opensearch.CreateApplicationInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate ClientToken for idempotency
	input.ClientToken = aws.String(id.UniqueId())

	// Additional fields.
	input.TagList = getTagsIn(ctx)

	out, err := conn.CreateApplication(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitApplicationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchClient(ctx)

	var state resourceApplicationModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchClient(ctx)

	var plan, state resourceApplicationModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input opensearch.UpdateApplicationInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateApplication(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitApplicationUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchClient(ctx)

	var state resourceApplicationModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := opensearch.DeleteApplicationInput{
		Id: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteApplication(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitApplicationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func (r *resourceApplication) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}
func waitApplicationCreated(ctx context.Context, conn *opensearch.Client, id string, timeout time.Duration) (*opensearch.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.ApplicationStatusCreating)},
		Target:                    []string{string(awstypes.ApplicationStatusActive)},
		Refresh:                   statusApplication(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.GetApplicationOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitApplicationUpdated(ctx context.Context, conn *opensearch.Client, id string, timeout time.Duration) (*opensearch.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.ApplicationStatusUpdating)},
		Target:                    []string{string(awstypes.ApplicationStatusActive)},
		Refresh:                   statusApplication(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.GetApplicationOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitApplicationDeleted(ctx context.Context, conn *opensearch.Client, id string, timeout time.Duration) (*opensearch.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			string(awstypes.ApplicationStatusDeleting),
			string(awstypes.ApplicationStatusActive),
		},
		Target:  []string{},
		Refresh: statusApplication(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.GetApplicationOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusApplication(ctx context.Context, conn *opensearch.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findApplicationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findApplicationByID(ctx context.Context, conn *opensearch.Client, id string) (*opensearch.GetApplicationOutput, error) {
	input := opensearch.GetApplicationInput{
		Id: aws.String(id),
	}

	out, err := conn.GetApplication(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

// Data structures for the OpenSearch Application resource
type resourceApplicationModel struct {
	framework.WithRegionModel
	ARN                      types.String                                                   `tfsdk:"arn"`
	AppConfigs               fwtypes.SetNestedObjectValueOf[appConfigModel]                 `tfsdk:"app_config"`
	DataSources              fwtypes.SetNestedObjectValueOf[dataSourceModel]                `tfsdk:"data_source"`
	IamIdentityCenterOptions fwtypes.ListNestedObjectValueOf[iamIdentityCenterOptionsModel] `tfsdk:"iam_identity_center_options"`
	ID                       types.String                                                   `tfsdk:"id"`
	Name                     types.String                                                   `tfsdk:"name"`
	Tags                     tftags.Map                                                     `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                     `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                                 `tfsdk:"timeouts"`
}

type appConfigModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type dataSourceModel struct {
	DataSourceArn         fwtypes.ARN  `tfsdk:"data_source_arn"`
	DataSourceDescription types.String `tfsdk:"data_source_description"`
}

type iamIdentityCenterOptionsModel struct {
	Enabled                                types.Bool  `tfsdk:"enabled"`
	IamIdentityCenterInstanceArn           fwtypes.ARN `tfsdk:"iam_identity_center_instance_arn"`
	IamRoleForIdentityCenterApplicationArn fwtypes.ARN `tfsdk:"iam_role_for_identity_center_application_arn"`
}

func sweepApplications(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := opensearch.ListApplicationsInput{}
	conn := client.OpenSearchClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := opensearch.NewListApplicationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.ApplicationSummaries {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceApplication, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.Id))),
			)
		}
	}

	return sweepResources, nil
}
