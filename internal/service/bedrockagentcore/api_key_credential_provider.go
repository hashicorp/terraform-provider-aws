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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_api_key_credential_provider", name="Api Key Credential Provider")
func newAPIKeyCredentialProviderResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &apiKeyCredentialProviderResource{}
	return r, nil
}

type apiKeyCredentialProviderResource struct {
	framework.ResourceWithModel[apiKeyCredentialProviderResourceModel]
}

const (
	ResNameAPIKeyCredentialProvider = "API Key Credential Provider"
)

func (r *apiKeyCredentialProviderResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
			"credential_provider_arn": schema.StringAttribute{
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

func (r *apiKeyCredentialProviderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan apiKeyCredentialProviderResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	// Extract apiKeyWo from the raw configuration because write‑only
	// attributes are not present in plan or state.
	var config apiKeyCredentialProviderResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	if response.Diagnostics.HasError() {
		return
	}
	ctxWithAPIKey := context.WithValue(ctx, apiKeyContextKey{}, apiKeyContext{
		APIKey:          plan.APIKey,
		APIKeyWO:        config.APIKeyWO,
		APIKeyWOVersion: plan.APIKeyWOVersion,
	})

	var input bedrockagentcorecontrol.CreateApiKeyCredentialProviderInput
	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Expand(ctxWithAPIKey, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateApiKeyCredentialProvider(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Flatten(ctxWithAPIKey, out, &plan))
	if response.Diagnostics.HasError() {
		return
	}
	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *apiKeyCredentialProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state apiKeyCredentialProviderResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	out, err := findAPIKeyCredentialProviderByName(ctx, conn, state.Name.ValueString())
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *apiKeyCredentialProviderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state apiKeyCredentialProviderResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	// Extract apiKeyWo from the raw configuration because write‑only
	// attributes are not present in plan or state.
	var config apiKeyCredentialProviderResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	if response.Diagnostics.HasError() {
		return
	}
	ctxWithAPIKey := context.WithValue(ctx, apiKeyContextKey{}, apiKeyContext{
		APIKey:          plan.APIKey,
		APIKeyWO:        config.APIKeyWO,
		APIKeyWOVersion: plan.APIKeyWOVersion,
	})

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateApiKeyCredentialProviderInput

		smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Expand(ctxWithAPIKey, plan, &input))
		if response.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateApiKeyCredentialProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Flatten(ctxWithAPIKey, out, &plan))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *apiKeyCredentialProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state apiKeyCredentialProviderResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
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

		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

func (r *apiKeyCredentialProviderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
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

type apiKeyCredentialProviderResourceModel struct {
	framework.WithRegionModel
	APIKey                types.String `tfsdk:"api_key"`
	APIKeySecretARN       fwtypes.ARN  `tfsdk:"api_key_secret_arn" autoflex:"-"`
	APIKeyWO              types.String `tfsdk:"api_key_wo"`
	APIKeyWOVersion       types.Int64  `tfsdk:"api_key_wo_version"`
	CredentialProviderARN fwtypes.ARN  `tfsdk:"credential_provider_arn"`
	Name                  types.String `tfsdk:"name"`
}

var (
	_ fwflex.Flattener     = &apiKeyCredentialProviderResourceModel{}
	_ fwflex.TypedExpander = &apiKeyCredentialProviderResourceModel{}
)

func (m *apiKeyCredentialProviderResourceModel) Flatten(ctxWithAPIKey context.Context, v any) (diags diag.Diagnostics) {
	if aContext, ok := apiKeyFrom(ctxWithAPIKey); ok {
		m.APIKey = aContext.APIKey
		m.APIKeyWO = aContext.APIKeyWO
		m.APIKeyWOVersion = aContext.APIKeyWOVersion
	}

	switch t := v.(type) {
	case bedrockagentcorecontrol.GetApiKeyCredentialProviderOutput:
		m.Name = fwflex.StringToFramework(ctxWithAPIKey, t.Name)
		m.CredentialProviderARN = fwtypes.ARNValue(aws.ToString(t.CredentialProviderArn))
		m.APIKeySecretARN = fwtypes.ARNValue(*t.ApiKeySecretArn.SecretArn)
	case bedrockagentcorecontrol.CreateApiKeyCredentialProviderOutput:
		m.Name = fwflex.StringToFramework(ctxWithAPIKey, t.Name)
		m.CredentialProviderARN = fwtypes.ARNValue(aws.ToString(t.CredentialProviderArn))
		m.APIKeySecretARN = fwtypes.ARNValue(*t.ApiKeySecretArn.SecretArn)
	case bedrockagentcorecontrol.UpdateApiKeyCredentialProviderOutput:
		m.Name = fwflex.StringToFramework(ctxWithAPIKey, t.Name)
		m.CredentialProviderARN = fwtypes.ARNValue(aws.ToString(t.CredentialProviderArn))
		m.APIKeySecretARN = fwtypes.ARNValue(*t.ApiKeySecretArn.SecretArn)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("flatten source type: %T", t),
		)
	}
	return diags
}

func (m apiKeyCredentialProviderResourceModel) ExpandTo(ctxWithAPIKey context.Context, targetType reflect.Type) (result any, diags diag.Diagnostics) {
	var apiKey *string
	if aContext, ok := apiKeyFrom(ctxWithAPIKey); ok {
		if !aContext.APIKeyWO.IsNull() {
			apiKey = fwflex.StringFromFramework(ctxWithAPIKey, aContext.APIKeyWO)
		} else if !aContext.APIKey.IsNull() {
			apiKey = fwflex.StringFromFramework(ctxWithAPIKey, aContext.APIKey)
		}
	}

	switch targetType {
	case reflect.TypeFor[bedrockagentcorecontrol.CreateApiKeyCredentialProviderInput]():
		r := bedrockagentcorecontrol.CreateApiKeyCredentialProviderInput{
			ApiKey: apiKey,
			Name:   fwflex.StringFromFramework(ctxWithAPIKey, m.Name),
		}
		return &r, diags

	case reflect.TypeFor[bedrockagentcorecontrol.UpdateApiKeyCredentialProviderInput]():
		r := bedrockagentcorecontrol.UpdateApiKeyCredentialProviderInput{
			ApiKey: apiKey,
			Name:   fwflex.StringFromFramework(ctxWithAPIKey, m.Name),
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
