// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"fmt"
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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_opensearch_application", name="Application")
// @Tags(identifierAttribute="arn")
func newApplicationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &applicationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type applicationResource struct {
	framework.ResourceWithModel[applicationResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *applicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
						"iam_identity_center_application_arn": framework.ARNAttributeComputedOnly(),
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

func (r *applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan applicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input opensearch.CreateApplicationInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.TagList = getTagsIn(ctx)

	outputCA, err := conn.CreateApplication(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating OpenSearch Application (%s)", name), err.Error())
		return
	}

	id := aws.ToString(outputCA.Id)
	outputGA, err := waitApplicationCreated(ctx, conn, id, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Application (%s) create", id), err.Error())
		return
	}

	// Set values for unknowns.
	resp.Diagnostics.Append(fwflex.Flatten(ctx, outputGA, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.ID)
	out, err := findApplicationByID(ctx, conn, id)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading OpenSearch Application (%s)", id), err.Error())
		return
	}

	// Set attributes for import.
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state applicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		id := fwflex.StringValueFromFramework(ctx, plan.ID)
		var input opensearch.UpdateApplicationInput
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateApplication(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating OpenSearch Application (%s)", id), err.Error())
			return
		}

		if _, err := waitApplicationDeleted(ctx, conn, id, r.UpdateTimeout(ctx, state.Timeouts)); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Application (%s) update", id), err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state applicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.ID)
	input := opensearch.DeleteApplicationInput{
		Id: aws.String(id),
	}
	_, err := conn.DeleteApplication(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting OpenSearch Application (%s)", id), err.Error())
		return
	}

	if _, err := waitApplicationDeleted(ctx, conn, id, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Application (%s) delete", id), err.Error())
		return
	}
}

func waitApplicationCreated(ctx context.Context, conn *opensearch.Client, id string, timeout time.Duration) (*opensearch.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationStatusCreating),
		Target:  enum.Slice(awstypes.ApplicationStatusActive),
		Refresh: statusApplication(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.GetApplicationOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitApplicationUpdated(ctx context.Context, conn *opensearch.Client, id string, timeout time.Duration) (*opensearch.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationStatusUpdating),
		Target:  enum.Slice(awstypes.ApplicationStatusActive),
		Refresh: statusApplication(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.GetApplicationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationDeleted(ctx context.Context, conn *opensearch.Client, id string, timeout time.Duration) (*opensearch.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationStatusDeleting, awstypes.ApplicationStatusActive),
		Target:  []string{},
		Refresh: statusApplication(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.GetApplicationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusApplication(conn *opensearch.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findApplicationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findApplicationByID(ctx context.Context, conn *opensearch.Client, id string) (*opensearch.GetApplicationOutput, error) {
	input := opensearch.GetApplicationInput{
		Id: aws.String(id),
	}

	return findApplication(ctx, conn, &input)
}

func findApplication(ctx context.Context, conn *opensearch.Client, input *opensearch.GetApplicationInput) (*opensearch.GetApplicationOutput, error) {
	output, err := conn.GetApplication(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

// Data structures for the OpenSearch Application resource
type applicationResourceModel struct {
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
	IamIdentityCenterApplicationArn        fwtypes.ARN `tfsdk:"iam_identity_center_application_arn"`
}
