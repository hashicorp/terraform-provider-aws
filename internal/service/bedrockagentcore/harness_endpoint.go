// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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

// @FrameworkResource("aws_bedrockagentcore_harness_endpoint", name="Harness Endpoint")
// @Tags(identifierAttribute="harness_endpoint_arn")
// @Testing(tagsTest=false)
func newHarnessEndpointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &harnessEndpointResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type harnessEndpointResource struct {
	framework.ResourceWithModel[harnessEndpointResourceModel]
	framework.WithTimeouts
}

func (r *harnessEndpointResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"harness_endpoint_arn": framework.ARNAttributeComputedOnly(),
			"harness_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"live_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,47}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.HarnessEndpointStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"target_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([1-9][0-9]{0,4})$`), ""),
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

func (r *harnessEndpointResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data harnessEndpointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	harnessID, name := fwflex.StringValueFromFramework(ctx, data.HarnessID), fwflex.StringValueFromFramework(ctx, data.EndpointName)
	var input bedrockagentcorecontrol.CreateHarnessEndpointInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateHarnessEndpoint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	out, err := waitHarnessEndpointCreated(ctx, conn, harnessID, name, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		// Taint the resource.
		response.State.SetAttribute(ctx, path.Root("harness_id"), harnessID)
		response.State.SetAttribute(ctx, path.Root(names.AttrName), name)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out.Endpoint, &data))
	if response.Diagnostics.HasError() {
		return
	}
	// The service only reports target_version while an update is transitioning;
	// once the endpoint is READY it leaves target_version empty and reflects the
	// served version in live_version. Fall back to live_version so an explicitly
	// configured target_version round-trips and an unset one gets a stable value.
	if data.TargetVersion.IsNull() {
		data.TargetVersion = data.LiveVersion
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *harnessEndpointResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data harnessEndpointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	harnessID, name := fwflex.StringValueFromFramework(ctx, data.HarnessID), fwflex.StringValueFromFramework(ctx, data.EndpointName)
	out, err := findHarnessEndpointByTwoPartKey(ctx, conn, harnessID, name)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out.Endpoint, &data))
	if response.Diagnostics.HasError() {
		return
	}
	if data.TargetVersion.IsNull() {
		data.TargetVersion = data.LiveVersion
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *harnessEndpointResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old harnessEndpointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		harnessID, name := fwflex.StringValueFromFramework(ctx, new.HarnessID), fwflex.StringValueFromFramework(ctx, new.EndpointName)
		var input bedrockagentcorecontrol.UpdateHarnessEndpointInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(create.UniqueId(ctx))

		_, err := conn.UpdateHarnessEndpoint(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}

		out, err := waitHarnessEndpointUpdated(ctx, conn, harnessID, name, r.UpdateTimeout(ctx, new.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out.Endpoint, &new))
		if response.Diagnostics.HasError() {
			return
		}
		if new.TargetVersion.IsNull() {
			new.TargetVersion = new.LiveVersion
		}
	} else {
		new.LiveVersion = old.LiveVersion
		new.Status = old.Status
		new.TargetVersion = old.TargetVersion
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *harnessEndpointResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data harnessEndpointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	harnessID, name := fwflex.StringValueFromFramework(ctx, data.HarnessID), fwflex.StringValueFromFramework(ctx, data.EndpointName)
	input := bedrockagentcorecontrol.DeleteHarnessEndpointInput{
		ClientToken:  aws.String(create.UniqueId(ctx)),
		EndpointName: aws.String(name),
		HarnessId:    aws.String(harnessID),
	}

	_, err := conn.DeleteHarnessEndpoint(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "was not found") {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	if _, err := waitHarnessEndpointDeleted(ctx, conn, harnessID, name, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}
}

func (r *harnessEndpointResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, ",")
	if len(parts) != 2 {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf(`unexpected format for import ID (%s), use: "harness_id,name"`, request.ID))
		return
	}

	harnessID, endpointName := parts[0], parts[1]

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("harness_id"), harnessID))
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrName), endpointName))
}

func waitHarnessEndpointCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, harnessID, endpointName string, timeout time.Duration) (*bedrockagentcorecontrol.GetHarnessEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.HarnessEndpointStatusCreating),
		Target:                    enum.Slice(awstypes.HarnessEndpointStatusReady),
		Refresh:                   statusHarnessEndpoint(conn, harnessID, endpointName),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetHarnessEndpointOutput); ok {
		if out.Endpoint != nil {
			retry.SetLastError(err, errors.New(aws.ToString(out.Endpoint.FailureReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitHarnessEndpointUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, harnessID, endpointName string, timeout time.Duration) (*bedrockagentcorecontrol.GetHarnessEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.HarnessEndpointStatusUpdating),
		Target:                    enum.Slice(awstypes.HarnessEndpointStatusReady),
		Refresh:                   statusHarnessEndpoint(conn, harnessID, endpointName),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetHarnessEndpointOutput); ok {
		if out.Endpoint != nil {
			retry.SetLastError(err, errors.New(aws.ToString(out.Endpoint.FailureReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitHarnessEndpointDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, harnessID, endpointName string, timeout time.Duration) (*bedrockagentcorecontrol.GetHarnessEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.HarnessEndpointStatusDeleting, awstypes.HarnessEndpointStatusReady),
		Target:  []string{},
		Refresh: statusHarnessEndpoint(conn, harnessID, endpointName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetHarnessEndpointOutput); ok {
		if out.Endpoint != nil {
			retry.SetLastError(err, errors.New(aws.ToString(out.Endpoint.FailureReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusHarnessEndpoint(conn *bedrockagentcorecontrol.Client, harnessID, endpointName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findHarnessEndpointByTwoPartKey(ctx, conn, harnessID, endpointName)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Endpoint.Status), nil
	}
}

func findHarnessEndpointByTwoPartKey(ctx context.Context, conn *bedrockagentcorecontrol.Client, harnessID, endpointName string) (*bedrockagentcorecontrol.GetHarnessEndpointOutput, error) {
	input := bedrockagentcorecontrol.GetHarnessEndpointInput{
		EndpointName: aws.String(endpointName),
		HarnessId:    aws.String(harnessID),
	}

	return findHarnessEndpoint(ctx, conn, &input)
}

func findHarnessEndpoint(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetHarnessEndpointInput) (*bedrockagentcorecontrol.GetHarnessEndpointOutput, error) {
	out, err := conn.GetHarnessEndpoint(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "was not found") {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Endpoint == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type harnessEndpointResourceModel struct {
	framework.WithRegionModel
	ARN           types.String                                       `tfsdk:"harness_endpoint_arn"`
	Description   types.String                                       `tfsdk:"description"`
	EndpointName  types.String                                       `tfsdk:"name"`
	HarnessID     types.String                                       `tfsdk:"harness_id"`
	LiveVersion   types.String                                       `tfsdk:"live_version"`
	Status        fwtypes.StringEnum[awstypes.HarnessEndpointStatus] `tfsdk:"status"`
	TargetVersion types.String                                       `tfsdk:"target_version"`
	Tags          tftags.Map                                         `tfsdk:"tags"`
	TagsAll       tftags.Map                                         `tfsdk:"tags_all"`
	Timeouts      timeouts.Value                                     `tfsdk:"timeouts"`
}
