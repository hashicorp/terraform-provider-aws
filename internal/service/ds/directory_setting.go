// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_directory_service_directory_setting", name="Directory Setting")
// @IdentityAttribute("directory_id")
// @IdentityAttribute("name")
// @ImportIDHandler("directorySettingsImportID")
// @Testing(generator=false)
// @Testing(preCheck="acctest.PreCheckDirectoryService")
// @Testing(domainTfVar="domain")
// @Testing(hasNoPreExistingResource=true)
// @Testing(checkDestroyNoop=true)
// @Testing(importStateIdAttributes="directory_id;name", importStateIdAttributesSep="flex.ResourceIdSeparator")
func newDirectorySettingsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &directorySettingsResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

type directorySettingsResource struct {
	framework.ResourceWithModel[directorySettingsResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
	framework.WithNoOpDelete
}

func (r *directorySettingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"directory_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					directoryIDValidator,
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrValue: schema.StringAttribute{
				Required: true,
			},
			names.AttrType: schema.StringAttribute{
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
			}),
		},
	}
}

func (r *directorySettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DSClient(ctx)

	var plan directorySettingsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	directoryID := plan.DirectoryID.ValueString()

	input := directoryservice.UpdateSettingsInput{
		DirectoryId: aws.String(directoryID),
		Settings: []awstypes.Setting{
			{
				Name:  plan.Name.ValueStringPointer(),
				Value: plan.Value.ValueStringPointer(),
			},
		},
	}

	// The UpdateSettings retries and the waiter share one deadline.
	createDeadline := time.Now().Add(r.CreateTimeout(ctx, plan.Timeouts))

	_, err := tfresource.RetryWhenIsA[*directoryservice.UpdateSettingsOutput, *awstypes.DirectoryUnavailableException](ctx, time.Until(createDeadline), func(ctx context.Context) (*directoryservice.UpdateSettingsOutput, error) {
		return conn.UpdateSettings(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, directoryID)
		return
	}

	entry, err := waitDirectorySettingUpdated(ctx, conn, directoryID, plan.Name.ValueString(), time.Until(createDeadline))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, directoryID)
		return
	}

	r.flatten(ctx, entry, &plan)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *directorySettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DSClient(ctx)

	var state directorySettingsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	entry, err := findDirectorySettingByTwoPartKey(ctx, conn, state.DirectoryID.ValueString(), state.Name.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.DirectoryID.ValueString())
		return
	}

	r.flatten(ctx, entry, &state)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *directorySettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DSClient(ctx)

	var plan, state directorySettingsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		directoryID := plan.DirectoryID.ValueString()

		input := directoryservice.UpdateSettingsInput{
			DirectoryId: aws.String(directoryID),
			Settings: []awstypes.Setting{
				{
					Name:  plan.Name.ValueStringPointer(),
					Value: plan.Value.ValueStringPointer(),
				},
			},
		}

		// The UpdateSettings retries and the waiter share one deadline.
		updateDeadline := time.Now().Add(r.UpdateTimeout(ctx, plan.Timeouts))

		_, err := tfresource.RetryWhenIsA[*directoryservice.UpdateSettingsOutput, *awstypes.DirectoryUnavailableException](ctx, time.Until(updateDeadline), func(ctx context.Context) (*directoryservice.UpdateSettingsOutput, error) {
			return conn.UpdateSettings(ctx, &input)
		})
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, directoryID)
			return
		}

		entry, err := waitDirectorySettingUpdated(ctx, conn, directoryID, plan.Name.ValueString(), time.Until(updateDeadline))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, directoryID)
			return
		}

		r.flatten(ctx, entry, &plan)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func waitDirectorySettingUpdated(ctx context.Context, conn *directoryservice.Client, directoryID, settingName string, timeout time.Duration) (*awstypes.SettingEntry, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectoryConfigurationStatusRequested, awstypes.DirectoryConfigurationStatusUpdating),
		Target:  enum.Slice(awstypes.DirectoryConfigurationStatusUpdated, awstypes.DirectoryConfigurationStatusDefault),
		Refresh: statusDirectorySetting(conn, directoryID, settingName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if entry, ok := outputRaw.(*awstypes.SettingEntry); ok {
		return entry, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusDirectorySetting(conn *directoryservice.Client, directoryID, settingName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		entry, err := findDirectorySettingByTwoPartKey(ctx, conn, directoryID, settingName)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return entry, string(entry.RequestStatus), nil
	}
}

func findDirectorySettingsByDirectoryID(ctx context.Context, conn *directoryservice.Client, directoryID string) ([]awstypes.SettingEntry, error) {
	input := directoryservice.DescribeSettingsInput{
		DirectoryId: aws.String(directoryID),
	}

	var output []awstypes.SettingEntry
	err := describeSettingsPages(ctx, conn, &input, func(page *directoryservice.DescribeSettingsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.SettingEntries...)

		return !lastPage
	})

	if errs.IsA[*awstypes.EntityDoesNotExistException](err) || errs.IsA[*awstypes.DirectoryDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	return output, smarterr.NewError(err)
}

func findDirectorySettingByTwoPartKey(ctx context.Context, conn *directoryservice.Client, directoryID, settingName string) (*awstypes.SettingEntry, error) {
	return findDirectorySetting(ctx, conn, directoryID, func(e *awstypes.SettingEntry) bool {
		return aws.ToString(e.Name) == settingName
	})
}

func findDirectorySetting(ctx context.Context, conn *directoryservice.Client, directoryID string, filter tfslices.Predicate[*awstypes.SettingEntry]) (*awstypes.SettingEntry, error) {
	entries, err := findDirectorySettingsByDirectoryID(ctx, conn, directoryID)
	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(entries, tfslices.PredicateValue(filter)))
}

func (r *directorySettingsResource) flatten(ctx context.Context, entry *awstypes.SettingEntry, data *directorySettingsResourceModel) {
	data.Name = fwflex.StringToFramework(ctx, entry.Name)
	data.Value = fwflex.StringToFramework(ctx, entry.RequestedValue)
	data.Type = fwflex.StringToFramework(ctx, entry.Type)
}

var _ inttypes.ImportIDParser = directorySettingsImportID{}

type directorySettingsImportID struct{}

func (directorySettingsImportID) Parse(id string) (string, map[string]any, error) {
	directoryID, settingName, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found || directoryID == "" || settingName == "" {
		return "", nil, fmt.Errorf("id %q should be in the format <directory-id>%s<name>", id, intflex.ResourceIdSeparator)
	}

	return id, map[string]any{
		"directory_id": directoryID,
		names.AttrName: settingName,
	}, nil
}

type directorySettingsResourceModel struct {
	framework.WithRegionModel
	DirectoryID types.String   `tfsdk:"directory_id"`
	Name        types.String   `tfsdk:"name"`
	Value       types.String   `tfsdk:"value"`
	Type        types.String   `tfsdk:"type"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}
