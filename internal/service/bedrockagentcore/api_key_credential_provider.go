// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagentcore_api_key_credential_provider", name="Api Key Credential Provider")
func newResourceAPIKeyCredentialProvider(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAPIKeyCredentialProvider{}
	return r, nil
}

const (
	ResNameAPIKeyCredentialProvider = "API Key Credential Provider"
)

type resourceAPIKeyCredentialProvider struct {
	framework.ResourceWithModel[resourceAPIKeyCredentialProviderModel]
}

func (r *resourceAPIKeyCredentialProvider) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRelative().AtParent().AtName("api_key_wo"),
					}...),
				},
			},
			"api_key_wo": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRelative().AtParent().AtName("api_key"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("api_key_wo_version"),
					}...),
				},
			},
			"api_key_wo_version": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("api_key_wo"),
					}...),
				},
			},
			"api_key_secret_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceAPIKeyCredentialProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan resourceAPIKeyCredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract apiKeyWo from the raw configuration because write‑only
	// attributes are not present in plan or state.
	var config resourceAPIKeyCredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Config.Get(ctx, &config))
	if resp.Diagnostics.HasError() {
		return
	}
	ctxWithAPIKey := context.WithValue(ctx, apiKeyContextKey{}, apiKeyContext{
		APIKey:          plan.APIKey,
		APIKeyWO:        config.APIKeyWO,
		APIKeyWOVersion: plan.APIKeyWOVersion,
	})

	var input bedrockagentcorecontrol.CreateApiKeyCredentialProviderInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctxWithAPIKey, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateApiKeyCredentialProvider(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctxWithAPIKey, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceAPIKeyCredentialProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceAPIKeyCredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAPIKeyCredentialProviderByName(ctx, conn, state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceAPIKeyCredentialProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state resourceAPIKeyCredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract apiKeyWo from the raw configuration because write‑only
	// attributes are not present in plan or state.
	var config resourceAPIKeyCredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Config.Get(ctx, &config))
	if resp.Diagnostics.HasError() {
		return
	}
	ctxWithAPIKey := context.WithValue(ctx, apiKeyContextKey{}, apiKeyContext{
		APIKey:          plan.APIKey,
		APIKeyWO:        config.APIKeyWO,
		APIKeyWOVersion: plan.APIKeyWOVersion,
	})

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateApiKeyCredentialProviderInput

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctxWithAPIKey, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateApiKeyCredentialProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctxWithAPIKey, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceAPIKeyCredentialProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceAPIKeyCredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteApiKeyCredentialProviderInput{
		Name: state.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteApiKeyCredentialProvider(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

func (r *resourceAPIKeyCredentialProvider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), req, resp)
}

func findAPIKeyCredentialProviderByName(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string) (*bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput, error) {
	input := bedrockagentcorecontrol.GetApiKeyCredentialProviderInput{
		Name: aws.String(name),
	}

	out, err := conn.GetApiKeyCredentialProvider(ctx, &input)
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

type apiKeyContextKey struct{}

type apiKeyContext struct {
	APIKey          types.String
	APIKeyWO        types.String
	APIKeyWOVersion types.Int64
}

func apiKeyFrom(ctx context.Context) (apiKeyContext, bool) {
	v := ctx.Value(apiKeyContextKey{})
	c, ok := v.(apiKeyContext)
	return c, ok
}

type resourceAPIKeyCredentialProviderModel struct {
	framework.WithRegionModel
	ARN             fwtypes.ARN  `tfsdk:"arn"`
	APIKey          types.String `tfsdk:"api_key"`
	APIKeyWO        types.String `tfsdk:"api_key_wo"`
	APIKeyWOVersion types.Int64  `tfsdk:"api_key_wo_version"`
	APIKeySecretARN fwtypes.ARN  `tfsdk:"api_key_secret_arn" autoflex:"-"`
	Name            types.String `tfsdk:"name"`
}

var (
	_ flex.Flattener     = &resourceAPIKeyCredentialProviderModel{}
	_ flex.TypedExpander = &resourceAPIKeyCredentialProviderModel{}
)

func (m *resourceAPIKeyCredentialProviderModel) Flatten(ctxWithAPIKey context.Context, v any) (diags diag.Diagnostics) {
	if aContext, ok := apiKeyFrom(ctxWithAPIKey); ok {
		m.APIKey = aContext.APIKey
		m.APIKeyWO = aContext.APIKeyWO
		m.APIKeyWOVersion = aContext.APIKeyWOVersion
	}

	switch t := v.(type) {
	case bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput:
		m.Name = flex.StringToFramework(ctxWithAPIKey, t.Name)
		m.ARN = fwtypes.ARNValue(aws.ToString(t.CredentialProviderArn))
		m.APIKeySecretARN = fwtypes.ARNValue(*t.ApiKeySecretArn.SecretArn)
	case bedrockagentcorecontrol.CreateApiKeyCredentialProviderOutput:
		m.Name = flex.StringToFramework(ctxWithAPIKey, t.Name)
		m.ARN = fwtypes.ARNValue(aws.ToString(t.CredentialProviderArn))
		m.APIKeySecretARN = fwtypes.ARNValue(*t.ApiKeySecretArn.SecretArn)
	case bedrockagentcorecontrol.UpdateApiKeyCredentialProviderOutput:
		m.Name = flex.StringToFramework(ctxWithAPIKey, t.Name)
		m.ARN = fwtypes.ARNValue(aws.ToString(t.CredentialProviderArn))
		m.APIKeySecretARN = fwtypes.ARNValue(*t.ApiKeySecretArn.SecretArn)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("flatten source type: %T", t),
		)
	}
	return diags
}

func (m resourceAPIKeyCredentialProviderModel) ExpandTo(ctxWithAPIKey context.Context, targetType reflect.Type) (result any, diags diag.Diagnostics) {
	var apiKey *string
	if aContext, ok := apiKeyFrom(ctxWithAPIKey); ok {
		if !aContext.APIKeyWO.IsNull() {
			apiKey = flex.StringFromFramework(ctxWithAPIKey, aContext.APIKeyWO)
		} else if !aContext.APIKey.IsNull() {
			apiKey = flex.StringFromFramework(ctxWithAPIKey, aContext.APIKey)
		}
	}

	switch targetType {
	case reflect.TypeFor[bedrockagentcorecontrol.CreateApiKeyCredentialProviderInput]():
		r := bedrockagentcorecontrol.CreateApiKeyCredentialProviderInput{
			ApiKey: apiKey,
			Name:   flex.StringFromFramework(ctxWithAPIKey, m.Name),
		}
		return &r, diags

	case reflect.TypeFor[bedrockagentcorecontrol.UpdateApiKeyCredentialProviderInput]():
		r := bedrockagentcorecontrol.UpdateApiKeyCredentialProviderInput{
			ApiKey: apiKey,
			Name:   flex.StringFromFramework(ctxWithAPIKey, m.Name),
		}
		return &r, diags
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("expand target type: %s", targetType.String()),
		)
		return nil, diags
	}
}
