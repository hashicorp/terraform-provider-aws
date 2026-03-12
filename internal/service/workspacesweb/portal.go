// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package workspacesweb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspacesweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workspacesweb_portal", name="Portal")
// @Tags(identifierAttribute="portal_arn")
// @Testing(tagsTest=true)
// @Testing(generator=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.Portal")
// @Testing(importStateIdAttribute="portal_arn")
func newPortalResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &portalResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type portalResource struct {
	framework.ResourceWithModel[portalResourceModel]
	framework.WithTimeouts
}

func (r *portalResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"additional_encryption_context": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"authentication_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AuthenticationType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"browser_settings_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"browser_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.BrowserType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreationDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"customer_managed_key": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_protection_settings_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDisplayName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrInstanceType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InstanceType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ip_access_settings_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"max_concurrent_sessions": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"network_settings_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_endpoint": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PortalStatus](),
				Computed:   true,
			},
			"renderer_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RendererType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"session_logger_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"trust_store_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_access_logging_settings_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_settings_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *portalResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data portalResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	var input workspacesweb.CreatePortalInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreatePortal(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating WorkSpacesWeb Portal", err.Error())
		return
	}

	data.PortalARN = fwflex.StringToFramework(ctx, output.PortalArn)
	data.PortalEndpoint = fwflex.StringToFramework(ctx, output.PortalEndpoint)

	// Wait for portal to be created
	portal, err := waitPortalCreated(ctx, conn, data.PortalARN.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for WorkSpacesWeb Portal (%s) create", data.PortalARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, portal, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *portalResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data portalResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	output, err := findPortalByARN(ctx, conn, data.PortalARN.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Portal (%s)", data.PortalARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *portalResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old portalResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	if !new.AuthenticationType.Equal(old.AuthenticationType) ||
		!new.BrowserSettingsARN.Equal(old.BrowserSettingsARN) ||
		!new.DataProtectionSettingsARN.Equal(old.DataProtectionSettingsARN) ||
		!new.DisplayName.Equal(old.DisplayName) ||
		!new.InstanceType.Equal(old.InstanceType) ||
		!new.IPAccessSettingsARN.Equal(old.IPAccessSettingsARN) ||
		!new.MaxConcurrentSessions.Equal(old.MaxConcurrentSessions) ||
		!new.NetworkSettingsARN.Equal(old.NetworkSettingsARN) ||
		!new.TrustStoreARN.Equal(old.TrustStoreARN) ||
		!new.UserAccessLoggingSettingsARN.Equal(old.UserAccessLoggingSettingsARN) ||
		!new.UserSettingsARN.Equal(old.UserSettingsARN) {
		var input workspacesweb.UpdatePortalInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdatePortal(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating WorkSpacesWeb Portal (%s)", new.PortalARN.ValueString()), err.Error())
			return
		}

		// Wait for portal to be updated
		portal, err := waitPortalUpdated(ctx, conn, new.PortalARN.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for WorkSpacesWeb Portal (%s) update", new.PortalARN.ValueString()), err.Error())
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, portal, &new)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *portalResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data portalResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DeletePortalInput{
		PortalArn: data.PortalARN.ValueStringPointer(),
	}
	_, err := conn.DeletePortal(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb Portal (%s)", data.PortalARN.ValueString()), err.Error())
		return
	}

	// Wait for portal to be deleted
	_, err = waitPortalDeleted(ctx, conn, data.PortalARN.ValueString(), r.DeleteTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for WorkSpacesWeb Portal (%s) delete", data.PortalARN.ValueString()), err.Error())
		return
	}
}

func (r *portalResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("portal_arn"), request, response)
}

// Waiters
func waitPortalCreated(ctx context.Context, conn *workspacesweb.Client, arn string, timeout time.Duration) (*awstypes.Portal, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PortalStatusPending),
		Target:                    enum.Slice(awstypes.PortalStatusIncomplete, awstypes.PortalStatusActive),
		Refresh:                   statusPortal(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Portal); ok {
		return out, err
	}

	return nil, err
}

func waitPortalUpdated(ctx context.Context, conn *workspacesweb.Client, arn string, timeout time.Duration) (*awstypes.Portal, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PortalStatusPending),
		Target:                    enum.Slice(awstypes.PortalStatusIncomplete, awstypes.PortalStatusActive),
		Refresh:                   statusPortal(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Portal); ok {
		return out, err
	}

	return nil, err
}

func waitPortalDeleted(ctx context.Context, conn *workspacesweb.Client, arn string, timeout time.Duration) (*awstypes.Portal, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PortalStatusActive, awstypes.PortalStatusIncomplete, awstypes.PortalStatusPending),
		Target:  []string{},
		Refresh: statusPortal(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Portal); ok {
		return out, err
	}

	return nil, err
}

// Status function
func statusPortal(conn *workspacesweb.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findPortalByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.PortalStatus), nil
	}
}

// Finder function
func findPortalByARN(ctx context.Context, conn *workspacesweb.Client, arn string) (*awstypes.Portal, error) {
	input := workspacesweb.GetPortalInput{
		PortalArn: aws.String(arn),
	}

	output, err := conn.GetPortal(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Portal == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Portal, nil
}

// Data model
type portalResourceModel struct {
	framework.WithRegionModel
	AdditionalEncryptionContext  fwtypes.MapOfString                             `tfsdk:"additional_encryption_context"`
	AuthenticationType           fwtypes.StringEnum[awstypes.AuthenticationType] `tfsdk:"authentication_type"`
	BrowserSettingsARN           fwtypes.ARN                                     `tfsdk:"browser_settings_arn"`
	BrowserType                  fwtypes.StringEnum[awstypes.BrowserType]        `tfsdk:"browser_type"`
	CreationDate                 timetypes.RFC3339                               `tfsdk:"creation_date"`
	CustomerManagedKey           types.String                                    `tfsdk:"customer_managed_key"`
	DataProtectionSettingsARN    types.String                                    `tfsdk:"data_protection_settings_arn"`
	DisplayName                  types.String                                    `tfsdk:"display_name"`
	InstanceType                 fwtypes.StringEnum[awstypes.InstanceType]       `tfsdk:"instance_type"`
	IPAccessSettingsARN          types.String                                    `tfsdk:"ip_access_settings_arn"`
	MaxConcurrentSessions        types.Int64                                     `tfsdk:"max_concurrent_sessions"`
	NetworkSettingsARN           types.String                                    `tfsdk:"network_settings_arn"`
	PortalARN                    types.String                                    `tfsdk:"portal_arn"`
	PortalEndpoint               types.String                                    `tfsdk:"portal_endpoint"`
	PortalStatus                 fwtypes.StringEnum[awstypes.PortalStatus]       `tfsdk:"portal_status"`
	RendererType                 fwtypes.StringEnum[awstypes.RendererType]       `tfsdk:"renderer_type"`
	SessionLoggerARN             types.String                                    `tfsdk:"session_logger_arn"`
	StatusReason                 types.String                                    `tfsdk:"status_reason"`
	Tags                         tftags.Map                                      `tfsdk:"tags"`
	TagsAll                      tftags.Map                                      `tfsdk:"tags_all"`
	Timeouts                     timeouts.Value                                  `tfsdk:"timeouts"`
	TrustStoreARN                types.String                                    `tfsdk:"trust_store_arn"`
	UserAccessLoggingSettingsARN types.String                                    `tfsdk:"user_access_logging_settings_arn"`
	UserSettingsARN              types.String                                    `tfsdk:"user_settings_arn"`
}
