// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arczonalshift

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arczonalshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arczonalshift/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	flex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_arczonalshift_zonal_autoshift_configuration", name="Zonal Autoshift Configuration")
// @ArnIdentity("resource_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/arczonalshift;arczonalshift.GetManagedResourceOutput")
// @Testing(hasNoPreExistingResource=true)
func newResourceZonalAutoshiftConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceZonalAutoshiftConfiguration{}

	return r, nil
}

const (
	ResNameZonalAutoshiftConfiguration = "Zonal Autoshift Configuration"
)

type resourceZonalAutoshiftConfiguration struct {
	framework.ResourceWithModel[resourceZonalAutoshiftConfigurationModel]
	framework.WithImportByIdentity
}

func (r *resourceZonalAutoshiftConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allowed_windows": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of time windows during which practice runs are allowed, in the format `Day:HH:MM-Day:HH:MM` (e.g., `Mon:09:00-Mon:17:00`). Cannot be used together with `blocked_windows`.",
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("blocked_windows")),
				},
			},
			"blocked_dates": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of dates when practice runs should not be started, in the format `YYYY-MM-DD`.",
			},
			"blocked_windows": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of time windows during which practice runs should not be started, in the format `Day:HH:MM-Day:HH:MM` (e.g., `Mon:00:00-Mon:08:00`). Cannot be used together with `allowed_windows`.",
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("allowed_windows")),
				},
			},
			names.AttrResourceARN: schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "The ARN of the managed resource to configure zonal autoshift for (e.g., an Application Load Balancer). Changing this creates a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zonal_autoshift_status": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.ZonalAutoshiftStatus](),
				Required:    true,
				Description: "The status of zonal autoshift. Valid values: `ENABLED`, `DISABLED`.",
			},
		},
		Blocks: map[string]schema.Block{
			"blocking_alarms": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[controlConditionModel](ctx),
				Description: "List of CloudWatch alarms that can block practice runs when in alarm state.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alarm_identifier": schema.StringAttribute{
							CustomType:  fwtypes.ARNType,
							Required:    true,
							Description: "ARN of the CloudWatch alarm.",
						},
						names.AttrType: schema.StringAttribute{
							CustomType:  fwtypes.StringEnumType[awstypes.ControlConditionType](),
							Required:    true,
							Description: "Type of control condition. Valid value: `CLOUDWATCH`.",
						},
					},
				},
			},
			"outcome_alarms": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[controlConditionModel](ctx),
				Description: "List of CloudWatch alarms monitored during practice runs. These alarms help determine the health of your application during zonal shifts.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alarm_identifier": schema.StringAttribute{
							CustomType:  fwtypes.ARNType,
							Required:    true,
							Description: "ARN of the CloudWatch alarm.",
						},
						names.AttrType: schema.StringAttribute{
							CustomType:  fwtypes.StringEnumType[awstypes.ControlConditionType](),
							Required:    true,
							Description: "Type of control condition. Valid value: `CLOUDWATCH`.",
						},
					},
				},
			},
		},
	}
}

func (r *resourceZonalAutoshiftConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var plan resourceZonalAutoshiftConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := arczonalshift.CreatePracticeRunConfigurationInput{
		ResourceIdentifier: plan.ResourceARN.ValueStringPointer(),
	}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.BlockedDates = flex.ExpandFrameworkStringValueList(ctx, plan.BlockedDates)
	input.BlockedWindows = flex.ExpandFrameworkStringValueList(ctx, plan.BlockedWindows)
	input.AllowedWindows = flex.ExpandFrameworkStringValueList(ctx, plan.AllowedWindows)

	out, err := conn.CreatePracticeRunConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceARN.String())
		return
	}

	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output from createPracticeRunConfiguration"), smerr.ID, plan.ResourceARN.String())
		return
	}

	statusInput := arczonalshift.UpdateZonalAutoshiftConfigurationInput{
		ResourceIdentifier:   plan.ResourceARN.ValueStringPointer(),
		ZonalAutoshiftStatus: plan.ZonalAutoshiftStatus.ValueEnum(),
	}

	out2, err := conn.UpdateZonalAutoshiftConfiguration(ctx, &statusInput)
	if err != nil {
		// cleanup practiceRunConfiguration created earlier
		deletePracticeRunConfigurationInput := arczonalshift.DeletePracticeRunConfigurationInput{
			ResourceIdentifier: plan.ResourceARN.ValueStringPointer(),
		}
		_, cleanupErr := conn.DeletePracticeRunConfiguration(ctx, &deletePracticeRunConfigurationInput)
		if cleanupErr != nil {
			smerr.AddError(ctx, &resp.Diagnostics, cleanupErr, smerr.ID, plan.ResourceARN.String())
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceARN.String())
		return
	}

	if out2 == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output from updateZonalAutoshiftConfiguration"), smerr.ID, plan.ResourceARN.String())
		return
	}

	plan.ResourceARN = flex.StringToFrameworkARN(ctx, out2.ResourceIdentifier)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceZonalAutoshiftConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var state resourceZonalAutoshiftConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findManagedResourceByIdentifier(ctx, conn, state.ResourceARN.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ARCZonalShift, create.ErrActionReading, ResNameZonalAutoshiftConfiguration, state.ResourceARN.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil || out.PracticeRunConfiguration == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(r.flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceZonalAutoshiftConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var plan, state resourceZonalAutoshiftConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	resourceIdentifier := plan.ResourceARN.ValueString()

	practiceRunChanged := !plan.OutcomeAlarms.Equal(state.OutcomeAlarms) ||
		!plan.BlockingAlarms.Equal(state.BlockingAlarms) ||
		!plan.BlockedDates.Equal(state.BlockedDates) ||
		!plan.BlockedWindows.Equal(state.BlockedWindows) ||
		!plan.AllowedWindows.Equal(state.AllowedWindows)

	if practiceRunChanged {
		var input arczonalshift.UpdatePracticeRunConfigurationInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.ResourceIdentifier = aws.String(resourceIdentifier)
		input.BlockedDates = flex.ExpandFrameworkStringValueList(ctx, plan.BlockedDates)
		if !plan.BlockedWindows.IsNull() {
			input.BlockedWindows = flex.ExpandFrameworkStringValueList(ctx, plan.BlockedWindows)
			input.AllowedWindows = []string{}
		} else if !plan.AllowedWindows.IsNull() {
			input.AllowedWindows = flex.ExpandFrameworkStringValueList(ctx, plan.AllowedWindows)
			input.BlockedWindows = []string{}
		} else {
			input.BlockedWindows = []string{}
			input.AllowedWindows = []string{}
		}

		_, err := conn.UpdatePracticeRunConfiguration(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceIdentifier)
			return
		}
	}

	if !plan.ZonalAutoshiftStatus.Equal(state.ZonalAutoshiftStatus) {
		input := arczonalshift.UpdateZonalAutoshiftConfigurationInput{
			ResourceIdentifier:   aws.String(resourceIdentifier),
			ZonalAutoshiftStatus: plan.ZonalAutoshiftStatus.ValueEnum(),
		}
		_, err := conn.UpdateZonalAutoshiftConfiguration(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceIdentifier)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceZonalAutoshiftConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var state resourceZonalAutoshiftConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	resourceIdentifier := state.ResourceARN.ValueString()

	statusInput := arczonalshift.UpdateZonalAutoshiftConfigurationInput{
		ResourceIdentifier:   aws.String(resourceIdentifier),
		ZonalAutoshiftStatus: awstypes.ZonalAutoshiftStatusDisabled,
	}
	_, err := conn.UpdateZonalAutoshiftConfiguration(ctx, &statusInput)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceIdentifier)
		return
	}

	deleteInput := arczonalshift.DeletePracticeRunConfigurationInput{
		ResourceIdentifier: aws.String(resourceIdentifier),
	}
	_, err = conn.DeletePracticeRunConfiguration(ctx, &deleteInput)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceIdentifier)
		return
	}
}

func (r *resourceZonalAutoshiftConfiguration) flatten(ctx context.Context, out *arczonalshift.GetManagedResourceOutput, data *resourceZonalAutoshiftConfigurationModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ResourceARN = flex.StringToFrameworkARN(ctx, out.Arn)
	data.ZonalAutoshiftStatus = fwtypes.StringEnumValue(out.ZonalAutoshiftStatus)
	diags.Append(flex.Flatten(ctx, out.PracticeRunConfiguration, data)...)
	data.BlockedDates = flex.FlattenFrameworkStringValueListOfString(ctx, out.PracticeRunConfiguration.BlockedDates)
	data.BlockedWindows = flex.FlattenFrameworkStringValueListOfString(ctx, out.PracticeRunConfiguration.BlockedWindows)
	data.AllowedWindows = flex.FlattenFrameworkStringValueListOfString(ctx, out.PracticeRunConfiguration.AllowedWindows)

	return diags
}

// Finder functions
func findManagedResourceByIdentifier(ctx context.Context, conn *arczonalshift.Client, resourceIdentifier string) (*arczonalshift.GetManagedResourceOutput, error) {
	input := arczonalshift.GetManagedResourceInput{
		ResourceIdentifier: aws.String(resourceIdentifier),
	}

	out, err := conn.GetManagedResource(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return out, nil
}

type resourceZonalAutoshiftConfigurationModel struct {
	framework.WithRegionModel
	AllowedWindows       fwtypes.ListOfString                                   `tfsdk:"allowed_windows"`
	BlockedDates         fwtypes.ListOfString                                   `tfsdk:"blocked_dates"`
	BlockedWindows       fwtypes.ListOfString                                   `tfsdk:"blocked_windows"`
	BlockingAlarms       fwtypes.ListNestedObjectValueOf[controlConditionModel] `tfsdk:"blocking_alarms"`
	OutcomeAlarms        fwtypes.ListNestedObjectValueOf[controlConditionModel] `tfsdk:"outcome_alarms"`
	ResourceARN          fwtypes.ARN                                            `tfsdk:"resource_arn"`
	ZonalAutoshiftStatus fwtypes.StringEnum[awstypes.ZonalAutoshiftStatus]      `tfsdk:"zonal_autoshift_status"`
}

type controlConditionModel struct {
	AlarmIdentifier fwtypes.ARN                                       `tfsdk:"alarm_identifier"`
	Type            fwtypes.StringEnum[awstypes.ControlConditionType] `tfsdk:"type"`
}
