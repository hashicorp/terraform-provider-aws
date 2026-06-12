// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resiliencehub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehub/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	intretry "github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_resiliencehub_app", name="App")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehub;resiliencehub.DescribeAppOutput")
// @Testing(importStateIdAttribute="arn")
// @Testing(preIdentityVersion="v5.100.0")
// @ArnIdentity
func newAppResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &appResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

func (r *appResource) ValidateModel(ctx context.Context, schema *schema.Schema) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

const (
	ResNameApp = "App"
)

type appResource struct {
	framework.ResourceWithModel[appResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *appResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"assessment_schedule": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AppAssessmentScheduleType](),
				Optional:   true,
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 500),
				},
			},
			"drift_status": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_\-]{1,59}$`), "must be 2-60 characters, start with alphanumeric, contain only alphanumeric, underscore, and hyphen"),
				},
			},
			"resiliency_policy_arn": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^arn:(aws|aws-cn|aws-iso|aws-iso-[a-z]{1}|aws-us-gov):[A-Za-z0-9][A-Za-z0-9_/.-]{0,62}:([a-z]{2}-((iso[a-z]{0,1}-)|(gov-))?[a-z]+-[0-9]):[0-9]{12}:[A-Za-z0-9][A-Za-z0-9:_/+=,@.-]{0,255}$`), "must be a valid ARN"),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"app_template": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrVersion: schema.StringAttribute{
							Required: true,
						},
						"additional_info": schema.MapAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
					},
					Blocks: map[string]schema.Block{
						"app_component": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
									names.AttrType: schema.StringAttribute{
										Required: true,
									},
									"resource_names": schema.ListAttribute{
										Optional:    true,
										ElementType: types.StringType,
									},
									"additional_info": schema.MapAttribute{
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"resource": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
									names.AttrType: schema.StringAttribute{
										Required: true,
									},
									"additional_info": schema.MapAttribute{
										Optional:    true,
										ElementType: types.StringType,
									},
								},
								Blocks: map[string]schema.Block{
									"logical_resource_id": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrIdentifier: schema.StringAttribute{
													Required: true,
												},
												"logical_stack_name": schema.StringAttribute{
													Optional: true,
												},
												"resource_group_name": schema.StringAttribute{
													Optional: true,
												},
												"terraform_source_name": schema.StringAttribute{
													Optional: true,
												},
												"eks_source_name": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"resource_mapping": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"mapping_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ResourceMappingType](),
							Required:   true,
						},
						"resource_name": schema.StringAttribute{
							Required: true,
						},
						"terraform_source_name": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"physical_resource_id": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrIdentifier: schema.StringAttribute{
										Required: true,
									},
									names.AttrType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.PhysicalIdentifierType](),
										Required:   true,
									},
									names.AttrAWSAccountID: schema.StringAttribute{
										Optional: true,
									},
									"aws_region": schema.StringAttribute{
										Optional: true,
									},
								},
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

