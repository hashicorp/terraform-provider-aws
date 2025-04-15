// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_appfabric_app_authorization", name="App Authorization")
// @Tags(identifierAttribute="arn")
// @Testing(serialize=true)
// @Testing(generator=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/appfabric/types;types.AppAuthorization")
// @Testing(importIgnore="credential")
func newAppAuthorizationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &appAuthorizationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type appAuthorizationResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *appAuthorizationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_bundle_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"auth_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AuthType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auth_url": schema.StringAttribute{
				Computed: true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"persona": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
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
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"api_key_credential": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[apiKeyCredentialModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"api_key": schema.StringAttribute{
										Required:  true,
										Sensitive: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 2048),
										},
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
									"client_id": schema.StringAttribute{ // nosemgrep:ci.literal-client_id-string-constant
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 2048),
										},
									},
									names.AttrClientSecret: schema.StringAttribute{
										Required:  true,
										Sensitive: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 2048),
										},
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
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
							},
						},
						"tenant_identifier": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 1024),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
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

	input := &appfabric.CreateAppAuthorizationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating AppFabric App (%s) Authorization", data.App.ValueString()), err.Error())
	}

	input.AppBundleIdentifier = data.AppBundleARN.ValueStringPointer()
	input.ClientToken = aws.String(uuid)
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateAppAuthorization(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating AppFabric App (%s) Authorization", data.App.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.AppAuthorizationARN = fwflex.StringToFramework(ctx, output.AppAuthorization.AppAuthorizationArn)
	id, err := data.setID()
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("flattening resource ID AppFabric App (%s) Authorization", data.App.ValueString()), err.Error())
		return
	}
	data.ID = types.StringValue(id)

	appAuthorization, err := waitAppAuthorizationCreated(ctx, conn, data.AppAuthorizationARN.ValueString(), data.AppBundleARN.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for AppFabric App Authorization (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns after creation is complete.
	data.AuthURL = fwflex.StringToFramework(ctx, appAuthorization.AuthUrl)
	if err := data.parseAuthURL(); err != nil {
		response.Diagnostics.AddError("parsing Auth URL", err.Error())

		return
	}
	data.CreatedAt = fwflex.TimeToFramework(ctx, appAuthorization.CreatedAt)
	data.Persona = fwflex.StringValueToFramework(ctx, appAuthorization.Persona)
	data.UpdatedAt = fwflex.TimeToFramework(ctx, appAuthorization.UpdatedAt)

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

	output, err := findAppAuthorizationByTwoPartKey(ctx, conn, data.AppAuthorizationARN.ValueString(), data.AppBundleARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading AppFabric App Authorization (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Setting it because of the dynamic nature of Auth URL.
	data.AuthURL = fwflex.StringToFramework(ctx, output.AuthUrl)
	if err := data.parseAuthURL(); err != nil {
		response.Diagnostics.AddError("parsing Auth URL", err.Error())

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

	conn := r.Meta().AppFabricClient(ctx)

	// Check if updates are necessary based on the changed attributes
	if !old.Credential.Equal(new.Credential) || !old.Tenant.Equal(new.Tenant) {
		var credentialsData []credentialModel
		response.Diagnostics.Append(new.Credential.ElementsAs(ctx, &credentialsData, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		input := &appfabric.UpdateAppAuthorizationInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		input.AppAuthorizationIdentifier = fwflex.StringFromFramework(ctx, new.AppAuthorizationARN)
		input.AppBundleIdentifier = fwflex.StringFromFramework(ctx, new.AppBundleARN)

		_, err := conn.UpdateAppAuthorization(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating AppFabric App Authorization (%s)", new.ID.ValueString()), err.Error())

			return
		}

		appAuthorization, err := waitAppAuthorizationUpdated(ctx, conn, new.AppAuthorizationARN.ValueString(), new.AppBundleARN.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for AppFabric App Authorization (%s) update", new.ID.ValueString()), err.Error())

			return
		}

		// Set values for unknowns.
		new.AuthURL = fwflex.StringToFramework(ctx, appAuthorization.AuthUrl)
		if err := new.parseAuthURL(); err != nil {
			response.Diagnostics.AddError("parsing Auth URL", err.Error())

			return
		}
		new.UpdatedAt = fwflex.TimeToFramework(ctx, appAuthorization.UpdatedAt)
	} else {
		new.AuthURL = old.AuthURL
		new.UpdatedAt = old.UpdatedAt
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

	input := appfabric.DeleteAppAuthorizationInput{
		AppAuthorizationIdentifier: fwflex.StringFromFramework(ctx, data.AppAuthorizationARN),
		AppBundleIdentifier:        fwflex.StringFromFramework(ctx, data.AppBundleARN),
	}
	_, err := conn.DeleteAppAuthorization(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting AppFabric App Authorization (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err = waitAppAuthorizationDeleted(ctx, conn, data.AppAuthorizationARN.ValueString(), data.AppBundleARN.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for AppFabric AppAuthenticator (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func findAppAuthorizationByTwoPartKey(ctx context.Context, conn *appfabric.Client, appAuthorizationARN, appBundleIdentifier string) (*awstypes.AppAuthorization, error) {
	in := &appfabric.GetAppAuthorizationInput{
		AppAuthorizationIdentifier: aws.String(appAuthorizationARN),
		AppBundleIdentifier:        aws.String(appBundleIdentifier),
	}

	output, err := conn.GetAppAuthorization(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AppAuthorization == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return output.AppAuthorization, nil
}

func statusAppAuthorization(ctx context.Context, conn *appfabric.Client, appAuthorizationARN, appBundleIdentifier string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findAppAuthorizationByTwoPartKey(ctx, conn, appAuthorizationARN, appBundleIdentifier)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitAppAuthorizationCreated(ctx context.Context, conn *appfabric.Client, appAuthorizationARN, appBundleIdentifier string, timeout time.Duration) (*awstypes.AppAuthorization, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.AppAuthorizationStatusPendingConnect, awstypes.AppAuthorizationStatusConnected),
		Refresh: statusAppAuthorization(ctx, conn, appAuthorizationARN, appBundleIdentifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AppAuthorization); ok {
		return output, err
	}

	return nil, err
}

func waitAppAuthorizationUpdated(ctx context.Context, conn *appfabric.Client, appAuthorizationARN, appBundleIdentifier string, timeout time.Duration) (*awstypes.AppAuthorization, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.AppAuthorizationStatusConnected, awstypes.AppAuthorizationStatusPendingConnect),
		Refresh: statusAppAuthorization(ctx, conn, appAuthorizationARN, appBundleIdentifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AppAuthorization); ok {
		return output, err
	}

	return nil, err
}

func waitAppAuthorizationDeleted(ctx context.Context, conn *appfabric.Client, appAuthorizationARN, appBundleIdentifier string, timeout time.Duration) (*awstypes.AppAuthorization, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AppAuthorizationStatusConnected, awstypes.AppAuthorizationStatusPendingConnect),
		Target:  []string{},
		Refresh: statusAppAuthorization(ctx, conn, appAuthorizationARN, appBundleIdentifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AppAuthorization); ok {
		return output, err
	}

	return nil, err
}

type appAuthorizationResourceModel struct {
	App                 types.String                                     `tfsdk:"app"`
	AppAuthorizationARN types.String                                     `tfsdk:"arn"`
	AppBundleARN        fwtypes.ARN                                      `tfsdk:"app_bundle_arn"`
	AuthType            fwtypes.StringEnum[awstypes.AuthType]            `tfsdk:"auth_type"`
	AuthURL             types.String                                     `tfsdk:"auth_url"`
	CreatedAt           timetypes.RFC3339                                `tfsdk:"created_at"`
	Credential          fwtypes.ListNestedObjectValueOf[credentialModel] `tfsdk:"credential"`
	ID                  types.String                                     `tfsdk:"id"`
	Persona             types.String                                     `tfsdk:"persona"`
	Tags                tftags.Map                                       `tfsdk:"tags"`
	TagsAll             tftags.Map                                       `tfsdk:"tags_all"`
	Tenant              fwtypes.ListNestedObjectValueOf[tenantModel]     `tfsdk:"tenant"`
	Timeouts            timeouts.Value                                   `tfsdk:"timeouts"`
	UpdatedAt           timetypes.RFC3339                                `tfsdk:"updated_at"`
}

const (
	appAuthorizationResourceIDPartCount = 2
)

func (m *appAuthorizationResourceModel) InitFromID() error {
	parts, err := flex.ExpandResourceId(m.ID.ValueString(), appAuthorizationResourceIDPartCount, false)
	if err != nil {
		return err
	}

	m.AppAuthorizationARN = types.StringValue(parts[0])
	m.AppBundleARN = fwtypes.ARNValue(parts[1])

	return nil
}

func (m *appAuthorizationResourceModel) setID() (string, error) {
	parts := []string{
		m.AppAuthorizationARN.ValueString(),
		m.AppBundleARN.ValueString(),
	}

	return flex.FlattenResourceId(parts, appAuthorizationResourceIDPartCount, false)
}

var (
	_ fwflex.Expander = credentialModel{}
)

type credentialModel struct {
	ApiKeyCredential fwtypes.ListNestedObjectValueOf[apiKeyCredentialModel] `tfsdk:"api_key_credential"`
	Oauth2Credential fwtypes.ListNestedObjectValueOf[oauth2CredentialModel] `tfsdk:"oauth2_credential"`
}

func (m credentialModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.ApiKeyCredential.IsNull():
		efsStorageConfigurationData, d := m.ApiKeyCredential.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CredentialMemberApiKeyCredential
		diags.Append(fwflex.Expand(ctx, efsStorageConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.Oauth2Credential.IsNull():
		fsxStorageConfigurationData, d := m.Oauth2Credential.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CredentialMemberOauth2Credential
		diags.Append(fwflex.Expand(ctx, fsxStorageConfigurationData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
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

func (m *appAuthorizationResourceModel) parseAuthURL() error {
	if m.AuthURL.IsNull() {
		return nil
	}

	fullURL := m.AuthURL.ValueString()

	index := strings.Index(fullURL, "oauth2")
	if index == -1 {
		return fmt.Errorf("the URL does not contain the 'oauth2' substring")
	}

	baseURL := fullURL[:index+len("oauth2")]
	m.AuthURL = types.StringValue(baseURL)

	return nil
}
