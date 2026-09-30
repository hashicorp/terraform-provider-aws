// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ds

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_directory_service_ad_assessment", name="AD Assessment")
// @IdentityAttribute("assessment_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(tagsTest=false)
// @Testing(identityTest=false)
func newADAssessmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &adAssessmentResource{}

	r.SetDefaultCreateTimeout(45 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type adAssessmentResource struct {
	framework.ResourceWithModel[adAssessmentResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

const (
	ResNameADAssessment = "AD Assessment"
)

func (r *adAssessmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"assessment_id": framework.IDAttribute(),
			"customer_dns_ips": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(2, 2),
				},
			},
			names.AttrDNSName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrSecurityGroupIDs: schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
				Computed:   true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 1),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
					setplanmodifier.RequiresReplace(),
				},
			},
			"self_managed_instance_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeBetween(2, 2),
				},
			},
			names.AttrSubnetIDs: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(2, 2),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
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

func (r *adAssessmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DSClient(ctx)

	var plan adAssessmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input = directoryservice.StartADAssessmentInput{
		AssessmentConfiguration: &awstypes.AssessmentConfiguration{
			CustomerDnsIps: flex.ExpandFrameworkStringValueList(ctx, plan.CustomerDnsIps),
			DnsName:        plan.DnsName.ValueStringPointer(),
			InstanceIds:    flex.ExpandFrameworkStringValueList(ctx, plan.SelfManagedInstanceIds),
			VpcSettings: &awstypes.DirectoryVpcSettings{
				SubnetIds: flex.ExpandFrameworkStringValueSet(ctx, plan.SubnetIds),
				VpcId:     plan.VpcId.ValueStringPointer(),
			},
		},
	}

	if !plan.SecurityGroupIds.IsNull() {
		input.AssessmentConfiguration.SecurityGroupIds = flex.ExpandFrameworkStringValueSet(ctx, plan.SecurityGroupIds)
	}

	out, err := conn.StartADAssessment(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}
	if out == nil || out.AssessmentId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID)
		return
	}
	plan.AssessmentId = flex.StringToFramework(ctx, out.AssessmentId)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root("assessment_id"), plan.AssessmentId))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	created, err := waitADAssessmentCreated(ctx, conn, plan.AssessmentId.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, created, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *adAssessmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DSClient(ctx)

	var state adAssessmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findADAssessmentByID(ctx, conn, state.AssessmentId.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AssessmentId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *adAssessmentResource) flatten(ctx context.Context, adAssessment *awstypes.Assessment, data *adAssessmentResourceModel) (diags diag.Diagnostics) {
	diags.Append(flex.Flatten(ctx, adAssessment, data)...)
	return diags
}

func (r *adAssessmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DSClient(ctx)

	var state adAssessmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := directoryservice.DeleteADAssessmentInput{
		AssessmentId: state.AssessmentId.ValueStringPointer(),
	}

	_, err := conn.DeleteADAssessment(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityDoesNotExistException](err) || errs.IsA[*awstypes.ServiceException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AssessmentId.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitADAssessmentDeleted(ctx, conn, state.AssessmentId.ValueString(), deleteTimeout)
	if err != nil {
		if errs.IsA[*awstypes.EntityDoesNotExistException](err) || errs.IsA[*awstypes.ServiceException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AssessmentId.String())
		return
	}
}

const (
	statusSuccess    = "SUCCESS"
	statusFailed     = "FAILED"
	statusPending    = "PENDING"
	statusInProgress = "IN_PROGRESS"
)

func waitADAssessmentCreated(ctx context.Context, conn *directoryservice.Client, id string, timeout time.Duration) (*awstypes.Assessment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusPending, statusInProgress},
		Target:                    []string{statusSuccess},
		Refresh:                   statusADAssessment(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Assessment); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitADAssessmentDeleted(ctx context.Context, conn *directoryservice.Client, id string, timeout time.Duration) (*awstypes.Assessment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusFailed, statusPending, statusSuccess, statusInProgress},
		Target:  []string{},
		Refresh: statusADAssessment(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Assessment); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusADAssessment(conn *directoryservice.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findADAssessmentByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, aws.ToString(out.Status), nil
	}
}

func findADAssessmentByID(ctx context.Context, conn *directoryservice.Client, id string) (*awstypes.Assessment, error) {
	input := directoryservice.DescribeADAssessmentInput{
		AssessmentId: aws.String(id),
	}

	out, err := conn.DescribeADAssessment(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Assessment == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Assessment, nil
}

type adAssessmentResourceModel struct {
	framework.WithRegionModel
	AssessmentId           types.String                      `tfsdk:"assessment_id"`
	CustomerDnsIps         fwtypes.ListValueOf[types.String] `tfsdk:"customer_dns_ips"`
	DnsName                types.String                      `tfsdk:"dns_name"`
	SecurityGroupIds       fwtypes.SetValueOf[types.String]  `tfsdk:"security_group_ids"`
	SelfManagedInstanceIds fwtypes.ListValueOf[types.String] `tfsdk:"self_managed_instance_ids"`
	SubnetIds              fwtypes.SetValueOf[types.String]  `tfsdk:"subnet_ids"`
	VpcId                  types.String                      `tfsdk:"vpc_id"`
	Timeouts               timeouts.Value                    `tfsdk:"timeouts"`
}

func sweepADAssessments(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := directoryservice.ListADAssessmentsInput{}
	conn := client.DSClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := directoryservice.NewListADAssessmentsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Assessments {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newADAssessmentResource, client,
				sweepfw.NewAttribute("assessment_id", aws.ToString(v.AssessmentId))),
			)
		}
	}

	return sweepResources, nil
}