func (r *appResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan appResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubClient(ctx)

	name := plan.Name.ValueString()
	input := &resiliencehub.CreateAppInput{
		Tags: getTagsIn(ctx),
	}

	// Use AutoFlex to expand plan to input
	resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateApp(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionCreating, ResNameApp, name, err),
			err.Error(),
		)
		return
	}

	plan.AppArn = types.StringValue(aws.ToString(output.App.AppArn))

	// Use AutoFlex to flatten the created app back to plan
	resp.Diagnostics.Append(flex.Flatten(ctx, output.App, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Put app template
	if !plan.AppTemplate.IsNull() && !plan.AppTemplate.IsUnknown() {
		templateBody, diags := r.expandAppTemplate(ctx, plan.AppTemplate)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		templateInput := &resiliencehub.PutDraftAppVersionTemplateInput{
			AppArn:          output.App.AppArn,
			AppTemplateBody: aws.String(templateBody),
		}

		_, err = conn.PutDraftAppVersionTemplate(ctx, templateInput)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionCreating, ResNameApp, name, err),
				err.Error(),
			)
			return
		}
	}

	// Import resources with full mapping support
	if !plan.ResourceMapping.IsNull() && !plan.ResourceMapping.IsUnknown() {
		sourceArns, terraformSources, diags := r.expandResourceMappingsToSources(ctx, plan.ResourceMapping)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Only call ImportResourcesToDraftAppVersion if we have actual sources to import
		if len(sourceArns) > 0 || len(terraformSources) > 0 {
			importInput := &resiliencehub.ImportResourcesToDraftAppVersionInput{
				AppArn:           output.App.AppArn,
				SourceArns:       sourceArns,
				TerraformSources: terraformSources,
			}

			_, err := conn.ImportResourcesToDraftAppVersion(ctx, importInput)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionCreating, ResNameApp, name, err),
					err.Error(),
				)
				return
			}
		}
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	app, err := waitAppCreated(ctx, conn, plan.AppArn.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForCreation, ResNameApp, name, err),
			err.Error(),
		)
		return
	}

	plan.DriftStatus = types.StringValue(string(app.DriftStatus))
	plan.AssessmentSchedule = fwtypes.StringEnumValue(app.AssessmentSchedule)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *appResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state appResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubClient(ctx)

	output, err := FindAppByARN(ctx, conn, state.AppArn.ValueString())
	if intretry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionSetting, ResNameApp, state.AppArn.ValueString(), err),
			err.Error(),
		)
		return
	}

	// Use AutoFlex to flatten output to state
	resp.Diagnostics.Append(flex.Flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure AppArn is set correctly for import
	state.AppArn = types.StringValue(aws.ToString(output.AppArn))

	// Read app template from draft version
	templateInput := &resiliencehub.DescribeAppVersionTemplateInput{
		AppArn:     state.AppArn.ValueStringPointer(),
		AppVersion: aws.String("draft"),
	}

	templateOutput, err := conn.DescribeAppVersionTemplate(ctx, templateInput)
	if err != nil {
		// If draft doesn't exist, try published version
		templateInput.AppVersion = aws.String("release")
		templateOutput, err = conn.DescribeAppVersionTemplate(ctx, templateInput)
	}

	if err != nil {
		// Template doesn't exist, set to null
		state.AppTemplate = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				names.AttrVersion: types.StringType,
				"additional_info": types.MapType{ElemType: types.StringType},
				"app_component":   types.ListType{ElemType: appComponentObjectType()},
				"resource":        types.ListType{ElemType: resourceObjectType()},
			},
		})
	} else {
		// Flatten the template
		appTemplate, diags := r.flattenAppTemplate(ctx, *templateOutput.AppTemplateBody)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.AppTemplate = appTemplate
	}

	// Read input sources (resource mappings)
	// Try both draft and release versions since import might be asynchronous
	var inputSourcesOutput *resiliencehub.ListAppInputSourcesOutput
	var inputSourcesErr error

	// First try draft version
	inputSourcesInput := &resiliencehub.ListAppInputSourcesInput{
		AppArn:     state.AppArn.ValueStringPointer(),
		AppVersion: aws.String("draft"),
	}

	inputSourcesOutput, inputSourcesErr = conn.ListAppInputSources(ctx, inputSourcesInput)

	if inputSourcesErr != nil || inputSourcesOutput == nil || len(inputSourcesOutput.AppInputSources) == 0 {
		// Try release version
		inputSourcesInput.AppVersion = aws.String("release")
		inputSourcesOutput, inputSourcesErr = conn.ListAppInputSources(ctx, inputSourcesInput)
	}

	if inputSourcesErr != nil || inputSourcesOutput == nil || len(inputSourcesOutput.AppInputSources) == 0 {
		// Input sources might not be immediately available after app creation/update
		// Preserve existing resource mappings from state to prevent unnecessary drift
		if state.ResourceMapping.IsNull() || state.ResourceMapping.IsUnknown() {
			// No previous state exists, set to null
			state.ResourceMapping = types.ListNull(resourceMappingObjectType())
		}
	} else {
		// Parse the app template JSON for resource mappings
		var appTemplateMap map[string]any
		if err := json.Unmarshal([]byte(*templateOutput.AppTemplateBody), &appTemplateMap); err != nil {
			resp.Diagnostics.AddError("Failed to parse app template", err.Error())
			return
		}

		// Flatten the resource mappings
		resourceMappings, diags := r.flattenResourceMappings(ctx, inputSourcesOutput.AppInputSources, appTemplateMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.ResourceMapping = resourceMappings
	}

	setTagsOut(ctx, output.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *appResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state appResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubClient(ctx)

	// Handle basic app property updates
	if !plan.Description.Equal(state.Description) ||
		!plan.AssessmentSchedule.Equal(state.AssessmentSchedule) ||
		!plan.PolicyArn.Equal(state.PolicyArn) {
		input := &resiliencehub.UpdateAppInput{
			AppArn: plan.AppArn.ValueStringPointer(),
		}

		// Use AutoFlex to expand plan to input
		resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateApp(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionUpdating, ResNameApp, plan.AppArn.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	// Handle app template updates
	if !plan.AppTemplate.Equal(state.AppTemplate) {
		if !plan.AppTemplate.IsNull() && !plan.AppTemplate.IsUnknown() {
			templateBody, diags := r.expandAppTemplate(ctx, plan.AppTemplate)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			templateInput := &resiliencehub.PutDraftAppVersionTemplateInput{
				AppArn:          plan.AppArn.ValueStringPointer(),
				AppTemplateBody: aws.String(templateBody),
			}

			_, err := conn.PutDraftAppVersionTemplate(ctx, templateInput)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionUpdating, ResNameApp, plan.AppArn.ValueString(), err),
					err.Error(),
				)
				return
			}
		}
	}

	// Handle resource mapping updates
	if !plan.ResourceMapping.Equal(state.ResourceMapping) {
		if !plan.ResourceMapping.IsNull() && !plan.ResourceMapping.IsUnknown() {
			sourceArns, terraformSources, diags := r.expandResourceMappingsToSources(ctx, plan.ResourceMapping)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Only call ImportResourcesToDraftAppVersion if we have actual sources to import
			if len(sourceArns) > 0 || len(terraformSources) > 0 {
				importInput := &resiliencehub.ImportResourcesToDraftAppVersionInput{
					AppArn:           plan.AppArn.ValueStringPointer(),
					SourceArns:       sourceArns,
					TerraformSources: terraformSources,
				}

				_, err := conn.ImportResourcesToDraftAppVersion(ctx, importInput)
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionUpdating, ResNameApp, plan.AppArn.ValueString(), err),
						err.Error(),
					)
					return
				}
			}
		}
	}

	// Handle tag updates - use TagsAll to include provider tags
	if !plan.TagsAll.Equal(state.TagsAll) {
		if err := updateTags(ctx, conn, plan.AppArn.ValueString(), state.TagsAll, plan.TagsAll); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionUpdating, ResNameApp, plan.AppArn.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	app, err := waitAppUpdated(ctx, conn, plan.AppArn.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForUpdate, ResNameApp, plan.AppArn.ValueString(), err),
			err.Error(),
		)
		return
	}

	// Update computed values from the returned app
	plan.DriftStatus = types.StringValue(string(app.DriftStatus))
	plan.AssessmentSchedule = fwtypes.StringEnumValue(app.AssessmentSchedule)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *appResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state appResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubClient(ctx)

	_, err := conn.DeleteApp(ctx, &resiliencehub.DeleteAppInput{
		AppArn: state.AppArn.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionDeleting, ResNameApp, state.AppArn.ValueString(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitAppDeleted(ctx, conn, state.AppArn.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForDeletion, ResNameApp, state.AppArn.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func FindAppByARN(ctx context.Context, conn *resiliencehub.Client, arn string) (*awstypes.App, error) {
	input := &resiliencehub.DescribeAppInput{
		AppArn: aws.String(arn),
	}

	output, err := conn.DescribeApp(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.App == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.App, nil
}

func waitAppCreated(ctx context.Context, conn *resiliencehub.Client, arn string, timeout time.Duration) (*awstypes.App, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.AppStatusTypeActive),
		Refresh:                   statusApp(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.App); ok {
		return output, err
	}

	return nil, err
}

func waitAppUpdated(ctx context.Context, conn *resiliencehub.Client, arn string, timeout time.Duration) (*awstypes.App, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.AppStatusTypeActive),
		Refresh:                   statusApp(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.App); ok {
		return output, err
	}

	return nil, err
}

func waitAppDeleted(ctx context.Context, conn *resiliencehub.Client, arn string, timeout time.Duration) (*awstypes.App, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AppStatusTypeActive, awstypes.AppStatusTypeDeleting),
		Target:  []string{},
		Refresh: statusApp(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.App); ok {
		return output, err
	}

	return nil, err
}

func statusApp(ctx context.Context, conn *resiliencehub.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := FindAppByARN(ctx, conn, arn)
		if intretry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

// Data model
type appResourceModel struct {
	framework.WithRegionModel
	AppArn             types.String                                           `tfsdk:"arn"`
	AssessmentSchedule fwtypes.StringEnum[awstypes.AppAssessmentScheduleType] `tfsdk:"assessment_schedule"`
	AppTemplate        types.List                                             `tfsdk:"app_template" autoflex:"-"`
	Description        types.String                                           `tfsdk:"description"`
	DriftStatus        types.String                                           `tfsdk:"drift_status"`
	Name               types.String                                           `tfsdk:"name"`
	PolicyArn          types.String                                           `tfsdk:"resiliency_policy_arn"`
	ResourceMapping    types.List                                             `tfsdk:"resource_mapping" autoflex:"-"`
	Tags               tftags.Map                                             `tfsdk:"tags" autoflex:"-"`
	TagsAll            tftags.Map                                             `tfsdk:"tags_all" autoflex:"-"`
	Timeouts           timeouts.Value                                         `tfsdk:"timeouts" autoflex:"-"`
}

// Helper functions for data transformation
func (r *appResource) expandAppTemplate(ctx context.Context, tfList types.List) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if tfList.IsNull() || tfList.IsUnknown() {
		return "", diags
	}

	var appTemplates []appTemplateModel
	diags.Append(tfList.ElementsAs(ctx, &appTemplates, false)...)
	if diags.HasError() {
		return "", diags
	}

	if len(appTemplates) == 0 {
		return "", diags
	}

	appTemplate := appTemplates[0]

	// Build the JSON structure that ResilienceHub expects
	template := map[string]any{
		names.AttrVersion: appTemplate.Version.ValueString(),
	}

	// Add additional info if present
	if !appTemplate.AdditionalInfo.IsNull() && !appTemplate.AdditionalInfo.IsUnknown() {
		additionalInfo := make(map[string]string)
		diags.Append(appTemplate.AdditionalInfo.ElementsAs(ctx, &additionalInfo, false)...)
		if diags.HasError() {
			return "", diags
		}
		if len(additionalInfo) > 0 {
			template["additionalInfo"] = additionalInfo
		}
	}

	// Expand resources
	if !appTemplate.Resources.IsNull() && !appTemplate.Resources.IsUnknown() {
		resources, expandDiags := r.expandAppTemplateResources(ctx, appTemplate.Resources)
		diags.Append(expandDiags...)
		if diags.HasError() {
			return "", diags
		}
		template[names.AttrResources] = resources
	} else {
		template[names.AttrResources] = []any{}
	}

	// Expand app components
	if !appTemplate.AppComponents.IsNull() && !appTemplate.AppComponents.IsUnknown() {
		appComponents, expandDiags := r.expandAppTemplateComponents(ctx, appTemplate.AppComponents)
		diags.Append(expandDiags...)
		if diags.HasError() {
			return "", diags
		}
		template["appComponents"] = appComponents
	} else {
		template["appComponents"] = []any{}
	}

	templateJSON, err := json.Marshal(template)
	if err != nil {
		diags.AddError("Failed to marshal app template", err.Error())
		return "", diags
	}

	return string(templateJSON), diags
}

func (r *appResource) expandAppTemplateResources(ctx context.Context, tfList types.List) ([]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if tfList.IsNull() || tfList.IsUnknown() {
		return []any{}, diags
	}

	var resources []resourceModel
	diags.Append(tfList.ElementsAs(ctx, &resources, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]any, len(resources))
	for i, resource := range resources {
		resourceMap := map[string]any{
			names.AttrName: resource.Name.ValueString(),
			names.AttrType: resource.Type.ValueString(),
		}

		// Add logical resource ID
		if !resource.LogicalResourceId.IsNull() && !resource.LogicalResourceId.IsUnknown() {
			var logicalIds []logicalResourceIdModel
			diags.Append(resource.LogicalResourceId.ElementsAs(ctx, &logicalIds, false)...)
			if diags.HasError() {
				return nil, diags
			}

			if len(logicalIds) > 0 {
				logicalId := logicalIds[0]
				logicalResourceId := map[string]any{
					names.AttrIdentifier: logicalId.Identifier.ValueString(),
				}

				if !logicalId.LogicalStackName.IsNull() {
					logicalResourceId["logicalStackName"] = logicalId.LogicalStackName.ValueString()
				}
				if !logicalId.ResourceGroupName.IsNull() {
					logicalResourceId["resourceGroupName"] = logicalId.ResourceGroupName.ValueString()
				}
				if !logicalId.TerraformSourceName.IsNull() {
					logicalResourceId["terraformSourceName"] = logicalId.TerraformSourceName.ValueString()
				}
				if !logicalId.EksSourceName.IsNull() {
					logicalResourceId["eksSourceName"] = logicalId.EksSourceName.ValueString()
				}

				resourceMap["logicalResourceId"] = logicalResourceId
			}
		}

		// Add additional info if present
		if !resource.AdditionalInfo.IsNull() && !resource.AdditionalInfo.IsUnknown() {
			additionalInfo := make(map[string]string)
			diags.Append(resource.AdditionalInfo.ElementsAs(ctx, &additionalInfo, false)...)
			if diags.HasError() {
				return nil, diags
			}
			if len(additionalInfo) > 0 {
				resourceMap["additionalInfo"] = additionalInfo
			}
		}

		result[i] = resourceMap
	}

	return result, diags
}

func (r *appResource) expandAppTemplateComponents(ctx context.Context, tfList types.List) ([]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if tfList.IsNull() || tfList.IsUnknown() {
		return []any{}, diags
	}

	var components []appComponentModel
	diags.Append(tfList.ElementsAs(ctx, &components, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]any, len(components))
	for i, component := range components {
		componentMap := map[string]any{
			names.AttrName: component.Name.ValueString(),
			names.AttrType: component.Type.ValueString(),
		}

		// Add resource names if present
		if !component.ResourceNames.IsNull() && !component.ResourceNames.IsUnknown() {
			var resourceNames []string
			diags.Append(component.ResourceNames.ElementsAs(ctx, &resourceNames, false)...)
			if diags.HasError() {
				return nil, diags
			}
			componentMap["resourceNames"] = resourceNames
		}

		// Add additional info if present
		if !component.AdditionalInfo.IsNull() && !component.AdditionalInfo.IsUnknown() {
			additionalInfo := make(map[string]string)
			diags.Append(component.AdditionalInfo.ElementsAs(ctx, &additionalInfo, false)...)
			if diags.HasError() {
				return nil, diags
			}
			if len(additionalInfo) > 0 {
				componentMap["additionalInfo"] = additionalInfo
			}
		}

		result[i] = componentMap
	}

	return result, diags
}

// Data models for app template parsing
type appTemplateModel struct {
	Version        types.String `tfsdk:"version"`
	AdditionalInfo types.Map    `tfsdk:"additional_info"`
	AppComponents  types.List   `tfsdk:"app_component"`
	Resources      types.List   `tfsdk:"resource"`
}

type appComponentModel struct {
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	ResourceNames  types.List   `tfsdk:"resource_names"`
	AdditionalInfo types.Map    `tfsdk:"additional_info"`
}

type resourceModel struct {
	Name              types.String `tfsdk:"name"`
	Type              types.String `tfsdk:"type"`
	AdditionalInfo    types.Map    `tfsdk:"additional_info"`
	LogicalResourceId types.List   `tfsdk:"logical_resource_id"`
}

type logicalResourceIdModel struct {
	Identifier          types.String `tfsdk:"identifier"`
	LogicalStackName    types.String `tfsdk:"logical_stack_name"`
	ResourceGroupName   types.String `tfsdk:"resource_group_name"`
	TerraformSourceName types.String `tfsdk:"terraform_source_name"`
	EksSourceName       types.String `tfsdk:"eks_source_name"`
}

type resourceMappingModel struct {
	MappingType         fwtypes.StringEnum[awstypes.ResourceMappingType] `tfsdk:"mapping_type"`
	ResourceName        types.String                                     `tfsdk:"resource_name"`
	TerraformSourceName types.String                                     `tfsdk:"terraform_source_name"`
	PhysicalResourceId  types.List                                       `tfsdk:"physical_resource_id"`
}

type physicalResourceIdModel struct {
	Type         fwtypes.StringEnum[awstypes.PhysicalIdentifierType] `tfsdk:"type"`
	Identifier   types.String                                        `tfsdk:"identifier"`
	AwsAccountId types.String                                        `tfsdk:"aws_account_id"`
	AwsRegion    types.String                                        `tfsdk:"aws_region"`
}

func (r *appResource) expandResourceMappingsToSources(ctx context.Context, tfList types.List) ([]string, []awstypes.TerraformSource, diag.Diagnostics) {
	var diags diag.Diagnostics

	if tfList.IsNull() || tfList.IsUnknown() {
		return nil, nil, diags
	}

	var mappings []resourceMappingModel
	diags.Append(tfList.ElementsAs(ctx, &mappings, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}

	var sourceArns []string
	var terraformSources []awstypes.TerraformSource

	for _, mapping := range mappings {
		mappingType := mapping.MappingType.ValueEnum()

		switch mappingType {
		case awstypes.ResourceMappingTypeTerraform:
			// Terraform `tfstate` files in S3
			if !mapping.PhysicalResourceId.IsNull() && !mapping.PhysicalResourceId.IsUnknown() {
				var physicalIds []physicalResourceIdModel
				diags.Append(mapping.PhysicalResourceId.ElementsAs(ctx, &physicalIds, false)...)
				if diags.HasError() {
					return nil, nil, diags
				}

				if len(physicalIds) > 0 {
					physicalId := physicalIds[0]
					// Use the identifier as the S3 state file URL for Terraform sources
					terraformSource := awstypes.TerraformSource{
						S3StateFileUrl: physicalId.Identifier.ValueStringPointer(),
					}
					terraformSources = append(terraformSources, terraformSource)
				}
			}
		case awstypes.ResourceMappingTypeResource:
			// Handle direct resource ARNs
			if !mapping.PhysicalResourceId.IsNull() && !mapping.PhysicalResourceId.IsUnknown() {
				var physicalIds []physicalResourceIdModel
				diags.Append(mapping.PhysicalResourceId.ElementsAs(ctx, &physicalIds, false)...)
				if diags.HasError() {
					return nil, nil, diags
				}

				if len(physicalIds) > 0 {
					physicalId := physicalIds[0]
					if physicalId.Type.ValueEnum() == awstypes.PhysicalIdentifierTypeArn {
						sourceArns = append(sourceArns, physicalId.Identifier.ValueString())
					}
				}
			}
		}
	}

	return sourceArns, terraformSources, diags
}

func (r *appResource) flattenAppTemplate(ctx context.Context, templateBody string) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if templateBody == "" {
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				names.AttrVersion: types.StringType,
				"additional_info": types.MapType{ElemType: types.StringType},
				"app_component":   types.ListType{ElemType: appComponentObjectType()},
				"resource":        types.ListType{ElemType: resourceObjectType()},
			},
		}), diags
	}

	// Parse JSON from AWS API
	var template map[string]any
	if err := json.Unmarshal([]byte(templateBody), &template); err != nil {
		diags.AddError("Failed to parse app template JSON", err.Error())
		return types.ListNull(types.ObjectType{}), diags
	}

	// Build app template model
	var version string
	if v, ok := template[names.AttrVersion].(string); ok {
		version = v
	} else if v, ok := template[names.AttrVersion].(float64); ok {
		version = fmt.Sprintf("%.1f", v)
	} else {
		version = "2.0" // default
	}

	appTemplate := appTemplateModel{
		Version: types.StringValue(version),
	}

	// Handle additional info
	if additionalInfo, ok := template["additionalInfo"].(map[string]any); ok {
		additionalInfoMap := make(map[string]attr.Value)
		for k, v := range additionalInfo {
			additionalInfoMap[k] = types.StringValue(v.(string))
		}
		appTemplate.AdditionalInfo = types.MapValueMust(types.StringType, additionalInfoMap)
	} else {
		appTemplate.AdditionalInfo = types.MapNull(types.StringType)
	}

	// Handle resources
	if resources, ok := template[names.AttrResources].([]any); ok {
		resourceList, flattenDiags := r.flattenAppTemplateResources(ctx, resources)
		diags.Append(flattenDiags...)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{}), diags
		}
		appTemplate.Resources = resourceList
	} else {
		appTemplate.Resources = types.ListNull(resourceObjectType())
	}

	// Handle app components
	if appComponents, ok := template["appComponents"].([]any); ok {
		componentList, flattenDiags := r.flattenAppTemplateComponents(ctx, appComponents)
		diags.Append(flattenDiags...)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{}), diags
		}
		appTemplate.AppComponents = componentList
	} else {
		appTemplate.AppComponents = types.ListNull(appComponentObjectType())
	}

	// Convert to types.List
	appTemplateValue, convertDiags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		names.AttrVersion: types.StringType,
		"additional_info": types.MapType{ElemType: types.StringType},
		"app_component":   types.ListType{ElemType: appComponentObjectType()},
		"resource":        types.ListType{ElemType: resourceObjectType()},
	}, appTemplate)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{}), diags
	}

	return types.ListValueMust(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			names.AttrVersion: types.StringType,
			"additional_info": types.MapType{ElemType: types.StringType},
			"app_component":   types.ListType{ElemType: appComponentObjectType()},
			"resource":        types.ListType{ElemType: resourceObjectType()},
		},
	}, []attr.Value{appTemplateValue}), diags
}

