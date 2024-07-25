// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_security_group_ingress_rule", name="Security Group Ingress Rule")
// @Tags(identifierAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;types.SecurityGroupRule")
func newSecurityGroupIngressRuleResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &securityGroupIngressRuleResource{}
	r.securityGroupRule = r

	return r, nil
}

type securityGroupIngressRuleResource struct {
	securityGroupRuleResource
}

func (*securityGroupIngressRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_vpc_security_group_ingress_rule"
}

func (r *securityGroupIngressRuleResource) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{
		{
			SourceSchema: legacySecurityGroupRuleResourceSchemaV2(ctx),
			StateMover:   r.moveStateResourceSecurityGroupRule,
		},
	}
}

func (r *securityGroupIngressRuleResource) create(ctx context.Context, data *securityGroupRuleResourceModel) (string, error) {
	conn := r.Meta().EC2Client(ctx)

	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       fwflex.StringFromFramework(ctx, data.SecurityGroupID),
		IpPermissions: []awstypes.IpPermission{data.expandIPPermission(ctx)},
	}

	output, err := conn.AuthorizeSecurityGroupIngress(ctx, input)

	if err != nil {
		return "", err
	}

	return aws.ToString(output.SecurityGroupRules[0].SecurityGroupRuleId), nil
}

func (r *securityGroupIngressRuleResource) delete(ctx context.Context, data *securityGroupRuleResourceModel) error {
	conn := r.Meta().EC2Client(ctx)

	_, err := conn.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{
		GroupId:              fwflex.StringFromFramework(ctx, data.SecurityGroupID),
		SecurityGroupRuleIds: fwflex.StringSliceValueFromFramework(ctx, data.ID),
	})

	return err
}

func (r *securityGroupIngressRuleResource) findByID(ctx context.Context, id string) (*awstypes.SecurityGroupRule, error) {
	conn := r.Meta().EC2Client(ctx)

	return findSecurityGroupIngressRuleByID(ctx, conn, id)
}

// moveStateResourceSecurityGroupRule transforms the state of an `aws_security_group_rule` resource to this resource's schema.
func (r *securityGroupIngressRuleResource) moveStateResourceSecurityGroupRule(ctx context.Context, request resource.MoveStateRequest, response *resource.MoveStateResponse) {
	if request.SourceTypeName != "aws_security_group_rule" {
		return
	}

	if request.SourceSchemaVersion != 2 {
		return
	}

	if !strings.HasSuffix(request.SourceProviderAddress, "hashicorp/aws") {
		return
	}

	var source legacySecurityGroupRuleResourceModel
	response.Diagnostics.Append(request.SourceState.Get(ctx, &source)...)
	if response.Diagnostics.HasError() {
		return
	}

	// TODO: Need to find the security group rule ID.

	// if typ := source.Type.ValueEnum(); typ != securityGroupRuleTypeIngress {
	// 	response.Diagnostics.AddError("Incorrect Type", string(typ))
	// 	return
	// }

	// nCIDRs := 0
	// if !source.CIDRBlocks.IsNull() {
	// 	nCIDRs += len(source.CIDRBlocks.Elements())
	// }
	// nIPv6CIDRs := 0
	// if !source.IPv6CIDRBlocksBlocks.IsNull() {
	// 	nIPv6CIDRs += len(source.IPv6CIDRBlocksBlocks.Elements())
	// }
	// nPrefxListIDs := 0
	// if !source.PrefixListIDs.IsNull() {
	// 	nPrefxListIDs += len(source.PrefixListIDs.Elements())
	// }
	// nSourceSecurityGroupIDs := 0
	// if !source.SourceSecurityGroupID.IsNull() && source.SourceSecurityGroupID.ValueString() != "" {
	// 	nSourceSecurityGroupIDs = 1
	// }

	// if nCIDRs+nIPv6CIDRs+nPrefxListIDs+nSourceSecurityGroupIDs > 1 {
	// 	response.Diagnostics.AddError("Multiple Sources", "Only one source is allowed")
	// 	return
	// }

	// target := &securityGroupRuleResourceModel{
	// 	// ARN: 				  r.securityGroupRuleARN(ctx, securityGroupRuleID),
	// 	Description: fwflex.EmptyStringAsNull(source.Description),
	// }

	// response.Diagnostics.Append(response.TargetState.Set(ctx, target)...)
}

// Base structure and methods for VPC security group rules.

type securityGroupRule interface {
	create(context.Context, *securityGroupRuleResourceModel) (string, error)
	delete(context.Context, *securityGroupRuleResourceModel) error
	findByID(context.Context, string) (*awstypes.SecurityGroupRule, error)
}

