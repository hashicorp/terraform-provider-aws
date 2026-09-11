// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehub

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehub/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_resiliencehub_app_version", name="App Version")
func newAppVersionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &appVersionResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameAppVersion = "App Version"

	appVersionResourceIDSeparator = ","
)

type appVersionResource struct {
	framework.ResourceWithModel[appVersionResourceModel]
	framework.WithTimeouts
}

func (r *appVersionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_template_body": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrIdentifier: schema.Int64Attribute{
				Computed: true,
			},
			"import_strategy": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ResourceImportStrategyType](),
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_arns": schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"version_name": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"terraform_source": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[terraformSourceModel](ctx),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"s3_state_file_url": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexache.MustCompile(`^https://[a-z0-9-]+\.s3\.[a-z0-9-]+\.amazonaws\.com/.+$|^s3://.+$`), "must be a valid S3 URL"),
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

func (r *appVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan appVersionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubClient(ctx)

	appARN := plan.AppARN.ValueString()
	timeout := r.CreateTimeout(ctx, plan.Timeouts)

	appVersion, identifier, err := putImportAndPublish(ctx, conn, &plan, timeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionCreating, ResNameAppVersion, appARN, err),
			err.Error(),
		)
		return
	}

	plan.AppVersion = types.StringValue(appVersion)
	plan.Identifier = types.Int64Value(identifier)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *appVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state appVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubClient(ctx)

	appARN := state.AppARN.ValueString()
	appVersion := state.AppVersion.ValueString()

	err := findAppVersionByTwoPartKey(ctx, conn, appARN, appVersion)
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionReading, ResNameAppVersion, appVersion, err),
			err.Error(),
		)
		return
	}

	// The template body, sources, and other inputs are create-time configuration
	// for an immutable published version. They are intentionally not refreshed
	// from the service, which augments the stored template (adding component IDs,
	// null fields, and normalizing the version), as doing so would produce a
	// permanent diff.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *appVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state appVersionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubClient(ctx)

	appARN := plan.AppARN.ValueString()
	timeout := r.UpdateTimeout(ctx, plan.Timeouts)

	// Any change to the template, sources, or version name results in a new
	// published application version.
	appVersion, identifier, err := putImportAndPublish(ctx, conn, &plan, timeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionUpdating, ResNameAppVersion, appARN, err),
			err.Error(),
		)
		return
	}

	plan.AppVersion = types.StringValue(appVersion)
	plan.Identifier = types.Int64Value(identifier)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *appVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Published application versions are immutable and there is no API to delete
	// an individual version; they are removed when the parent application is
	// deleted. Removing this resource only removes it from Terraform state.
	tflog.Warn(ctx, "Resilience Hub application versions cannot be deleted individually; removing from state only. The published version is removed when the parent application is deleted.")
}

func (r *appVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, appVersionResourceIDSeparator)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in the format <app_arn>%s<app_version>, got: %s", appVersionResourceIDSeparator, req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_arn"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_version"), parts[1])...)
}