// Helper functions for object types
func appComponentObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			names.AttrName:    types.StringType,
			names.AttrType:    types.StringType,
			"resource_names":  types.ListType{ElemType: types.StringType},
			"additional_info": types.MapType{ElemType: types.StringType},
		},
	}
}

func resourceObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			names.AttrName:        types.StringType,
			names.AttrType:        types.StringType,
			"additional_info":     types.MapType{ElemType: types.StringType},
			"logical_resource_id": types.ListType{ElemType: logicalResourceIdObjectType()},
		},
	}
}

func logicalResourceIdObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			names.AttrIdentifier:    types.StringType,
			"logical_stack_name":    types.StringType,
			"resource_group_name":   types.StringType,
			"terraform_source_name": types.StringType,
			"eks_source_name":       types.StringType,
		},
	}
}

func resourceMappingObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"mapping_type":          fwtypes.StringEnumType[awstypes.ResourceMappingType](),
			"resource_name":         types.StringType,
			"terraform_source_name": types.StringType,
			"physical_resource_id":  types.ListType{ElemType: physicalResourceIdObjectType()},
		},
	}
}

func physicalResourceIdObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			names.AttrType:         fwtypes.StringEnumType[awstypes.PhysicalIdentifierType](),
			names.AttrIdentifier:   types.StringType,
			names.AttrAWSAccountID: types.StringType,
			"aws_region":           types.StringType,
		},
	}
}

