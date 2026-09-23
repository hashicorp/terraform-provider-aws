// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_connect_approved_origin", name="Approved Origin")
func newApprovedOriginResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &approvedOriginResource{}

	return r, nil
}

type approvedOriginResource struct {
	framework.ResourceWithModel[approvedOriginResourceModel]
	framework.WithNoUpdate
}

func (r *approvedOriginResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrInstanceID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"origin": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *approvedOriginResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data approvedOriginResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConnectClient(ctx)

	instanceID := fwflex.StringValueFromFramework(ctx, data.InstanceID)
	origin := fwflex.StringValueFromFramework(ctx, data.Origin)
	input := connect.AssociateApprovedOriginInput{
		InstanceId: aws.String(instanceID),
		Origin:     aws.String(origin),
	}

	_, err := conn.AssociateApprovedOrigin(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Connect Approved Origin (%s/%s)", instanceID, origin), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *approvedOriginResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data approvedOriginResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConnectClient(ctx)

	instanceID := fwflex.StringValueFromFramework(ctx, data.InstanceID)
	origin := fwflex.StringValueFromFramework(ctx, data.Origin)
	_, err := findApprovedOriginByTwoPartKey(ctx, conn, instanceID, origin)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Connect Approved Origin (%s/%s)", instanceID, origin), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *approvedOriginResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data approvedOriginResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConnectClient(ctx)

	instanceID := fwflex.StringValueFromFramework(ctx, data.InstanceID)
	origin := fwflex.StringValueFromFramework(ctx, data.Origin)
	input := connect.DisassociateApprovedOriginInput{
		InstanceId: aws.String(instanceID),
		Origin:     aws.String(origin),
	}
	_, err := conn.DisassociateApprovedOrigin(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Connect Approved Origin (%s/%s)", instanceID, origin), err.Error())

		return
	}
}

func (r *approvedOriginResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const approvedOriginIDParts = 2
	parts, err := intflex.ExpandResourceId(request.ID, approvedOriginIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrInstanceID), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("origin"), parts[1])...)
}

func findApprovedOriginByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, origin string) (*string, error) {
	input := &connect.ListApprovedOriginsInput{
		InstanceId: aws.String(instanceID),
	}

	return findApprovedOrigin(ctx, conn, input, func(v string) bool {
		return v == origin
	})
}

func findApprovedOrigin(ctx context.Context, conn *connect.Client, input *connect.ListApprovedOriginsInput, filter tfslices.Predicate[string]) (*string, error) {
	output, err := findApprovedOrigins(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findApprovedOrigins(ctx context.Context, conn *connect.Client, input *connect.ListApprovedOriginsInput, filter tfslices.Predicate[string]) ([]string, error) {
	var output []string

	pages := connect.NewListApprovedOriginsPaginator(conn, input)
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

		for _, v := range page.Origins {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

type approvedOriginResourceModel struct {
	framework.WithRegionModel
	InstanceID types.String `tfsdk:"instance_id"`
	Origin     types.String `tfsdk:"origin"`
}
