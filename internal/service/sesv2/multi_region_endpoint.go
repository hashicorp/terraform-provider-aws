// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_sesv2_multi_region_endpoint", name="Multi Region Endpoint")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
// @NoImport
func newMultiRegionEndpointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &multiRegionEndpointResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type multiRegionEndpointResource struct {
	framework.ResourceWithModel[multiRegionEndpointResourceModel]
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *multiRegionEndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"endpoint_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"endpoint_id":     framework.IDAttribute(),
			"routes":          framework.ResourceComputedListOfObjectsAttribute[routeModel](ctx),
			names.AttrTags:    tftags.TagsAttributeForceNew(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"details": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[detailsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"routes_details": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[routeDetailsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrRegion: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
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

func (r *multiRegionEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var data multiRegionEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input sesv2.CreateMultiRegionEndpointInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateMultiRegionEndpoint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.EndpointName.ValueString())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, data.EndpointName.ValueString())
		return
	}

	endpoint, err := waitMultiRegionEndpointCreated(ctx, conn, data.EndpointName.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.EndpointName.ValueString())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, endpoint, &data))
	data.ARN = fwflex.StringValueToFramework(ctx, multiRegionEndpointARN(ctx, r.Meta(), data.EndpointName.ValueString()))
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *multiRegionEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var data multiRegionEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findMultiRegionEndpointByName(ctx, conn, data.EndpointName.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.EndpointName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	data.ARN = fwflex.StringValueToFramework(ctx, multiRegionEndpointARN(ctx, r.Meta(), data.EndpointName.ValueString()))
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *multiRegionEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var data multiRegionEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	input := sesv2.DeleteMultiRegionEndpointInput{
		EndpointName: data.EndpointName.ValueStringPointer(),
	}

	_, err := conn.DeleteMultiRegionEndpoint(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.EndpointName.ValueString())
		return
	}

	_, err = waitMultiRegionEndpointDeleted(ctx, conn, data.EndpointName.ValueString(), r.DeleteTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.EndpointName.ValueString())
		return
	}
}

const (
	statusCreating = string(awstypes.StatusCreating)
	statusDeleting = string(awstypes.StatusDeleting)
	statusFailed   = string(awstypes.StatusFailed)
	statusReady    = string(awstypes.StatusReady)
)

func waitMultiRegionEndpointCreated(ctx context.Context, conn *sesv2.Client, name string, timeout time.Duration) (*sesv2.GetMultiRegionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusCreating},
		Target:                    []string{statusReady},
		Refresh:                   statusMultiRegionEndpoint(conn, name),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*sesv2.GetMultiRegionEndpointOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitMultiRegionEndpointDeleted(ctx context.Context, conn *sesv2.Client, name string, timeout time.Duration) (*sesv2.GetMultiRegionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusReady, statusFailed},
		Target:  []string{},
		Refresh: statusMultiRegionEndpoint(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*sesv2.GetMultiRegionEndpointOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusMultiRegionEndpoint(conn *sesv2.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findMultiRegionEndpointByName(ctx, conn, name)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findMultiRegionEndpointByName(ctx context.Context, conn *sesv2.Client, name string) (*sesv2.GetMultiRegionEndpointOutput, error) {
	input := sesv2.GetMultiRegionEndpointInput{
		EndpointName: aws.String(name),
	}

	out, err := conn.GetMultiRegionEndpoint(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

// Helper function to build resource ARN (required to use ListTagsForResource) as the same is not returned by Get/CreateMultiRegionEndpoint
func multiRegionEndpointARN(ctx context.Context, meta *conns.AWSClient, endpointName string) string {
	return awsarn.ARN{
		Partition: meta.Partition(ctx),
		Service:   "ses",
		Region:    meta.Region(ctx),
		AccountID: meta.AccountID(ctx),
		Resource:  fmt.Sprintf("multi-region-endpoint/%s", endpointName),
	}.String()
}

type multiRegionEndpointResourceModel struct {
	framework.WithRegionModel
	ARN          types.String                                  `tfsdk:"arn"`
	EndpointID   types.String                                  `tfsdk:"endpoint_id"`
	EndpointName types.String                                  `tfsdk:"endpoint_name"`
	Details      fwtypes.ListNestedObjectValueOf[detailsModel] `tfsdk:"details"`
	Routes       fwtypes.ListNestedObjectValueOf[routeModel]   `tfsdk:"routes"`
	Tags         tftags.Map                                    `tfsdk:"tags"`
	TagsAll      tftags.Map                                    `tfsdk:"tags_all"`
	Timeouts     timeouts.Value                                `tfsdk:"timeouts"`
}

type routeModel struct {
	Region types.String `tfsdk:"region"`
}

type detailsModel struct {
	RoutesDetails fwtypes.ListNestedObjectValueOf[routeDetailsModel] `tfsdk:"routes_details"`
}

type routeDetailsModel struct {
	Region types.String `tfsdk:"region"`
}