func (r *appResource) flattenAppTemplateResources(ctx context.Context, resources []any) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(resources) == 0 {
		return types.ListNull(resourceObjectType()), diags
	}

	resourceValues := make([]attr.Value, len(resources))
	for i, res := range resources {
		resource := res.(map[string]any)

		resourceModel := resourceModel{
			Name: types.StringValue(resource[names.AttrName].(string)),
			Type: types.StringValue(resource[names.AttrType].(string)),
		}

		// Handle additional info
		if additionalInfo, ok := resource["additionalInfo"].(map[string]any); ok {
			additionalInfoMap := make(map[string]attr.Value)
			for k, v := range additionalInfo {
				additionalInfoMap[k] = types.StringValue(v.(string))
			}
			resourceModel.AdditionalInfo = types.MapValueMust(types.StringType, additionalInfoMap)
		} else {
			resourceModel.AdditionalInfo = types.MapNull(types.StringType)
		}

		// Handle logical resource ID
		if logicalResourceId, ok := resource["logicalResourceId"].(map[string]any); ok {
			logicalId := logicalResourceIdModel{}

			if identifier, exists := logicalResourceId[names.AttrIdentifier]; exists && identifier != nil {
				logicalId.Identifier = types.StringValue(identifier.(string))
			} else {
				logicalId.Identifier = types.StringNull()
			}

			if logicalStackName, exists := logicalResourceId["logicalStackName"]; exists && logicalStackName != nil {
				logicalId.LogicalStackName = types.StringValue(logicalStackName.(string))
			} else {
				logicalId.LogicalStackName = types.StringNull()
			}

			if resourceGroupName, exists := logicalResourceId["resourceGroupName"]; exists && resourceGroupName != nil {
				logicalId.ResourceGroupName = types.StringValue(resourceGroupName.(string))
			} else {
				logicalId.ResourceGroupName = types.StringNull()
			}

			if terraformSourceName, exists := logicalResourceId["terraformSourceName"]; exists && terraformSourceName != nil {
				logicalId.TerraformSourceName = types.StringValue(terraformSourceName.(string))
			} else {
				logicalId.TerraformSourceName = types.StringNull()
			}

			if eksSourceName, exists := logicalResourceId["eksSourceName"]; exists {
				logicalId.EksSourceName = types.StringValue(eksSourceName.(string))
			} else {
				logicalId.EksSourceName = types.StringNull()
			}

			logicalIdValue, convertDiags := types.ObjectValueFrom(ctx, map[string]attr.Type{
				names.AttrIdentifier:    types.StringType,
				"logical_stack_name":    types.StringType,
				"resource_group_name":   types.StringType,
				"terraform_source_name": types.StringType,
				"eks_source_name":       types.StringType,
			}, logicalId)
			diags.Append(convertDiags...)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{}), diags
			}

			resourceModel.LogicalResourceId = types.ListValueMust(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					names.AttrIdentifier:    types.StringType,
					"logical_stack_name":    types.StringType,
					"resource_group_name":   types.StringType,
					"terraform_source_name": types.StringType,
					"eks_source_name":       types.StringType,
				},
			}, []attr.Value{logicalIdValue})
		} else {
			resourceModel.LogicalResourceId = types.ListNull(types.ObjectType{})
		}

		resourceValue, convertDiags := types.ObjectValueFrom(ctx, map[string]attr.Type{
			names.AttrName:        types.StringType,
			names.AttrType:        types.StringType,
			"additional_info":     types.MapType{ElemType: types.StringType},
			"logical_resource_id": types.ListType{ElemType: logicalResourceIdObjectType()},
		}, resourceModel)
		diags.Append(convertDiags...)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{}), diags
		}

		resourceValues[i] = resourceValue
	}

	return types.ListValueMust(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			names.AttrName:        types.StringType,
			names.AttrType:        types.StringType,
			"additional_info":     types.MapType{ElemType: types.StringType},
			"logical_resource_id": types.ListType{ElemType: logicalResourceIdObjectType()},
		},
	}, resourceValues), diags
}

