package auditmanager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const iamPropagationTimeout = 2 * time.Minute

func init() {
	_sp.registerFrameworkResourceFactory(newResourceAssessment)
}

func newResourceAssessment(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceAssessment{}, nil
}

const (
	ResNameAssessment = "Assessment"
)

type resourceAssessment struct {
	framework.ResourceWithConfigure
}

func (r *resourceAssessment) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_auditmanager_assessment"
}

func (r *resourceAssessment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
			},
			"framework_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
			},
			// The roles attribute is split into "roles" and "roles_all" to account for roles
			// that are given access to assessments by default. It isn't possible for this attribute
			// to be both Required (CreateAssessment and UpdateAssessment both require non-empty
			// values) and Computed (capturing roles with access by default and returned in
			// the response output). "roles" stores the items specifically added by the practitioner,
			// while "roles_all" will track everything with access to the assessment.
			//
			// Both attributes are defined as schema.SetAttribute's here rather than in the Blocks
			// section to allow for Required/Computed to be set explicitly.
			"roles": schema.SetAttribute{
				Required:    true,
				ElementType: types.ObjectType{AttrTypes: assessmentRolesAttrTypes},
			},
			"roles_all": schema.SetAttribute{
				Computed:    true,
				ElementType: types.ObjectType{AttrTypes: assessmentRolesAttrTypes},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"tags":     tftags.TagsAttribute(),
			"tags_all": tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"assessment_reports_destination": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"destination": schema.StringAttribute{
							Required: true,
						},
						"destination_type": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.AssessmentReportDestinationType](),
							},
						},
					},
				},
			},
			"scope": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"aws_accounts": schema.SetNestedBlock{
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"aws_services": schema.SetNestedBlock{
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"service_name": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceAssessment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AuditManagerClient()

	var plan resourceAssessmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var reportsDestination []assessmentReportsDestinationData
	resp.Diagnostics.Append(plan.AssessmentReportsDestination.ElementsAs(ctx, &reportsDestination, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var roles []assessmentRolesData
	resp.Diagnostics.Append(plan.Roles.ElementsAs(ctx, &roles, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var scope []assessmentScopeData
	resp.Diagnostics.Append(plan.Scope.ElementsAs(ctx, &scope, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopeInput, d := expandAssessmentScope(ctx, scope)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := auditmanager.CreateAssessmentInput{
		AssessmentReportsDestination: expandAssessmentReportsDestination(reportsDestination),
		FrameworkId:                  aws.String(plan.FrameworkID.ValueString()),
		Name:                         aws.String(plan.Name.ValueString()),
		Roles:                        expandAssessmentRoles(roles),
		Scope:                        scopeInput,
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	defaultTagsConfig := r.Meta().DefaultTagsConfig
	ignoreTagsConfig := r.Meta().IgnoreTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(plan.Tags))
	plan.TagsAll = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	// Include retry handling to allow for IAM propagation
	//
	// Example:
	//   ResourceNotFoundException: The operation tried to access a nonexistent resource. The resource
	//   might not be specified correctly, or its status might not be active. Check and try again.
	var out *auditmanager.CreateAssessmentOutput
	err := tfresource.Retry(ctx, iamPropagationTimeout, func() *sdkv2resource.RetryError {
		var err error
		out, err = conn.CreateAssessment(ctx, &in)
		if err != nil {
			var nfe *awstypes.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return sdkv2resource.RetryableError(err)
			}
			return sdkv2resource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameAssessment, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil || out.Assessment == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameAssessment, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	resp.Diagnostics.Append(state.refreshFromOutput(ctx, r.Meta(), out.Assessment)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceAssessment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AuditManagerClient()

	var state resourceAssessmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindAssessmentByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.AddWarning(
			"AWS Resource Not Found During Refresh",
			fmt.Sprintf("Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original Error: %s", err.Error()),
		)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionReading, ResNameAssessment, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(state.refreshFromOutput(ctx, r.Meta(), out)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAssessment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().AuditManagerClient()

	var plan, state resourceAssessmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.AssessmentReportsDestination.Equal(state.AssessmentReportsDestination) ||
		!plan.Description.Equal(state.Description) ||
		!plan.Name.Equal(state.Name) ||
		!plan.Roles.Equal(state.Roles) ||
		!plan.Scope.Equal(state.Scope) {
		var reportsDestination []assessmentReportsDestinationData
		resp.Diagnostics.Append(plan.AssessmentReportsDestination.ElementsAs(ctx, &reportsDestination, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var roles []assessmentRolesData
		resp.Diagnostics.Append(plan.Roles.ElementsAs(ctx, &roles, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var scope []assessmentScopeData
		resp.Diagnostics.Append(plan.Scope.ElementsAs(ctx, &scope, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		scopeInput, d := expandAssessmentScope(ctx, scope)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in := &auditmanager.UpdateAssessmentInput{
			AssessmentId:                 aws.String(plan.ID.ValueString()),
			AssessmentName:               aws.String(plan.Name.ValueString()),
			AssessmentReportsDestination: expandAssessmentReportsDestination(reportsDestination),
			Roles:                        expandAssessmentRoles(roles),
			Scope:                        scopeInput,
		}

		if !plan.Description.IsNull() {
			in.AssessmentDescription = aws.String(plan.Description.ValueString())
		}

		out, err := conn.UpdateAssessment(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AuditManager, create.ErrActionUpdating, ResNameAssessment, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}
		if out == nil || out.Assessment == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AuditManager, create.ErrActionUpdating, ResNameAssessment, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		resp.Diagnostics.Append(state.refreshFromOutput(ctx, r.Meta(), out.Assessment)...)
		state.Roles = plan.Roles
	}

	if !plan.TagsAll.Equal(state.TagsAll) {
		if err := UpdateTags(ctx, conn, plan.ARN.ValueString(), state.TagsAll, plan.TagsAll); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AuditManager, create.ErrActionUpdating, ResNameAssessment, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}
		state.Tags = plan.Tags
		state.TagsAll = plan.TagsAll
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAssessment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AuditManagerClient()

	var state resourceAssessmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteAssessment(ctx, &auditmanager.DeleteAssessmentInput{
		AssessmentId: aws.String(state.ID.ValueString()),
	})
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionDeleting, ResNameAssessment, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceAssessment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceAssessment) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func FindAssessmentByID(ctx context.Context, conn *auditmanager.Client, id string) (*awstypes.Assessment, error) {
	in := &auditmanager.GetAssessmentInput{
		AssessmentId: aws.String(id),
	}
	out, err := conn.GetAssessment(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &sdkv2resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Assessment == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Assessment, nil
}

var (
	assessmentReportsDestinationAttrTypes = map[string]attr.Type{
		"destination":      types.StringType,
		"destination_type": types.StringType,
	}

	assessmentRolesAttrTypes = map[string]attr.Type{
		"role_arn":  types.StringType,
		"role_type": types.StringType,
	}

	assessmentScopeAttrTypes = map[string]attr.Type{
		"aws_accounts": types.SetType{ElemType: types.ObjectType{AttrTypes: assessmentScopeAWSAccountsAttrTypes}},
		"aws_services": types.SetType{ElemType: types.ObjectType{AttrTypes: assessmentScopeAWSServicesAttrTypes}},
	}

	assessmentScopeAWSAccountsAttrTypes = map[string]attr.Type{ // nosemgrep:ci.aws-in-var-name
		"id": types.StringType,
	}

	assessmentScopeAWSServicesAttrTypes = map[string]attr.Type{ // nosemgrep:ci.aws-in-var-name
		"service_name": types.StringType,
	}
)

type resourceAssessmentData struct {
	ARN                          types.String `tfsdk:"arn"`
	AssessmentReportsDestination types.List   `tfsdk:"assessment_reports_destination"`
	Description                  types.String `tfsdk:"description"`
	ID                           types.String `tfsdk:"id"`
	FrameworkID                  types.String `tfsdk:"framework_id"`
	Name                         types.String `tfsdk:"name"`
	Roles                        types.Set    `tfsdk:"roles"`
	RolesAll                     types.Set    `tfsdk:"roles_all"`
	Scope                        types.List   `tfsdk:"scope"`
	Status                       types.String `tfsdk:"status"`
	Tags                         types.Map    `tfsdk:"tags"`
	TagsAll                      types.Map    `tfsdk:"tags_all"`
}

type assessmentReportsDestinationData struct {
	Destination     types.String `tfsdk:"destination"`
	DestinationType types.String `tfsdk:"destination_type"`
}

type assessmentRolesData struct {
	RoleARN  types.String `tfsdk:"role_arn"`
	RoleType types.String `tfsdk:"role_type"`
}

type assessmentScopeData struct {
	AWSAccounts types.Set `tfsdk:"aws_accounts"`
	AWSServices types.Set `tfsdk:"aws_services"`
}

type assessmentScopeAWSAccountsData struct {
	ID types.String `tfsdk:"id"`
}

type assessmentScopeAWSServicesData struct {
	ServiceName types.String `tfsdk:"service_name"`
}

// refreshFromOutput writes state data from an AWS response object
func (rd *resourceAssessmentData) refreshFromOutput(ctx context.Context, meta *conns.AWSClient, out *awstypes.Assessment) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil || out.Metadata == nil {
		return diags
	}
	metadata := out.Metadata

	rd.ARN = flex.StringToFramework(ctx, out.Arn)
	rd.Description = flex.StringToFramework(ctx, metadata.Description)
	if out.Framework != nil {
		rd.FrameworkID = flex.StringToFramework(ctx, out.Framework.Id)
	}
	rd.ID = flex.StringToFramework(ctx, metadata.Id)
	rd.Name = flex.StringToFramework(ctx, metadata.Name)
	rd.Status = flex.StringValueToFramework(ctx, metadata.Status)

	reportsDestination, d := flattenAssessmentReportsDestination(ctx, metadata.AssessmentReportsDestination)
	diags.Append(d...)
	rd.AssessmentReportsDestination = reportsDestination
	roles, d := flattenAssessmentRoles(ctx, metadata.Roles)
	diags.Append(d...)
	rd.RolesAll = roles
	scope, d := flattenAssessmentScope(ctx, metadata.Scope)
	diags.Append(d...)
	rd.Scope = scope

	defaultTagsConfig := meta.DefaultTagsConfig
	ignoreTagsConfig := meta.IgnoreTagsConfig
	tags := KeyValueTags(out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	// AWS APIs often return empty lists of tags when none have been configured.
	if tags := tags.RemoveDefaultConfig(defaultTagsConfig).Map(); len(tags) == 0 {
		rd.Tags = tftags.Null
	} else {
		rd.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags)
	}
	rd.TagsAll = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.Map())

	return diags
}

func expandAssessmentReportsDestination(tfList []assessmentReportsDestinationData) *awstypes.AssessmentReportsDestination {
	if len(tfList) == 0 {
		return nil
	}
	rd := tfList[0]
	return &awstypes.AssessmentReportsDestination{
		Destination:     aws.String(rd.Destination.ValueString()),
		DestinationType: awstypes.AssessmentReportDestinationType(rd.DestinationType.ValueString()),
	}
}

func expandAssessmentRoles(tfList []assessmentRolesData) []awstypes.Role {
	var roles []awstypes.Role
	for _, item := range tfList {
		new := awstypes.Role{
			RoleArn:  aws.String(item.RoleARN.ValueString()),
			RoleType: awstypes.RoleType(item.RoleType.ValueString()),
		}
		roles = append(roles, new)
	}
	return roles
}

func expandAssessmentScope(ctx context.Context, tfList []assessmentScopeData) (*awstypes.Scope, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	scope := tfList[0]

	var accounts []assessmentScopeAWSAccountsData
	diags.Append(scope.AWSAccounts.ElementsAs(ctx, &accounts, false)...)
	var services []assessmentScopeAWSServicesData
	diags.Append(scope.AWSServices.ElementsAs(ctx, &services, false)...)

	return &awstypes.Scope{
		AwsAccounts: expandAssessmentScopeAWSAccounts(accounts),
		AwsServices: expandAssessmentScopeAWSServices(services),
	}, diags
}

func expandAssessmentScopeAWSAccounts(tfList []assessmentScopeAWSAccountsData) []awstypes.AWSAccount { // nosemgrep:ci.aws-in-func-name
	var accounts []awstypes.AWSAccount
	for _, item := range tfList {
		new := awstypes.AWSAccount{
			Id: aws.String(item.ID.ValueString()),
		}
		accounts = append(accounts, new)
	}
	return accounts
}

func expandAssessmentScopeAWSServices(tfList []assessmentScopeAWSServicesData) []awstypes.AWSService { // nosemgrep:ci.aws-in-func-name
	var services []awstypes.AWSService
	for _, item := range tfList {
		new := awstypes.AWSService{
			ServiceName: aws.String(item.ServiceName.ValueString()),
		}
		services = append(services, new)
	}
	return services
}

func flattenAssessmentReportsDestination(ctx context.Context, apiObject *awstypes.AssessmentReportsDestination) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: assessmentReportsDestinationAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"destination":      flex.StringToFramework(ctx, apiObject.Destination),
		"destination_type": flex.StringValueToFramework(ctx, apiObject.DestinationType),
	}
	objVal, d := types.ObjectValue(assessmentReportsDestinationAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenAssessmentRoles(ctx context.Context, apiObject []awstypes.Role) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: assessmentRolesAttrTypes}

	if len(apiObject) == 0 {
		return types.SetValueMust(elemType, []attr.Value{}), diags
	}

	elems := []attr.Value{}
	for _, role := range apiObject {
		obj := map[string]attr.Value{
			"role_arn":  flex.StringToFramework(ctx, role.RoleArn),
			"role_type": flex.StringValueToFramework(ctx, role.RoleType),
		}
		objVal, d := types.ObjectValue(assessmentRolesAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func flattenAssessmentScope(ctx context.Context, apiObject *awstypes.Scope) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: assessmentScopeAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	accounts, d := flattenAssessmentScopeAWSAccounts(ctx, apiObject.AwsAccounts)
	diags.Append(d...)
	services, d := flattenAssessmentScopeAWSServices(ctx, apiObject.AwsServices)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"aws_accounts": accounts,
		"aws_services": services,
	}
	objVal, d := types.ObjectValue(assessmentScopeAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenAssessmentScopeAWSAccounts(ctx context.Context, apiObject []awstypes.AWSAccount) (types.Set, diag.Diagnostics) { // nosemgrep:ci.aws-in-func-name
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: assessmentScopeAWSAccountsAttrTypes}

	if apiObject == nil {
		return types.SetValueMust(elemType, []attr.Value{}), diags
	}

	elems := []attr.Value{}
	for _, account := range apiObject {
		obj := map[string]attr.Value{
			"id": flex.StringToFramework(ctx, account.Id),
		}
		objVal, d := types.ObjectValue(assessmentScopeAWSAccountsAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func flattenAssessmentScopeAWSServices(ctx context.Context, apiObject []awstypes.AWSService) (types.Set, diag.Diagnostics) { // nosemgrep:ci.aws-in-func-name
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: assessmentScopeAWSServicesAttrTypes}

	if apiObject == nil {
		return types.SetValueMust(elemType, []attr.Value{}), diags
	}

	elems := []attr.Value{}
	for _, service := range apiObject {
		obj := map[string]attr.Value{
			"service_name": flex.StringToFramework(ctx, service.ServiceName),
		}
		objVal, d := types.ObjectValue(assessmentScopeAWSServicesAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}
