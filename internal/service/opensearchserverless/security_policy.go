package opensearchserverless

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	sdkstructure "github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceSecurityPolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceSecurityPolicy{}, nil
}

type resourceSecurityPolicyData struct {
	Description   types.String `tfsdk:"description"`
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Policy        types.String `tfsdk:"policy"`
	PolicyVersion types.String `tfsdk:"policy_version"`
	Type          types.String `tfsdk:"type"`
}

const (
	ResNameSecurityPolicy = "Security Policy"
	idSeparator           = "/"
)

type resourceSecurityPolicy struct {
	framework.ResourceWithConfigure
}

func (r *resourceSecurityPolicy) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_opensearchserverless_security_policy"
}

func (r *resourceSecurityPolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
				},
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
				},
			},
			"policy": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 20480),
				},
			},
			"policy_version": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.SecurityPolicyType](),
				},
			},
		},
	}
}

func (r *resourceSecurityPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceSecurityPolicyData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient()

	in := &opensearchserverless.CreateSecurityPolicyInput{
		ClientToken: aws.String(sdkresource.UniqueId()),
		Name:        aws.String(plan.Name.ValueString()),
		Policy:      aws.String(plan.Policy.ValueString()),
		Type:        awstypes.SecurityPolicyType(plan.Type.ValueString()),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	out, err := conn.CreateSecurityPolicy(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionCreating, ResNameSecurityPolicy, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	state := plan
	resp.Diagnostics.Append(state.refreshFromOutput(ctx, r.Meta(), out.SecurityPolicyDetail)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSecurityPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchServerlessClient()

	var state resourceSecurityPolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSecurityPolicyByNameAndType(ctx, conn, state.ID.ValueString(), state.Type.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(state.refreshFromOutput(ctx, r.Meta(), out)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSecurityPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchServerlessClient()

	var plan, state resourceSecurityPolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.Policy.Equal(state.Policy) {
		input := &opensearchserverless.UpdateSecurityPolicyInput{
			ClientToken:   aws.String(sdkresource.UniqueId()),
			Name:          flex.StringFromFramework(ctx, plan.Name),
			PolicyVersion: flex.StringFromFramework(ctx, plan.PolicyVersion),
			Type:          awstypes.SecurityPolicyType(plan.Type.ValueString()),
		}

		if !plan.Description.Equal(state.Description) {
			input.Description = aws.String(plan.Description.ValueString())
		}

		out, err := conn.UpdateSecurityPolicy(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating Security Policy (%s)", plan.Name.ValueString()), err.Error())
			return
		}
		resp.Diagnostics.Append(state.refreshFromOutput(ctx, r.Meta(), out.SecurityPolicyDetail)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceSecurityPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchServerlessClient()

	var state resourceSecurityPolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteSecurityPolicy(ctx, &opensearchserverless.DeleteSecurityPolicyInput{
		ClientToken: aws.String(sdkresource.UniqueId()),
		Name:        aws.String(state.Name.ValueString()),
		Type:        awstypes.SecurityPolicyType(state.Type.ValueString()),
	})
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionDeleting, ResNameSecurityPolicy, state.Name.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceSecurityPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, idSeparator)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err := fmt.Errorf("unexpected format for ID (%[1]s), expected security-policy-name%[2]ssecurity-policy-type", req.ID, idSeparator)
		resp.Diagnostics.AddError(fmt.Sprintf("importing Security Policy (%s)", req.ID), err.Error())
		return
	}

	state := resourceSecurityPolicyData{
		Name: types.StringValue(parts[0]),
		Type: types.StringValue(parts[1]),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// refreshFromOutput writes state data from an AWS response object
func (rd *resourceSecurityPolicyData) refreshFromOutput(ctx context.Context, meta *conns.AWSClient, out *awstypes.SecurityPolicyDetail) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	rd.ID = flex.StringToFramework(ctx, out.Name)
	rd.Description = flex.StringToFramework(ctx, out.Description)
	rd.Name = flex.StringToFramework(ctx, out.Name)
	rd.Type = flex.StringValueToFramework(ctx, out.Type)
	rd.PolicyVersion = flex.StringToFramework(ctx, out.PolicyVersion)

	policyBytes, err := out.Policy.MarshalSmithyDocument()
	if err != nil {
		diags.AddError(fmt.Sprintf("refreshing state for Security Policy (%s)", rd.Name), err.Error())
		return diags
	}

	p := string(policyBytes)

	p, err = verify.SecondJSONUnlessEquivalent(*flex.StringFromFramework(ctx, rd.Policy), p)
	if err != nil {
		diags.AddError(fmt.Sprintf("refreshing state for Security Policy (%s)", rd.Name), err.Error())
	}

	p, err = sdkstructure.NormalizeJsonString(p)
	if err != nil {
		diags.AddError(fmt.Sprintf("refreshing state for Security Policy (%s)", rd.Name), err.Error())
	}

	rd.Policy = flex.StringToFramework(ctx, &p)

	return diags
}
