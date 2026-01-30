// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mpa

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mpa"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mpa/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_mpa_identity_source", name="Identity Source")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/mpa;mpa.GetIdentitySourceOutput")
func newIdentitySourceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &identitySourceResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameIdentitySource = "Identity Source"
)

type identitySourceResource struct {
	framework.ResourceWithModel[identitySourceResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *identitySourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatusCode: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatusMessage: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"identity_source_parameters": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[identitySourceParametersModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"iam_identity_center": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[iamIdentityCenterModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"instance_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									names.AttrRegion: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *identitySourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().MPAClient(ctx)

	var plan identitySourceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input mpa.CreateIdentitySourceInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)
	input.ClientToken = aws.String(sdkid.UniqueId())

	out, err := conn.CreateIdentitySource(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	plan.ID = fwflex.StringToFramework(ctx, out.IdentitySourceArn)
	plan.ARN = fwflex.StringToFramework(ctx, out.IdentitySourceArn)

	output, err := waitIdentitySourceCreated(ctx, conn, plan.ID.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *identitySourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Get the resource from AWS
	// 4. Remove resource from state if it is not found
	// 5. Set the arguments and attributes
	// 6. Set the state

	conn := r.Meta().MPAClient(ctx)

	var state identitySourceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findIdentitySourceByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *identitySourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().MPAClient(ctx)

	var state identitySourceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := mpa.DeleteIdentitySourceInput{
		IdentitySourceArn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteIdentitySource(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	if _, err := waitIdentitySourceDeleted(ctx, conn, state.ID.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func statusIdentitySource(conn *mpa.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findIdentitySourceByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return output, string(output.Status), nil
	}
}

func waitIdentitySourceCreated(ctx context.Context, conn *mpa.Client, id string, timeout time.Duration) (*mpa.GetIdentitySourceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IdentitySourceStatusCreating),
		Target:  enum.Slice(awstypes.IdentitySourceStatusActive),
		Refresh: statusIdentitySource(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*mpa.GetIdentitySourceOutput); ok {
		return output, err
	}

	return nil, err
}

func waitIdentitySourceDeleted(ctx context.Context, conn *mpa.Client, id string, timeout time.Duration) (*mpa.GetIdentitySourceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IdentitySourceStatusDeleting, awstypes.IdentitySourceStatusActive),
		Target:  []string{},
		Refresh: statusIdentitySource(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*mpa.GetIdentitySourceOutput); ok {
		return output, err
	}

	return nil, err
}

func findIdentitySourceByID(ctx context.Context, conn *mpa.Client, id string) (*mpa.GetIdentitySourceOutput, error) {
	input := mpa.GetIdentitySourceInput{
		IdentitySourceArn: aws.String(id),
	}

	out, err := conn.GetIdentitySource(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
type identitySourceResourceModel struct {
	framework.WithRegionModel

	ARN                      types.String                                                   `tfsdk:"arn"`
	CreationTime             timetypes.RFC3339                                              `tfsdk:"creation_time"`
	ID                       types.String                                                   `tfsdk:"id"`
	IdentitySourceParameters fwtypes.ListNestedObjectValueOf[identitySourceParametersModel] `tfsdk:"identity_source_parameters"`
	Name                     types.String                                                   `tfsdk:"name"`
	Status                   types.String                                                   `tfsdk:"status"`
	StatusCode               types.String                                                   `tfsdk:"status_code"`
	StatusMessage            types.String                                                   `tfsdk:"status_message"`
	Tags                     tftags.Map                                                     `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                     `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                                 `tfsdk:"timeouts"`
}

type identitySourceParametersModel struct {
	IamIdentityCenter fwtypes.ListNestedObjectValueOf[iamIdentityCenterModel] `tfsdk:"iam_identity_center"`
}

type iamIdentityCenterModel struct {
	InstanceArn fwtypes.ARN  `tfsdk:"instance_arn"`
	Region      types.String `tfsdk:"region"`
}

// TIP: ==== SWEEPERS ====
// When acceptance testing resources, interrupted or failed tests may
// leave behind orphaned resources in an account. To facilitate cleaning
// up lingering resources, each resource implementation should include
// a corresponding "sweeper" function.
//
// The sweeper function lists all resources of a given type and sets the
// appropriate identifers required to delete the resource via the Delete
// method implemented above.
//
// Once the sweeper function is implemented, register it in sweep.go
// as follows:
//
//	awsv2.Register("aws_mpa_identity_source", sweepIdentitySources)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
func sweepIdentitySources(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := mpa.ListIdentitySourcesInput{}
	conn := client.MPAClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := mpa.NewListIdentitySourcesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.IdentitySources {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newIdentitySourceResource, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.IdentitySourceArn))),
			)
		}
	}

	return sweepResources, nil
}
