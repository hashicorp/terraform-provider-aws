// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/evs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/evs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_evs_environment", name="Environment")
// @Tags(identifierAttribute="environment_arn")
// @Testing(tagsTest=false)
func newEnvironmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &environmentResource{}

	r.SetDefaultCreateTimeout(45 * time.Minute)
	r.SetDefaultDeleteTimeout(45 * time.Minute)

	return r, nil
}

type environmentResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[environmentResourceModel]
	framework.WithTimeouts
}

func (r *environmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment_arn": framework.ARNAttributeComputedOnly(),
			"environment_id":  framework.IDAttribute(),
			"environment_name": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9_-]+`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kms_key_id": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"terms_accepted": schema.BoolAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"vcf_version": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VcfVersion](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"connectivity_info": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[connectivityInfoModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"private_route_server_peerings": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.Set{
								setvalidator.SizeBetween(2, 2),
							},
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
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

func (r *environmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data environmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	var input evs.CreateEnvironmentInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	conn := r.Meta().EVSClient(ctx)

	output, err := conn.CreateEnvironment(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating EVS Environment", err.Error())

		return
	}

	// Set values for unknowns.
	env := output.Environment
	id := aws.ToString(env.EnvironmentId)
	data.EnvironmentARN = fwflex.StringToFramework(ctx, env.EnvironmentArn)
	data.EnvironmentID = fwflex.StringValueToFramework(ctx, id)

	if _, err := waitEnvironmentCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EVS Environment (%s) create", id), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *environmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data environmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EVSClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.EnvironmentID)
	output, err := findEnvironmentByID(ctx, conn, id)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EVS Environment (%s)", id), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (r *environmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data environmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EVSClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.EnvironmentID)
	input := evs.DeleteEnvironmentInput{
		EnvironmentId: aws.String(id),
	}
	_, err := conn.DeleteEnvironment(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EVS Environment (%s)", id), err.Error())

		return
	}

	if _, err := waitEnvironmentDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EVS Environment (%s) delete", id), err.Error())

		return
	}
}

func (r *environmentResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("environment_id"), request, response)
}

func findEnvironmentByID(ctx context.Context, conn *evs.Client, id string) (*awstypes.Environment, error) {
	input := evs.GetEnvironmentInput{
		EnvironmentId: aws.String(id),
	}

	output, err := conn.GetEnvironment(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Environment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if state := output.Environment.EnvironmentState; state == awstypes.EnvironmentStateDeleted {
		return nil, &retry.NotFoundError{
			Message: string(state),
		}
	}

	return output.Environment, nil
}

func statusEnvironment(ctx context.Context, conn *evs.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findEnvironmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.EnvironmentState), nil
	}
}

func waitEnvironmentCreated(ctx context.Context, conn *evs.Client, id string, timeout time.Duration) (*awstypes.Environment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EnvironmentStateCreating),
		Target:  enum.Slice(awstypes.EnvironmentStateCreated),
		Refresh: statusEnvironment(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Environment); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateDetails)))

		return output, err
	}

	return nil, err
}

func waitEnvironmentDeleted(ctx context.Context, conn *evs.Client, id string, timeout time.Duration) (*awstypes.Environment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EnvironmentStateDeleting),
		Target:  []string{},
		Refresh: statusEnvironment(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Environment); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateDetails)))

		return output, err
	}

	return nil, err
}

type environmentResourceModel struct {
	ConnectivityInfo fwtypes.ListNestedObjectValueOf[connectivityInfoModel] `tfsdk:"connectivity_info"`
	EnvironmentARN   types.String                                           `tfsdk:"environment_arn"`
	EnvironmentID    types.String                                           `tfsdk:"environment_id"`
	EnvironmentName  types.String                                           `tfsdk:"environment_name"`
	KMSKeyID         fwtypes.ARN                                            `tfsdk:"kms_key_id"`
	Tags             tftags.Map                                             `tfsdk:"tags"`
	TagsAll          tftags.Map                                             `tfsdk:"tags_all"`
	TermsAccepted    types.Bool                                             `tfsdk:"terms_accepted"`
	Timeouts         timeouts.Value                                         `tfsdk:"timeouts"`
	VCFVersion       fwtypes.StringEnum[awstypes.VcfVersion]                `tfsdk:"vcf_version"`
	VpcID            types.String                                           `tfsdk:"vpc_id"`
}

type connectivityInfoModel struct {
	PrivateRouteServerPeerings fwtypes.SetOfString `tfsdk:"private_route_server_peerings"`
}