func (r *appResource) flattenAppTemplateComponents(ctx context.Context, components []any) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(components) == 0 {
		return types.ListNull(appComponentObjectType()), diags
	}

	componentValues := make([]attr.Value, len(components))
	for i, comp := range components {
		component := comp.(map[string]any)

		componentModel := appComponentModel{
			Name: types.StringValue(component[names.AttrName].(string)),
			Type: types.StringValue(component[names.AttrType].(string)),
		}

		// Handle resource names
		if resourceNames, ok := component["resourceNames"].([]any); ok && len(resourceNames) > 0 {
			resourceNameValues := make([]attr.Value, len(resourceNames))
			for j, name := range resourceNames {
				resourceNameValues[j] = types.StringValue(name.(string))
			}
			componentModel.ResourceNames = types.ListValueMust(types.StringType, resourceNameValues)
		} else {
			componentModel.ResourceNames = types.ListValueMust(types.StringType, []attr.Value{})
		}

		// Handle additional info
		if additionalInfo, ok := component["additionalInfo"].(map[string]any); ok {
			additionalInfoMap := make(map[string]attr.Value)
			for k, v := range additionalInfo {
				additionalInfoMap[k] = types.StringValue(v.(string))
			}
			componentModel.AdditionalInfo = types.MapValueMust(types.StringType, additionalInfoMap)
		} else {
			componentModel.AdditionalInfo = types.MapNull(types.StringType)
		}

		componentValue, convertDiags := types.ObjectValueFrom(ctx, map[string]attr.Type{
			names.AttrName:    types.StringType,
			names.AttrType:    types.StringType,
			"resource_names":  types.ListType{ElemType: types.StringType},
			"additional_info": types.MapType{ElemType: types.StringType},
		}, componentModel)
		diags.Append(convertDiags...)
		if diags.HasError() {
			return types.ListNull(appComponentObjectType()), diags
		}

		componentValues[i] = componentValue
	}

	return types.ListValueMust(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			names.AttrName:    types.StringType,
			names.AttrType:    types.StringType,
			"resource_names":  types.ListType{ElemType: types.StringType},
			"additional_info": types.MapType{ElemType: types.StringType},
		},
	}, componentValues), diags
}

