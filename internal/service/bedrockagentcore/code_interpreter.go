// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_code_interpreter", name="Code Interpreter")
// @Tags(identifierAttribute="code_interpreter_arn")
// @Testing(tagsTest=false)
func newCodeInterpreterResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &codeInterpreterResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type codeInterpreterResource struct {
	framework.ResourceWithModel[codeInterpreterResourceModel]
	framework.WithTimeouts
}

func (r *codeInterpreterResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"code_interpreter_arn": framework.ARNAttributeComputedOnly(),
			"code_interpreter_id":  framework.IDAttribute(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			names.AttrExecutionRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrNetworkConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[codeInterpreterNetworkConfigurationModel](ctx),
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
						"network_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.CodeInterpreterNetworkMode](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrVPCConfig: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrSecurityGroups: schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.RequiresReplace(),
										},
									},
									names.AttrSubnets: schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.RequiresReplace(),
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

func (r *codeInterpreterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data codeInterpreterResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreateCodeInterpreterInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	var (
		out *bedrockagentcorecontrol.CreateCodeInterpreterOutput
		err error
	)
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		out, err = conn.CreateCodeInterpreter(ctx, &input)

		// IAM propagation.
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "CodeInterpreter role validation failed") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	codeInterpreterID := aws.ToString(out.CodeInterpreterId)

	if _, err := waitCodeInterpreterCreated(ctx, conn, codeInterpreterID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, codeInterpreterID)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *codeInterpreterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data codeInterpreterResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	codeInterpreterID := fwflex.StringValueFromFramework(ctx, data.CodeInterpreterID)
	out, err := findCodeInterpreterByID(ctx, conn, codeInterpreterID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, codeInterpreterID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *codeInterpreterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data codeInterpreterResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	codeInterpreterID := fwflex.StringValueFromFramework(ctx, data.CodeInterpreterID)
	input := bedrockagentcorecontrol.DeleteCodeInterpreterInput{
		CodeInterpreterId: aws.String(codeInterpreterID),
	}

	_, err := conn.DeleteCodeInterpreter(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, codeInterpreterID)
		return
	}

	if _, err := waitCodeInterpreterDeleted(ctx, conn, codeInterpreterID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, codeInterpreterID)
		return
	}
}

func (r *codeInterpreterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("code_interpreter_id"), request, response)
}

func waitCodeInterpreterCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetCodeInterpreterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.CodeInterpreterStatusCreating),
		Target:                    enum.Slice(awstypes.CodeInterpreterStatusReady),
		Refresh:                   statusCodeInterpreter(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetCodeInterpreterOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitCodeInterpreterDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetCodeInterpreterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CodeInterpreterStatusDeleting, awstypes.CodeInterpreterStatusReady),
		Target:  []string{},
		Refresh: statusCodeInterpreter(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetCodeInterpreterOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusCodeInterpreter(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findCodeInterpreterByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findCodeInterpreterByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetCodeInterpreterOutput, error) {
	input := bedrockagentcorecontrol.GetCodeInterpreterInput{
		CodeInterpreterId: aws.String(id),
	}

	return findCodeInterpreter(ctx, conn, &input)
}

func findCodeInterpreter(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetCodeInterpreterInput) (*bedrockagentcorecontrol.GetCodeInterpreterOutput, error) {
	out, err := conn.GetCodeInterpreter(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type codeInterpreterResourceModel struct {
	framework.WithRegionModel
	CodeInterpreterARN   types.String                                                              `tfsdk:"code_interpreter_arn"`
	CodeInterpreterID    types.String                                                              `tfsdk:"code_interpreter_id"`
	Description          types.String                                                              `tfsdk:"description"`
	ExecutionRoleARN     fwtypes.ARN                                                               `tfsdk:"execution_role_arn"`
	Name                 types.String                                                              `tfsdk:"name"`
	NetworkConfiguration fwtypes.ListNestedObjectValueOf[codeInterpreterNetworkConfigurationModel] `tfsdk:"network_configuration"`
	Tags                 tftags.Map                                                                `tfsdk:"tags"`
	TagsAll              tftags.Map                                                                `tfsdk:"tags_all"`
	Timeouts             timeouts.Value                                                            `tfsdk:"timeouts"`
}

type codeInterpreterNetworkConfigurationModel struct {
	NetworkMode fwtypes.StringEnum[awstypes.CodeInterpreterNetworkMode] `tfsdk:"network_mode"`
	VPCConfig   fwtypes.ListNestedObjectValueOf[vpcConfigModel]         `tfsdk:"vpc_config"`
}