type securityGroupRuleResource struct {
	securityGroupRule
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *securityGroupRuleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cidr_ipv4": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.IPv4CIDRNetworkAddress(),
				},
			},
			"cidr_ipv6": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.IPv6CIDRNetworkAddress(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"from_port": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(-1, 65535),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"ip_protocol": schema.StringAttribute{
				CustomType: ipProtocolType{},
				Required:   true,
			},
			"prefix_list_id": schema.StringAttribute{
				Optional: true,
			},
			"referenced_security_group_id": schema.StringAttribute{
				Optional: true,
			},
			"security_group_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"security_group_rule_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"to_port": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(-1, 65535),
				},
			},
		},
	}
}

func (r *securityGroupRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data securityGroupRuleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	securityGroupRuleID, err := r.securityGroupRule.create(ctx, &data)

	if err != nil {
		response.Diagnostics.AddError("creating VPC Security Group Rule", err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = r.securityGroupRuleARN(ctx, securityGroupRuleID)
	data.SecurityGroupRuleID = types.StringValue(securityGroupRuleID)
	data.setID()

	conn := r.Meta().EC2Client(ctx)
	if err := createTags(ctx, conn, data.ID.ValueString(), getTagsIn(ctx)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("setting VPC Security Group Rule (%s) tags", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *securityGroupRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data securityGroupRuleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	output, err := r.securityGroupRule.findByID(ctx, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Security Group Rule (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.ARN = r.securityGroupRuleARN(ctx, data.ID.ValueString())
	data.CIDRIPv4 = fwflex.StringToFramework(ctx, output.CidrIpv4)
	data.CIDRIPv6 = fwflex.StringToFramework(ctx, output.CidrIpv6)
	data.Description = fwflex.StringToFramework(ctx, output.Description)
	data.IPProtocol = fwflex.StringToFrameworkValuable[ipProtocol](ctx, output.IpProtocol)
	data.PrefixListID = fwflex.StringToFramework(ctx, output.PrefixListId)
	data.ReferencedSecurityGroupID = flattenReferencedSecurityGroup(ctx, output.ReferencedGroupInfo, r.Meta().AccountID)
	data.SecurityGroupID = fwflex.StringToFramework(ctx, output.GroupId)
	data.SecurityGroupRuleID = fwflex.StringToFramework(ctx, output.SecurityGroupRuleId)

	// If planned from_port or to_port are null and values of -1 are returned, propagate null.
	if v := aws.ToInt32(output.FromPort); v == -1 && data.FromPort.IsNull() {
		data.FromPort = types.Int64Null()
	} else {
		data.FromPort = fwflex.Int32ToFramework(ctx, output.FromPort)
	}
	if v := aws.ToInt32(output.ToPort); v == -1 && data.ToPort.IsNull() {
		data.ToPort = types.Int64Null()
	} else {
		data.ToPort = fwflex.Int32ToFramework(ctx, output.ToPort)
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *securityGroupRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new securityGroupRuleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	if !new.CIDRIPv4.Equal(old.CIDRIPv4) ||
		!new.CIDRIPv6.Equal(old.CIDRIPv6) ||
		!new.Description.Equal(old.Description) ||
		!new.FromPort.Equal(old.FromPort) ||
		!new.IPProtocol.Equal(old.IPProtocol) ||
		!new.PrefixListID.Equal(old.PrefixListID) ||
		!new.ReferencedSecurityGroupID.Equal(old.ReferencedSecurityGroupID) ||
		!new.ToPort.Equal(old.ToPort) {
		input := &ec2.ModifySecurityGroupRulesInput{
			GroupId: fwflex.StringFromFramework(ctx, new.SecurityGroupID),
			SecurityGroupRules: []awstypes.SecurityGroupRuleUpdate{{
				SecurityGroupRule:   new.expandSecurityGroupRuleRequest(ctx),
				SecurityGroupRuleId: fwflex.StringFromFramework(ctx, new.ID),
			}},
		}

		_, err := conn.ModifySecurityGroupRules(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating VPC Security Group Rule (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *securityGroupRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data securityGroupRuleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting VPC Security Group Rule", map[string]interface{}{
		names.AttrID: data.ID.ValueString(),
	})
	err := r.securityGroupRule.delete(ctx, &data)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound, errCodeInvalidSecurityGroupRuleIdNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPC Security Group Rule (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *securityGroupRuleResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if !request.State.Raw.IsNull() && !request.Plan.Raw.IsNull() {
		var old, new securityGroupRuleResourceModel
		response.Diagnostics.Append(request.State.Get(ctx, &old)...)
		if response.Diagnostics.HasError() {
			return
		}
		response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
		if response.Diagnostics.HasError() {
			return
		}

		// When you modify a rule, you cannot change the rule's source type.
		if new, old := new.sourceAttributeName(), old.sourceAttributeName(); new != old {
			response.RequiresReplace = []path.Path{path.Root(old), path.Root(new)}
		}
	}

	r.SetTagsAll(ctx, request, response)
}

func (r *securityGroupRuleResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("cidr_ipv4"),
			path.MatchRoot("cidr_ipv6"),
			path.MatchRoot("prefix_list_id"),
			path.MatchRoot("referenced_security_group_id"),
		),
	}
}

func (r *securityGroupRuleResource) securityGroupRuleARN(_ context.Context, id string) types.String {
	return types.StringValue(r.RegionalARN(names.EC2, fmt.Sprintf("security-group-rule/%s", id)))
}

func flattenReferencedSecurityGroup(ctx context.Context, apiObject *awstypes.ReferencedSecurityGroup, accountID string) types.String {
	if apiObject == nil {
		return types.StringNull()
	}

	if apiObject.UserId == nil || aws.ToString(apiObject.UserId) == accountID {
		return fwflex.StringToFramework(ctx, apiObject.GroupId)
	}

	// [UserID/]GroupID.
	return types.StringValue(strings.Join([]string{aws.ToString(apiObject.UserId), aws.ToString(apiObject.GroupId)}, "/"))
}

type securityGroupRuleResourceModel struct {
	ARN                       types.String `tfsdk:"arn"`
	CIDRIPv4                  types.String `tfsdk:"cidr_ipv4"`
	CIDRIPv6                  types.String `tfsdk:"cidr_ipv6"`
	Description               types.String `tfsdk:"description"`
	FromPort                  types.Int64  `tfsdk:"from_port"`
	ID                        types.String `tfsdk:"id"`
	IPProtocol                ipProtocol   `tfsdk:"ip_protocol"`
	PrefixListID              types.String `tfsdk:"prefix_list_id"`
	ReferencedSecurityGroupID types.String `tfsdk:"referenced_security_group_id"`
	SecurityGroupID           types.String `tfsdk:"security_group_id"`
	SecurityGroupRuleID       types.String `tfsdk:"security_group_rule_id"`
	Tags                      types.Map    `tfsdk:"tags"`
	TagsAll                   types.Map    `tfsdk:"tags_all"`
	ToPort                    types.Int64  `tfsdk:"to_port"`
}

func (model *securityGroupRuleResourceModel) InitFromID() error {
	model.SecurityGroupRuleID = model.ID

	return nil
}

func (model *securityGroupRuleResourceModel) setID() {
	model.ID = model.SecurityGroupRuleID
}

func (model *securityGroupRuleResourceModel) expandIPPermission(ctx context.Context) awstypes.IpPermission {
	apiObject := awstypes.IpPermission{
		FromPort:   fwflex.Int32FromFramework(ctx, model.FromPort),
		IpProtocol: fwflex.StringFromFramework(ctx, model.IPProtocol),
		ToPort:     fwflex.Int32FromFramework(ctx, model.ToPort),
	}

	if !model.CIDRIPv4.IsNull() {
		apiObject.IpRanges = []awstypes.IpRange{{
			CidrIp:      fwflex.StringFromFramework(ctx, model.CIDRIPv4),
			Description: fwflex.StringFromFramework(ctx, model.Description),
		}}
	}

	if !model.CIDRIPv6.IsNull() {
		apiObject.Ipv6Ranges = []awstypes.Ipv6Range{{
			CidrIpv6:    fwflex.StringFromFramework(ctx, model.CIDRIPv6),
			Description: fwflex.StringFromFramework(ctx, model.Description),
		}}
	}

	if !model.PrefixListID.IsNull() {
		apiObject.PrefixListIds = []awstypes.PrefixListId{{
			PrefixListId: fwflex.StringFromFramework(ctx, model.PrefixListID),
			Description:  fwflex.StringFromFramework(ctx, model.Description),
		}}
	}

	if !model.ReferencedSecurityGroupID.IsNull() {
		apiObject.UserIdGroupPairs = []awstypes.UserIdGroupPair{{
			Description: fwflex.StringFromFramework(ctx, model.Description),
		}}

		// [UserID/]GroupID.
		if parts := strings.Split(model.ReferencedSecurityGroupID.ValueString(), "/"); len(parts) == 2 {
			apiObject.UserIdGroupPairs[0].GroupId = aws.String(parts[1])
			apiObject.UserIdGroupPairs[0].UserId = aws.String(parts[0])
		} else {
			apiObject.UserIdGroupPairs[0].GroupId = fwflex.StringFromFramework(ctx, model.ReferencedSecurityGroupID)
		}
	}

	return apiObject
}

func (model *securityGroupRuleResourceModel) expandSecurityGroupRuleRequest(ctx context.Context) *awstypes.SecurityGroupRuleRequest {
	apiObject := &awstypes.SecurityGroupRuleRequest{
		CidrIpv4:          fwflex.StringFromFramework(ctx, model.CIDRIPv4),
		CidrIpv6:          fwflex.StringFromFramework(ctx, model.CIDRIPv6),
		Description:       fwflex.StringFromFramework(ctx, model.Description),
		FromPort:          fwflex.Int32FromFramework(ctx, model.FromPort),
		IpProtocol:        fwflex.StringFromFramework(ctx, model.IPProtocol),
		PrefixListId:      fwflex.StringFromFramework(ctx, model.PrefixListID),
		ReferencedGroupId: fwflex.StringFromFramework(ctx, model.ReferencedSecurityGroupID),
		ToPort:            fwflex.Int32FromFramework(ctx, model.ToPort),
	}

	return apiObject
}

func (model *securityGroupRuleResourceModel) sourceAttributeName() string {
	switch {
	case !model.CIDRIPv4.IsNull():
		return "cidr_ipv4"
	case !model.CIDRIPv6.IsNull():
		return "cidr_ipv6"
	case !model.PrefixListID.IsNull():
		return "prefix_list_id"
	case !model.ReferencedSecurityGroupID.IsNull():
		return "referenced_security_group_id"
	}

	return ""
}

var (
	_ basetypes.StringTypable                    = (*ipProtocolType)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*ipProtocol)(nil)
)

type ipProtocolType struct {
	basetypes.StringType
}

func (t ipProtocolType) Equal(o attr.Type) bool {
	other, ok := o.(ipProtocolType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (ipProtocolType) String() string {
	return "IPProtocolType"
}

func (t ipProtocolType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return ipProtocolNull(), diags
	}
	if in.IsUnknown() {
		return ipProtocolUnknown(), diags
	}

	return ipProtocolValue(in.ValueString()), diags
}

func (t ipProtocolType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (ipProtocolType) ValueType(context.Context) attr.Value {
	return ipProtocol{}
}

type ipProtocol struct {
	basetypes.StringValue
}

func (v ipProtocol) Equal(o attr.Value) bool {
	other, ok := o.(ipProtocol)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (ipProtocol) Type(context.Context) attr.Type {
	return ipProtocolType{}
}

func (v ipProtocol) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(ipProtocol)
	if !ok {
		return false, diags
	}

	return protocolForValue(newValue.ValueString()) == protocolForValue(v.ValueString()), diags
}

func ipProtocolNull() ipProtocol {
	return ipProtocol{StringValue: basetypes.NewStringNull()}
}

func ipProtocolUnknown() ipProtocol {
	return ipProtocol{StringValue: basetypes.NewStringUnknown()}
}

func ipProtocolValue(value string) ipProtocol {
	return ipProtocol{StringValue: basetypes.NewStringValue(value)}
}

// legacySecurityGroupRuleResourceSchemaV2 returns version 2 of the schema for the `aws_security_group_rule` resource.
func legacySecurityGroupRuleResourceSchemaV2(ctx context.Context) *schema.Schema {
	return &schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cidr_blocks": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"from_port": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"ipv6_cidr_blocks": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"prefix_list_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			names.AttrProtocol: schema.StringAttribute{
				CustomType: ipProtocolType{},
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"security_group_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"security_group_rule_id": schema.StringAttribute{
				Computed: true,
			},
			"self": schema.BoolAttribute{
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"source_security_group_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"to_port": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[securityGroupRuleType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
		Version: 2,
	}
}

type legacySecurityGroupRuleResourceModel struct {
	CIDRBlocks            fwtypes.ListValueOf[types.String]         `tfsdk:"cidr_blocks"`
	Description           types.String                              `tfsdk:"description"`
	FromPort              types.Int64                               `tfsdk:"from_port"`
	ID                    types.String                              `tfsdk:"id"`
	IPv6CIDRBlocksBlocks  fwtypes.ListValueOf[types.String]         `tfsdk:"ipv6_cidr_blocks"`
	PrefixListIDs         fwtypes.ListValueOf[types.String]         `tfsdk:"prefix_list_ids"`
	Protocol              ipProtocol                                `tfsdk:"protocol"`
	SecurityGroupID       types.String                              `tfsdk:"security_group_id"`
	SecurityGroupRuleID   types.String                              `tfsdk:"security_group_rule_id"`
	Self                  types.Bool                                `tfsdk:"self"`
	SourceSecurityGroupID types.String                              `tfsdk:"source_security_group_id"`
	Timeouts              timeouts.Value                            `tfsdk:"timeouts"`
	ToPort                types.Int64                               `tfsdk:"to_port"`
	Type                  fwtypes.StringEnum[securityGroupRuleType] `tfsdk:"type"`
}