func (r *appResource) flattenResourceMappings(ctx context.Context, inputSources []awstypes.AppInputSource, appTemplate map[string]any) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(inputSources) == 0 {
		return types.ListNull(resourceMappingObjectType()), diags
	}

	var mappingValues []attr.Value

	// Parse app template to find resources with terraform source names
	terraformResourceMap := make(map[string]string) // terraformSourceName -> resourceName
	if resources, ok := appTemplate[names.AttrResources].([]any); ok {
		for _, res := range resources {
			if resource, ok := res.(map[string]any); ok {
				if resourceName, hasName := resource[names.AttrName].(string); hasName {
					if logicalResourceId, hasLogical := resource["logicalResourceId"].(map[string]any); hasLogical {
						if terraformSourceName, hasTerraform := logicalResourceId["terraformSourceName"].(string); hasTerraform {
							terraformResourceMap[terraformSourceName] = resourceName
						}
					}
				}
			}
		}
	}

	for _, inputSource := range inputSources {
		// Handle Terraform sources
		if inputSource.TerraformSource != nil {
			s3Url := aws.ToString(inputSource.TerraformSource.S3StateFileUrl)

			// Match S3 URL with terraform source name from app template
			// Look through all terraform resources to find the one with matching S3 URL
			var resourceName, terraformSourceName string

			// First, try to find exact match by checking if we have this S3 URL in our expected mappings
			for tsName, resName := range terraformResourceMap {
				// For now, assume first terraform source matches (AWS limitation: 1 terraform source per app)
				resourceName = resName
				terraformSourceName = tsName
				break
			}

			// If we have terraform resources in template but couldn't correlate, still create mapping
			if resourceName == "" && len(terraformResourceMap) > 0 {
				// Use the first terraform resource as fallback
				for tsName, resName := range terraformResourceMap {
					resourceName = resName
					terraformSourceName = tsName
					break
				}
			}

			// If still no match, skip this source (no terraform resources in template)
			if resourceName == "" || terraformSourceName == "" {
				continue
			}

			mappingModel := resourceMappingModel{
				MappingType:         fwtypes.StringEnumValue(awstypes.ResourceMappingTypeTerraform),
				ResourceName:        types.StringValue(resourceName),
				TerraformSourceName: types.StringValue(terraformSourceName),
			}

			// Create physical resource ID with S3 URL
			physicalId := physicalResourceIdModel{
				Type:       fwtypes.StringEnumValue(awstypes.PhysicalIdentifierTypeNative),
				Identifier: types.StringValue(s3Url),
			}

			physicalIdValue, convertDiags := types.ObjectValueFrom(ctx, map[string]attr.Type{
				names.AttrType:         fwtypes.StringEnumType[awstypes.PhysicalIdentifierType](),
				names.AttrIdentifier:   types.StringType,
				names.AttrAWSAccountID: types.StringType,
				"aws_region":           types.StringType,
			}, physicalId)
			diags.Append(convertDiags...)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{}), diags
			}

			mappingModel.PhysicalResourceId = types.ListValueMust(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					names.AttrType:         fwtypes.StringEnumType[awstypes.PhysicalIdentifierType](),
					names.AttrIdentifier:   types.StringType,
					names.AttrAWSAccountID: types.StringType,
					"aws_region":           types.StringType,
				},
			}, []attr.Value{physicalIdValue})

			mappingValue, convertDiags := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"mapping_type":          fwtypes.StringEnumType[awstypes.ResourceMappingType](),
				"resource_name":         types.StringType,
				"terraform_source_name": types.StringType,
				"physical_resource_id":  types.ListType{ElemType: types.ObjectType{}},
			}, mappingModel)
			diags.Append(convertDiags...)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{}), diags
			}

			mappingValues = append(mappingValues, mappingValue)
		}

		// Handle source ARN (direct resource mapping)
		if inputSource.SourceArn != nil {
			mappingModel := resourceMappingModel{
				MappingType:         fwtypes.StringEnumValue(awstypes.ResourceMappingTypeResource),
				ResourceName:        types.StringValue("resource"), // Default name
				TerraformSourceName: types.StringNull(),
			}

			// Create physical resource ID with ARN
			physicalId := physicalResourceIdModel{
				Type:       fwtypes.StringEnumValue(awstypes.PhysicalIdentifierTypeArn),
				Identifier: types.StringValue(aws.ToString(inputSource.SourceArn)),
			}

			physicalIdValue, convertDiags := types.ObjectValueFrom(ctx, map[string]attr.Type{
				names.AttrType:         fwtypes.StringEnumType[awstypes.PhysicalIdentifierType](),
				names.AttrIdentifier:   types.StringType,
				names.AttrAWSAccountID: types.StringType,
				"aws_region":           types.StringType,
			}, physicalId)
			diags.Append(convertDiags...)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{}), diags
			}

			mappingModel.PhysicalResourceId = types.ListValueMust(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					names.AttrType:         fwtypes.StringEnumType[awstypes.PhysicalIdentifierType](),
					names.AttrIdentifier:   types.StringType,
					names.AttrAWSAccountID: types.StringType,
					"aws_region":           types.StringType,
				},
			}, []attr.Value{physicalIdValue})

			mappingValue, convertDiags := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"mapping_type":          fwtypes.StringEnumType[awstypes.ResourceMappingType](),
				"resource_name":         types.StringType,
				"terraform_source_name": types.StringType,
				"physical_resource_id":  types.ListType{ElemType: types.ObjectType{}},
			}, mappingModel)
			diags.Append(convertDiags...)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{}), diags
			}

			mappingValues = append(mappingValues, mappingValue)
		}
	}

	if len(mappingValues) == 0 {
		return types.ListNull(types.ObjectType{}), diags
	}

	return types.ListValueMust(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"mapping_type":          fwtypes.StringEnumType[awstypes.ResourceMappingType](),
			"resource_name":         types.StringType,
			"terraform_source_name": types.StringType,
			"physical_resource_id":  types.ListType{ElemType: types.ObjectType{}},
		},
	}, mappingValues), diags
}