// putImportAndPublish sets the draft template, imports any configured sources,
// waits for the import to complete, and publishes a new application version.
// It returns the published version and its numeric identifier.
func putImportAndPublish(ctx context.Context, conn *resiliencehub.Client, plan *appVersionResourceModel, timeout time.Duration) (string, int64, error) {
	appARN := plan.AppARN.ValueString()

	if _, err := conn.PutDraftAppVersionTemplate(ctx, &resiliencehub.PutDraftAppVersionTemplateInput{
		AppArn:          aws.String(appARN),
		AppTemplateBody: plan.AppTemplateBody.ValueStringPointer(),
	}); err != nil {
		return "", 0, fmt.Errorf("putting draft app version template: %w", err)
	}

	sourceARNs := make([]string, 0)
	if !plan.SourceARNs.IsNull() {
		if diags := plan.SourceARNs.ElementsAs(ctx, &sourceARNs, false); diags.HasError() {
			return "", 0, fmt.Errorf("expanding source_arns")
		}
	}
	var terraformSources []awstypes.TerraformSource
	if diags := flex.Expand(ctx, plan.TerraformSources, &terraformSources); diags.HasError() {
		return "", 0, fmt.Errorf("expanding terraform_source")
	}

	if len(sourceARNs) > 0 || len(terraformSources) > 0 {
		input := &resiliencehub.ImportResourcesToDraftAppVersionInput{
			AppArn:           aws.String(appARN),
			SourceArns:       sourceARNs,
			TerraformSources: terraformSources,
		}
		if !plan.ImportStrategy.IsNull() {
			input.ImportStrategy = awstypes.ResourceImportStrategyType(plan.ImportStrategy.ValueString())
		}

		if _, err := conn.ImportResourcesToDraftAppVersion(ctx, input); err != nil {
			return "", 0, fmt.Errorf("importing resources to draft app version: %w", err)
		}

		if err := waitDraftAppVersionResourcesImported(ctx, conn, appARN, timeout); err != nil {
			return "", 0, fmt.Errorf("waiting for draft app version resources import: %w", err)
		}
	}

	input := &resiliencehub.PublishAppVersionInput{
		AppArn: aws.String(appARN),
	}
	if !plan.VersionName.IsNull() {
		input.VersionName = plan.VersionName.ValueStringPointer()
	}

	output, err := conn.PublishAppVersion(ctx, input)
	if err != nil {
		return "", 0, fmt.Errorf("publishing app version: %w", err)
	}

	return aws.ToString(output.AppVersion), aws.ToInt64(output.Identifier), nil
}

func findAppVersionByTwoPartKey(ctx context.Context, conn *resiliencehub.Client, appARN, appVersion string) error {
	_, err := conn.DescribeAppVersion(ctx, &resiliencehub.DescribeAppVersionInput{
		AppArn:     aws.String(appARN),
		AppVersion: aws.String(appVersion),
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return &retry.NotFoundError{
			LastError: err,
		}
	}

	return err
}

func waitDraftAppVersionResourcesImported(ctx context.Context, conn *resiliencehub.Client, appARN string, timeout time.Duration) error {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceImportStatusTypePending, awstypes.ResourceImportStatusTypeInProgress),
		Target:  enum.Slice(awstypes.ResourceImportStatusTypeSuccess),
		Refresh: statusDraftAppVersionResourcesImport(ctx, conn, appARN),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func statusDraftAppVersionResourcesImport(ctx context.Context, conn *resiliencehub.Client, appARN string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := conn.DescribeDraftAppVersionResourcesImportStatus(ctx, &resiliencehub.DescribeDraftAppVersionResourcesImportStatusInput{
			AppArn: aws.String(appARN),
		})
		if err != nil {
			return nil, "", err
		}

		if output.Status == awstypes.ResourceImportStatusTypeFailed {
			return output, string(output.Status), fmt.Errorf("resource import failed: %s", aws.ToString(output.ErrorMessage))
		}

		return output, string(output.Status), nil
	}
}

type appVersionResourceModel struct {
	framework.WithRegionModel
	AppARN           fwtypes.ARN                                             `tfsdk:"app_arn"`
	AppTemplateBody  jsontypes.Normalized                                    `tfsdk:"app_template_body" autoflex:"-"`
	AppVersion       types.String                                            `tfsdk:"app_version"`
	Identifier       types.Int64                                             `tfsdk:"identifier"`
	ImportStrategy   fwtypes.StringEnum[awstypes.ResourceImportStrategyType] `tfsdk:"import_strategy"`
	SourceARNs       fwtypes.SetOfString                                     `tfsdk:"source_arns"`
	TerraformSources fwtypes.SetNestedObjectValueOf[terraformSourceModel]    `tfsdk:"terraform_source"`
	VersionName      types.String                                            `tfsdk:"version_name"`
	Timeouts         timeouts.Value                                          `tfsdk:"timeouts" autoflex:"-"`
}

type terraformSourceModel struct {
	S3StateFileURL types.String `tfsdk:"s3_state_file_url"`
}
