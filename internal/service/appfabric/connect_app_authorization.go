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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Connect App Authorization")
func newConnectAppAuthorizationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceConnectAppAuthorization{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameConnectAppAuthorization = "Connect App Authorization"
)

type resourceConnectAppAuthorization struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[connectAppAuthorizationResourceModel]
	framework.WithNoOpDelete
	framework.WithTimeouts
}

func (r *resourceConnectAppAuthorization) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_appfabric_connect_app_authorization"
}

func (r *resourceConnectAppAuthorization) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Computed: true,
			},
			"app_bundle_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_authorization_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tenant": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[connectionTenantModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"tenant_display_name": types.StringType,
						"tenant_identifier":   types.StringType,
					},
				},
			},
			"id": framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"auth_request": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[authRequestModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							Required: true,
						},
						"redirect_uri": schema.StringAttribute{
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

func (r *resourceConnectAppAuthorization) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data connectAppAuthorizationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	input := &appfabric.ConnectAppAuthorizationInput{
		AppBundleIdentifier:        aws.String(data.AppBundleARN.ValueString()),
		AppAuthorizationIdentifier: aws.String(data.AppAuthorizationARN.ValueString()),
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.ConnectAppAuthorization(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("Creating Connection for App Authorization", err.Error())

		return
	}

	// Set values for unknowns.
	appAuthorization := output.AppAuthorizationSummary
	data.AppBundleARN = fwflex.StringToFramework(ctx, appAuthorization.AppBundleArn)
	data.AppAuthorizationARN = fwflex.StringToFramework(ctx, appAuthorization.AppAuthorizationArn)
	data.setID()

	con, err := waitConnectAppAuthorizationCreated(ctx, conn, data.AppAuthorizationARN.ValueString(), data.AppBundleARN.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for App Authorization (%s) Conect", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	var tenant connectionTenantModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, con.Tenant, &tenant)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Tenant = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tenant)
	data.App = fwflex.StringToFramework(ctx, appAuthorization.App)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *resourceConnectAppAuthorization) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data connectAppAuthorizationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	output, err := findConnectAppAuthorizationByAppAuth(ctx, conn, data.AppAuthorizationARN.ValueString(), data.AppBundleARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading App Fabric AppAuthorization Connection ID  (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	var tenant connectionTenantModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Tenant, &tenant)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Tenant = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tenant)
	data.App = fwflex.StringToFramework(ctx, output.App)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func waitConnectAppAuthorizationCreated(ctx context.Context, conn *appfabric.Client, appAuthorizationARN, appBundleArn string, timeout time.Duration) (*awstypes.AppAuthorization, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.AppAuthorizationStatusConnected),
		Refresh:                   statusConnectAppAuthorization(ctx, conn, appAuthorizationARN, appBundleArn),
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

func statusConnectAppAuthorization(ctx context.Context, conn *appfabric.Client, appAuthorizationARN, appBundleArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findConnectAppAuthorizationByAppAuth(ctx, conn, appAuthorizationARN, appBundleArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findConnectAppAuthorizationByAppAuth(ctx context.Context, conn *appfabric.Client, appAuthorizationARN, appBundleIdentifier string) (*awstypes.AppAuthorization, error) {
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

	if output == nil || output.AppAuthorization == nil ||  output.AppAuthorization.Status != awstypes.AppAuthorizationStatusConnected {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return output.AppAuthorization, nil
}

type connectAppAuthorizationResourceModel struct {
	App                 types.String                                           `tfsdk:"app"`
	AppAuthorizationARN types.String                                           `tfsdk:"app_authorization_arn"`
	AppBundleARN        types.String                                           `tfsdk:"app_bundle_arn"`
	AuthRequest         fwtypes.ListNestedObjectValueOf[authRequestModel]      `tfsdk:"auth_request"`
	Tenant              fwtypes.ListNestedObjectValueOf[connectionTenantModel] `tfsdk:"tenant"`
	ID                  types.String                                           `tfsdk:"id"`
	Timeouts            timeouts.Value                                         `tfsdk:"timeouts"`
}

const (
	connectAppAuthorizationIDPartCount = 2
)

func (m *connectAppAuthorizationResourceModel) InitFromID() error {
	parts, err := flex.ExpandResourceId(m.ID.ValueString(), connectAppAuthorizationIDPartCount, false)
	if err != nil {
		return err
	}

	m.AppAuthorizationARN = types.StringValue(parts[0])
	m.AppBundleARN = types.StringValue(parts[1])

	return nil
}

func (m *connectAppAuthorizationResourceModel) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.AppAuthorizationARN.ValueString(), m.AppBundleARN.ValueString()}, connectAppAuthorizationIDPartCount, false)))
}

type authRequestModel struct {
	Code        types.String `tfsdk:"code"`
	RedirectUri types.String `tfsdk:"redirect_uri"`
}

type connectionTenantModel struct {
	TenantDisplayName types.String `tfsdk:"tenant_display_name"`
	TenantIdentifier  types.String `tfsdk:"tenant_identifier"`
}
