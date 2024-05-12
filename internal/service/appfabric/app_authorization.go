// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	// tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="App Authorization")
func newAppAuthorizationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &appAuthorizationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAppAuthorization = "App Authorization"
)

type appAuthorizationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *appAuthorizationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_appfabric_app_authorization"
}

func (r *appAuthorizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app_bundle_identifier": schema.StringAttribute{
				Required: true,
			},
			"app_bundle_arn": framework.ARNAttributeComputedOnly(),
			"app": schema.StringAttribute{
				Required: true,
			},
			"auth_type": schema.StringAttribute{
				Required: true,
			},
			"auth_url": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID:  framework.IDAttribute(),
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			// names.AttrTags:    tftags.TagsAttribute(),
			// names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"persona": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"credential": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[credentialModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					// listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"api_key_credential": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[apiKeyCredentialModel](ctx),
							Validators: []validator.List{
								// listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"api_key": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"oauth2_credential": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[oauth2CredentialModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"client_id": schema.StringAttribute{
										Required: true,
									},
									"client_secret": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"tenant": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tenantModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tenant_display_name": schema.StringAttribute{
							Required: true,
						},
						"tenant_identifier": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *appAuthorizationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data appAuthorizationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	var credentialsData []credentialModel
	response.Diagnostics.Append(data.Credential.ElementsAs(ctx, &credentialsData, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	credential, d := expandCredentialsValue(ctx, credentialsData)
	response.Diagnostics.Append(d...)

	input := &appfabric.CreateAppAuthorizationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Credential = credential
	// input.Tags = getTagsIn(ctx)

	output, err := conn.CreateAppAuthorization(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionCreating, ResNameAppAuthorization, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	appAuth := output.AppAuthorization
	data.AppAuthorizationArn = fwflex.StringToFramework(ctx, appAuth.AppAuthorizationArn)
	data.setID()

	_, err = waitAppAuthorizationCreated(ctx, conn, data.AppAuthorizationArn.ValueString(), data.AppBundleIdentifier.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for App Fabric App Authorization (%s) to be created", data.AppAuthorizationArn.ValueString()), err.Error())

		return
	}

	// Set values for unknowns after creation is complete.
	data.AppBundleArn = fwflex.StringToFramework(ctx, appAuth.AppBundleArn)
	data.Status = fwflex.StringValueToFramework(ctx, appAuth.Status)
	data.Persona = fwflex.StringValueToFramework(ctx, appAuth.Persona)
	data.AuthUrl = fwflex.StringToFramework(ctx, appAuth.AuthUrl)
	data.CreatedAt = fwflex.TimeToFramework(ctx, appAuth.CreatedAt)
	data.UpdatedAt = fwflex.TimeToFramework(ctx, appAuth.UpdatedAt)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *appAuthorizationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data appAuthorizationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	output, err := findAppAuthorizationByTwoPartKey(ctx, conn, data.AppAuthorizationArn.ValueString(), data.AppBundleIdentifier.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading App Fabric App Authorization ID  (%s)", data.AppAuthorizationArn.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *appAuthorizationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new appAuthorizationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := new.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	if !old.App.Equal(new.App) ||
		!old.AuthType.Equal(new.AuthType) ||
		!old.Credential.Equal(new.Credential) ||
		!old.Tenant.Equal(new.Tenant) ||
		// !old.Status.Equal(new.Status) ||
		!old.AppBundleIdentifier.Equal(new.AppBundleIdentifier) {
		// !old.Tags.Equal(new.Tags) {

		var credentialsData []credentialModel
		response.Diagnostics.Append(new.Credential.ElementsAs(ctx, &credentialsData, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		credential, d := expandCredentialsValue(ctx, credentialsData)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}

		input := &appfabric.UpdateAppAuthorizationInput{
			AppAuthorizationIdentifier: new.AppAuthorizationArn.ValueStringPointer(),
			AppBundleIdentifier:        new.AppBundleIdentifier.ValueStringPointer(),
		}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}
		// Additional fields.
		input.Credential = credential

		_, err := conn.UpdateAppAuthorization(ctx, input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AppFabric, create.ErrActionUpdating, ResNameAppAuthorization, new.ID.String(), err),
				err.Error(),
			)
			return
		}
		_, err = waitAppAuthorizationUpdated(ctx, conn, new.AppAuthorizationArn.ValueString(), new.AppBundleIdentifier.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *appAuthorizationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data appAuthorizationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	_, err := conn.DeleteAppAuthorization(ctx, &appfabric.DeleteAppAuthorizationInput{
		AppAuthorizationIdentifier: aws.String(data.AppAuthorizationArn.ValueString()),
		AppBundleIdentifier:        aws.String(data.AppBundleIdentifier.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting App Fabric AppAuthorizations (%s)", data.AppAuthorizationArn.ValueString()), err.Error())

		return
	}

	if _, err = waitAppAuthorizationDeleted(ctx, conn, data.AppAuthorizationArn.ValueString(), data.AppBundleIdentifier.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for App Fabric AppAuthenticator (%s) delete", data.AppAuthorizationArn.ValueString()), err.Error())

		return
	}
}

func (r *appAuthorizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitAppAuthorizationCreated(ctx context.Context, conn *appfabric.Client, appAuthorizationArn, appBundleIdentifier string, timeout time.Duration) (*awstypes.AppAuthorization, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.AppAuthorizationStatusPendingConnect, awstypes.AppAuthorizationStatusConnected),
		Refresh:                   statusAppAuthorization(ctx, conn, appAuthorizationArn, appBundleIdentifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AppAuthorization); ok {
		return out, err
	}

	return nil, err
}

func waitAppAuthorizationUpdated(ctx context.Context, conn *appfabric.Client, appAuthorizationArn, appBundleIdentifier string, timeout time.Duration) (*awstypes.AppAuthorization, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.AppAuthorizationStatusConnected, awstypes.AppAuthorizationStatusPendingConnect),
		Refresh:                   statusAppAuthorization(ctx, conn, appAuthorizationArn, appBundleIdentifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AppAuthorization); ok {
		return out, err
	}

	return nil, err
}

func waitAppAuthorizationDeleted(ctx context.Context, conn *appfabric.Client, appAuthorizationArn, appBundleIdentifier string, timeout time.Duration) (*awstypes.AppAuthorization, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AppAuthorizationStatusConnected, awstypes.AppAuthorizationStatusPendingConnect),
		Target:  []string{},
		Refresh: statusAppAuthorization(ctx, conn, appAuthorizationArn, appBundleIdentifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AppAuthorization); ok {
		return out, err
	}

	return nil, err
}

func statusAppAuthorization(ctx context.Context, conn *appfabric.Client, appAuthorizationArn, appBundleIdentifier string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findAppAuthorizationByTwoPartKey(ctx, conn, appAuthorizationArn, appBundleIdentifier)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findAppAuthorizationByTwoPartKey(ctx context.Context, conn *appfabric.Client, appAuthorizationArn, appBundleIdentifier string) (*awstypes.AppAuthorization, error) {
	in := &appfabric.GetAppAuthorizationInput{
		AppAuthorizationIdentifier: aws.String(appAuthorizationArn),
		AppBundleIdentifier:        aws.String(appBundleIdentifier),
	}

	out, err := conn.GetAppAuthorization(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.AppAuthorization == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.AppAuthorization, nil
}

func expandCredentialsValue(ctx context.Context, credentialModels []credentialModel) (awstypes.Credential, diag.Diagnostics) {
	credentials := []awstypes.Credential{}
	var diags diag.Diagnostics

	for _, item := range credentialModels {
		if !item.ApiKeyCredential.IsNull() && (len(item.ApiKeyCredential.Elements()) > 0) {
			var apiKey []apiKeyCredentialModel
			diags.Append(item.ApiKeyCredential.ElementsAs(ctx, &apiKey, false)...)
			apiKeycredential := expandAppAuthorizationApiKeyCredential(ctx, apiKey)
			credentials = append(credentials, apiKeycredential)
		}
		if (!item.Oauth2Credential.IsNull()) && (len(item.Oauth2Credential.Elements()) > 0) {
			var oath2Credentials []oauth2CredentialModel
			diags.Append(item.Oauth2Credential.ElementsAs(ctx, &oath2Credentials, false)...)
			oath2Credential := expandAppAuthorizationOauth2Credential(ctx, oath2Credentials)
			credentials = append(credentials, oath2Credential)
		}
	}

	return credentials[0], diags
}

func expandAppAuthorizationApiKeyCredential(ctx context.Context, apiKeyCredential []apiKeyCredentialModel) *awstypes.CredentialMemberApiKeyCredential {
	if len(apiKeyCredential) == 0 {
		return nil
	}

	return &awstypes.CredentialMemberApiKeyCredential{
		Value: awstypes.ApiKeyCredential{
			ApiKey: fwflex.StringFromFramework(ctx, apiKeyCredential[0].ApiKey),
		},
	}
}

func expandAppAuthorizationOauth2Credential(ctx context.Context, oauth2Credential []oauth2CredentialModel) *awstypes.CredentialMemberOauth2Credential {
	if len(oauth2Credential) == 0 {
		return nil
	}

	return &awstypes.CredentialMemberOauth2Credential{
		Value: awstypes.Oauth2Credential{
			ClientId:     fwflex.StringFromFramework(ctx, oauth2Credential[0].ClientId),
			ClientSecret: fwflex.StringFromFramework(ctx, oauth2Credential[0].ClientSecret),
		},
	}
}

// func flattenAppAuthorizationCredentialsModel(ctx context.Context, apiObject awstypes.Credential) (types.Object, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	result := types.ObjectUnknown(credentialModelAttrTypes)

// 	obj := map[string]attr.Value{}

// 	switch v := apiObject.(type) {
// 	case *awstypes.CredentialMemberApiKeyCredential:
// 		credentials, d := flattenAppAuthorizationCredentialsResourceModel(ctx, &v.Value, nil, "apiKey")
// 		diags.Append(d...)
// 		if d.HasError() {
// 			return result, diags
// 		}
// 		obj = map[string]attr.Value{
// 			"api_key_credential": credentials,
// 			"oauth2_credential":  types.ListNull(oauth2CredentialModelAttrTypes),
// 		}
// 	case *awstypes.CredentialMemberOauth2Credential:
// 		credentials, d := flattenAppAuthorizationCredentialsResourceModel(ctx, nil, &v.Value, "oath2")
// 		diags.Append(d...)
// 		if d.HasError() {
// 			return result, diags
// 		}
// 		obj = map[string]attr.Value{
// 			"api_key_credential": types.ListNull(apiKeyCredentialModelAttrTypes),
// 			"oauth2_credential":  credentials,
// 		}
// 	}

// 	result, d := types.ObjectValue(credentialModelAttrTypes, obj)
// 	diags.Append(d...)

// 	return result, diags
// }

// func flattenAppAuthorizationCredentialsResourceModel(ctx context.Context, apiKeyApiObject *awstypes.ApiKeyCredential, oath2ApiObject *awstypes.Oauth2Credential, credentialType string) (types.List, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	var elemType types.ObjectType
// 	var obj map[string]attr.Value
// 	var objVal basetypes.ObjectValue

// 	if credentialType == "apiKey" {
// 		var d diag.Diagnostics
// 		elemType = types.ObjectType{AttrTypes: apiKeyModelAttrTypes}
// 		obj = map[string]attr.Value{
// 			"api_key": fwflex.StringToFramework(ctx, apiKeyApiObject.ApiKey),
// 		}
// 		objVal, d = types.ObjectValue(apiKeyModelAttrTypes, obj)
// 		diags.Append(d...)
// 	} else if credentialType == "oath2" {
// 		var d diag.Diagnostics
// 		elemType = types.ObjectType{AttrTypes: oauth2ModelAttrTypes}
// 		obj = map[string]attr.Value{
// 			"client_id":     fwflex.StringToFramework(ctx, oath2ApiObject.ClientId),
// 			"client_secret": fwflex.StringToFramework(ctx, oath2ApiObject.ClientSecret),
// 		}
// 		objVal, d = types.ObjectValue(oauth2ModelAttrTypes, obj)
// 		diags.Append(d...)
// 	}

// 	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
// 	diags.Append(d...)

// 	return listVal, diags
// }

// var (
// 	apiKeyModelAttrTypes = map[string]attr.Type{
// 		"api_key": types.StringType,
// 	}

// 	oauth2ModelAttrTypes = map[string]attr.Type{
// 		"client_id":     types.StringType,
// 		"client_secret": types.StringType,
// 	}

// 	apiKeyCredentialModelAttrTypes = types.ObjectType{AttrTypes: apiKeyModelAttrTypes}
// 	oauth2CredentialModelAttrTypes = types.ObjectType{AttrTypes: oauth2ModelAttrTypes}

// 	credentialModelAttrTypes = map[string]attr.Type{
// 		"api_key_credential": types.ListType{ElemType: apiKeyCredentialModelAttrTypes},
// 		"oauth2_credential":  types.ListType{ElemType: oauth2CredentialModelAttrTypes},
// 	}
// )

type appAuthorizationResourceModel struct {
	AppAuthorizationArn types.String                                     `tfsdk:"arn"`
	AppBundleIdentifier types.String                                     `tfsdk:"app_bundle_identifier"`
	App                 types.String                                     `tfsdk:"app"`
	ID                  types.String                                     `tfsdk:"id"`
	AppBundleArn        types.String                                     `tfsdk:"app_bundle_arn"`
	AuthType            types.String                                     `tfsdk:"auth_type"`
	AuthUrl             types.String                                     `tfsdk:"auth_url"`
	CreatedAt           timetypes.RFC3339                                `tfsdk:"created_at"`
	Credential          fwtypes.ListNestedObjectValueOf[credentialModel] `tfsdk:"credential"`
	Tenant              fwtypes.ListNestedObjectValueOf[tenantModel]     `tfsdk:"tenant"`
	Timeouts            timeouts.Value                                   `tfsdk:"timeouts"`
	UpdatedAt           timetypes.RFC3339                                `tfsdk:"updated_at"`
	Persona             types.String                                     `tfsdk:"persona"`
	Status              types.String                                     `tfsdk:"status"`
	// Tags                  types.Map                                                `tfsdk:"tags"`
	// TagsAll               types.Map                                                `tfsdk:"tags_all"`
}

const (
	appAuthorizationResourceIDPartCount = 2
)

func (m *appAuthorizationResourceModel) InitFromID() error {
	parts, err := flex.ExpandResourceId(m.ID.ValueString(), appAuthorizationResourceIDPartCount, false)
	if err != nil {
		return err
	}

	m.AppAuthorizationArn = types.StringValue(parts[0])
	m.AppBundleIdentifier = types.StringValue(parts[1])

	return nil
}

func (m *appAuthorizationResourceModel) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.AppAuthorizationArn.ValueString(), m.AppBundleIdentifier.ValueString()}, appAuthorizationResourceIDPartCount, false)))
}

type credentialModel struct {
	ApiKeyCredential fwtypes.ListNestedObjectValueOf[apiKeyCredentialModel] `tfsdk:"api_key_credential"`
	Oauth2Credential fwtypes.ListNestedObjectValueOf[oauth2CredentialModel] `tfsdk:"oauth2_credential"`
}

type apiKeyCredentialModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

type oauth2CredentialModel struct {
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

type tenantModel struct {
	TenantDisplayName types.String `tfsdk:"tenant_display_name"`
	TenantIdentifier  types.String `tfsdk:"tenant_identifier"`
}
