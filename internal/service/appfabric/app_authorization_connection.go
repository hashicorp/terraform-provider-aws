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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="App Authorization Connection")
func newAppAuthorizationConnectionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &appAuthorizationConnectionResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)

	return r, nil
}

type appAuthorizationConnectionResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithNoOpDelete
	framework.WithTimeouts
}

func (*appAuthorizationConnectionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_appfabric_app_authorization_connection"
}

func (r *appAuthorizationConnectionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_authorization_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
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
			"tenant": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tenantModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[tenantModel](ctx),
				},
			},
			names.AttrID: framework.IDAttribute(),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *appAuthorizationConnectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data appAuthorizationConnectionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	input := &appfabric.ConnectAppAuthorizationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.AppBundleIdentifier = fwflex.StringFromFramework(ctx, data.AppBundleARN)
	input.AppAuthorizationIdentifier = fwflex.StringFromFramework(ctx, data.AppAuthorizationARN)

	_, err := conn.ConnectAppAuthorization(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating AppFabric App Authorization (%s) Connection", data.AppAuthorizationARN.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	appAuthorization, err := waitConnectAppAuthorizationCreated(ctx, conn, data.AppAuthorizationARN.ValueString(), data.AppBundleARN.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for AppFabric App Authorization Connection (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.App = fwflex.StringToFramework(ctx, appAuthorization.App)

	var tenant tenantModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, appAuthorization.Tenant, &tenant)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.Tenant = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tenant)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *appAuthorizationConnectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data appAuthorizationConnectionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	output, err := findAppAuthorizationConnectionByTwoPartKey(ctx, conn, data.AppAuthorizationARN.ValueString(), data.AppBundleARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading AppFabric App Authorization Connection (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findAppAuthorizationConnectionByTwoPartKey(ctx context.Context, conn *appfabric.Client, appAuthorizationARN, appBundleIdentifier string) (*awstypes.AppAuthorization, error) {
	input := &appfabric.GetAppAuthorizationInput{
		AppAuthorizationIdentifier: aws.String(appAuthorizationARN),
		AppBundleIdentifier:        aws.String(appBundleIdentifier),
	}

	output, err := conn.GetAppAuthorization(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AppAuthorization == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AppAuthorization, nil
}

func statusConnectAppAuthorization(ctx context.Context, conn *appfabric.Client, appAuthorizationARN, appBundleArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAppAuthorizationConnectionByTwoPartKey(ctx, conn, appAuthorizationARN, appBundleArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitConnectAppAuthorizationCreated(ctx context.Context, conn *appfabric.Client, appAuthorizationARN, appBundleArn string, timeout time.Duration) (*awstypes.AppAuthorization, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AppAuthorizationStatusPendingConnect),
		Target:  enum.Slice(awstypes.AppAuthorizationStatusConnected),
		Refresh: statusConnectAppAuthorization(ctx, conn, appAuthorizationARN, appBundleArn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*awstypes.AppAuthorization); ok {
		return out, err
	}

	return nil, err
}

type appAuthorizationConnectionResourceModel struct {
	App                 types.String                                      `tfsdk:"app"`
	AppAuthorizationARN fwtypes.ARN                                       `tfsdk:"app_authorization_arn"`
	AppBundleARN        fwtypes.ARN                                       `tfsdk:"app_bundle_arn"`
	AuthRequest         fwtypes.ListNestedObjectValueOf[authRequestModel] `tfsdk:"auth_request"`
	ID                  types.String                                      `tfsdk:"id"`
	Tenant              fwtypes.ListNestedObjectValueOf[tenantModel]      `tfsdk:"tenant"`
	Timeouts            timeouts.Value                                    `tfsdk:"timeouts"`
}

const (
	appAuthorizationConnectionResourceIDPartCount = 2
)

func (m *appAuthorizationConnectionResourceModel) InitFromID() error {
	parts, err := flex.ExpandResourceId(m.ID.ValueString(), appAuthorizationConnectionResourceIDPartCount, false)
	if err != nil {
		return err
	}

	m.AppAuthorizationARN = fwtypes.ARNValue(parts[0])
	m.AppBundleARN = fwtypes.ARNValue(parts[1])

	return nil
}

func (m *appAuthorizationConnectionResourceModel) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.AppAuthorizationARN.ValueString(), m.AppBundleARN.ValueString()}, appAuthorizationConnectionResourceIDPartCount, false)))
}

type authRequestModel struct {
	Code        types.String `tfsdk:"code"`
	RedirectURI types.String `tfsdk:"redirect_uri"`
}
