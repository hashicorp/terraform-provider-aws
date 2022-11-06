package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/intf"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	registerFrameworkResourceFactory(newResourceSecurityGroupIngressRule)
}

// newResourceSecurityGroupIngressRule instantiates a new Resource for the aws_vpc_security_group_ingress_rule resource.
func newResourceSecurityGroupIngressRule(context.Context) (intf.ResourceWithConfigureAndImportState, error) {
	return &resourceSecurityGroupIngressRule{}, nil
}

type resourceSecurityGroupIngressRule struct {
	meta *conns.AWSClient
}

// Metadata should return the full name of the resource, such as
// examplecloud_thing.
func (r *resourceSecurityGroupIngressRule) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_vpc_security_group_ingress_rule"
}

// GetSchema returns the schema for this resource.
func (r *resourceSecurityGroupIngressRule) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"arn": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"cidr_ipv4": {
				Type:     types.StringType,
				Optional: true,
			},
			"cidr_ipv6": {
				Type:     types.StringType,
				Optional: true,
			},
			"description": {
				Type:     types.StringType,
				Optional: true,
			},
			"from_port": {
				Type:     types.Int64Type,
				Optional: true,
			},
			"id": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"ip_protocol": {
				Type:     types.StringType,
				Required: true,
			},
			"security_group_id": {
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			"tags":     tftags.TagsAttribute(),
			"tags_all": tftags.TagsAttributeComputedOnly(),
			"to_port": {
				Type:     types.Int64Type,
				Optional: true,
			},
		},
	}

	return schema, nil
}

// Configure enables provider-level data or clients to be set in the
// provider-defined Resource type.
func (r *resourceSecurityGroupIngressRule) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		r.meta = v
	}
}

// Create is called when the provider must create a new resource.
// Config and planned state values should be read from the CreateRequest and new state values set on the CreateResponse.
func (r *resourceSecurityGroupIngressRule) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceSecurityGroupIngressRuleData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.meta.EC2Conn
	defaultTagsConfig := r.meta.DefaultTagsConfig
	ignoreTagsConfig := r.meta.IgnoreTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(data.Tags))

	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       aws.String(data.SecurityGroupID.Value),
		IpPermissions: []*ec2.IpPermission{r.expandIPPermission(ctx, &data)},
	}

	output, err := conn.AuthorizeSecurityGroupIngressWithContext(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating EC2 Security Group Ingress Rule", err.Error())

		return
	}

	data.ID = types.String{Value: aws.StringValue(output.SecurityGroupRules[0].SecurityGroupRuleId)}

	if len(tags) > 0 {
		if err := UpdateTagsWithContext(ctx, conn, data.ID.Value, nil, tags); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("adding EC2 Security Group Ingress Rule (%s) tags", data.ID.Value), err.Error())

			return
		}
	}

	// Set values for unknowns.
	data.ARN = r.arn(ctx, data.ID.Value)
	data.TagsAll = flex.FlattenFrameworkStringValueMap(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Read is called when the provider must read resource values in order to update state.
// Planned state values should be read from the ReadRequest and new state values set on the ReadResponse.
func (r *resourceSecurityGroupIngressRule) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceSecurityGroupIngressRuleData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.meta.EC2Conn
	defaultTagsConfig := r.meta.DefaultTagsConfig
	ignoreTagsConfig := r.meta.IgnoreTagsConfig

	output, err := FindSecurityGroupIngressRuleByID(ctx, conn, data.ID.Value)

	if tfresource.NotFound(err) {
		tflog.Warn(ctx, "EC2 Security Group Ingress Rule not found, removing from state", map[string]interface{}{
			"id": data.ID.Value,
		})
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Security Group Ingress Rule (%s)", data.ID.Value), err.Error())

		return
	}

	data.ARN = r.arn(ctx, data.ID.Value)
	data.CIDRIPv4 = flex.ToFrameworkStringValue(ctx, output.CidrIpv4)
	data.CIDRIPv6 = flex.ToFrameworkStringValue(ctx, output.CidrIpv6)
	data.Description = flex.ToFrameworkStringValue(ctx, output.Description)
	data.FromPort = flex.ToFrameworkInt64Value(ctx, output.FromPort)
	data.IPProtocol = flex.ToFrameworkStringValue(ctx, output.IpProtocol)
	data.SecurityGroupID = flex.ToFrameworkStringValue(ctx, output.GroupId)
	data.ToPort = flex.ToFrameworkInt64Value(ctx, output.ToPort)

	// If planned tags are null and no tags are returned, propagate null.
	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	if tags := tags.RemoveDefaultConfig(defaultTagsConfig).Map(); len(tags) == 0 && data.Tags.IsNull() {
		data.Tags = types.Map{ElemType: types.StringType, Null: true}
	} else {
		data.Tags = flex.FlattenFrameworkStringValueMap(ctx, tags)
	}
	data.TagsAll = flex.FlattenFrameworkStringValueMap(ctx, tags.Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Update is called to update the state of the resource.
// Config, planned state, and prior state values should be read from the UpdateRequest and new state values set on the UpdateResponse.
func (r *resourceSecurityGroupIngressRule) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceSecurityGroupIngressRuleData

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.meta.EC2Conn

	if !new.CIDRIPv4.Equal(old.CIDRIPv4) ||
		!new.CIDRIPv6.Equal(old.CIDRIPv6) ||
		!new.Description.Equal(old.Description) ||
		!new.FromPort.Equal(old.FromPort) ||
		!new.IPProtocol.Equal(old.IPProtocol) ||
		!new.ToPort.Equal(old.ToPort) {
		input := &ec2.ModifySecurityGroupRulesInput{
			SecurityGroupRules: []*ec2.SecurityGroupRuleUpdate{{
				SecurityGroupRule:   r.expandSecurityGroupRuleRequest(ctx, &new),
				SecurityGroupRuleId: aws.String(new.ID.Value),
			}},
		}

		_, err := conn.ModifySecurityGroupRulesWithContext(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating EC2 Security Group Ingress Rule (%s)", new.ID.Value), err.Error())

			return
		}
	}

	if !new.TagsAll.Equal(old.TagsAll) {
		if err := UpdateTagsWithContext(ctx, conn, new.ID.Value, old.TagsAll, new.TagsAll); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating EC2 Security Group Ingress Rule (%s) tags", new.ID.Value), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

// Delete is called when the provider must delete the resource.
// Config values may be read from the DeleteRequest.
//
// If execution completes without error, the framework will automatically call DeleteResponse.State.RemoveResource(),
// so it can be omitted from provider logic.
func (r *resourceSecurityGroupIngressRule) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceSecurityGroupIngressRuleData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.meta.EC2Conn

	tflog.Debug(ctx, "deleting EC2 Security Group Ingress Rule", map[string]interface{}{
		"id": data.ID.Value,
	})
	_, err := conn.RevokeSecurityGroupIngressWithContext(ctx, &ec2.RevokeSecurityGroupIngressInput{
		GroupId:              aws.String(data.SecurityGroupID.Value),
		SecurityGroupRuleIds: aws.StringSlice([]string{data.ID.Value}),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound, errCodeInvalidSecurityGroupRuleIdNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Security Group Ingress Rule (%s)", data.ID.Value), err.Error())

		return
	}
}

// ImportState is called when the provider must import the state of a resource instance.
// This method must return enough state so the Read method can properly refresh the full resource.
//
// If setting an attribute with the import identifier, it is recommended to use the ImportStatePassthroughID() call in this method.
func (r *resourceSecurityGroupIngressRule) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// ModifyPlan is called when the provider has an opportunity to modify
// the plan: once during the plan phase when Terraform is determining
// the diff that should be shown to the user for approval, and once
// during the apply phase with any unknown values from configuration
// filled in with their final values.
//
// The planned new state is represented by
// ModifyPlanResponse.Plan. It must meet the following
// constraints:
// 1. Any non-Computed attribute set in config must preserve the exact
// config value or return the corresponding attribute value from the
// prior state (ModifyPlanRequest.State).
// 2. Any attribute with a known value must not have its value changed
// in subsequent calls to ModifyPlan or Create/Read/Update.
// 3. Any attribute with an unknown value may either remain unknown
// or take on any value of the expected type.
//
// Any errors will prevent further resource-level plan modifications.
func (r *resourceSecurityGroupIngressRule) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	defaultTagsConfig := r.meta.DefaultTagsConfig
	ignoreTagsConfig := r.meta.IgnoreTagsConfig

	var planTags types.Map

	response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root("tags"), &planTags)...)

	if response.Diagnostics.HasError() {
		return
	}

	resourceTags := tftags.New(planTags)

	if defaultTagsConfig.TagsEqual(resourceTags) {
		response.Diagnostics.AddError(
			`"tags" are identical to those in the "default_tags" configuration block of the provider`,
			"please de-duplicate and try again")
	}

	allTags := defaultTagsConfig.MergeTags(resourceTags).IgnoreConfig(ignoreTagsConfig)

	response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root("tags_all"), flex.FlattenFrameworkStringValueMap(ctx, allTags.Map()))...)
}

func (r *resourceSecurityGroupIngressRule) arn(_ context.Context, id string) types.String {
	arn := arn.ARN{
		Partition: r.meta.Partition,
		Service:   ec2.ServiceName,
		Region:    r.meta.Region,
		AccountID: r.meta.AccountID,
		Resource:  fmt.Sprintf("security-group-rule/%s", id),
	}.String()
	return types.String{Value: arn}
}

func (r *resourceSecurityGroupIngressRule) expandIPPermission(_ context.Context, data *resourceSecurityGroupIngressRuleData) *ec2.IpPermission {
	apiObject := &ec2.IpPermission{}

	if !data.CIDRIPv4.IsNull() {
		apiObject.IpRanges = []*ec2.IpRange{{
			CidrIp: aws.String(data.CIDRIPv4.Value),
		}}

		if !data.Description.IsNull() {
			apiObject.IpRanges[0].Description = aws.String(data.Description.Value)
		}
	}

	if !data.CIDRIPv6.IsNull() {
		apiObject.Ipv6Ranges = []*ec2.Ipv6Range{{
			CidrIpv6: aws.String(data.CIDRIPv6.Value),
		}}

		if !data.Description.IsNull() {
			apiObject.IpRanges[0].Description = aws.String(data.Description.Value)
		}
	}

	if !data.FromPort.IsNull() {
		apiObject.FromPort = aws.Int64(data.FromPort.Value)
	}

	if !data.IPProtocol.IsNull() {
		apiObject.IpProtocol = aws.String(data.IPProtocol.Value)
	}

	if !data.ToPort.IsNull() {
		apiObject.ToPort = aws.Int64(data.ToPort.Value)
	}

	return apiObject
}

func (r *resourceSecurityGroupIngressRule) expandSecurityGroupRuleRequest(_ context.Context, data *resourceSecurityGroupIngressRuleData) *ec2.SecurityGroupRuleRequest {
	apiObject := &ec2.SecurityGroupRuleRequest{}

	if !data.CIDRIPv4.IsNull() {
		apiObject.CidrIpv4 = aws.String(data.CIDRIPv4.Value)
	}

	if !data.CIDRIPv6.IsNull() {
		apiObject.CidrIpv6 = aws.String(data.CIDRIPv6.Value)
	}

	if !data.Description.IsNull() {
		apiObject.Description = aws.String(data.Description.Value)
	}

	if !data.FromPort.IsNull() {
		apiObject.FromPort = aws.Int64(data.FromPort.Value)
	}

	if !data.IPProtocol.IsNull() {
		apiObject.IpProtocol = aws.String(data.IPProtocol.Value)
	}

	if !data.ToPort.IsNull() {
		apiObject.ToPort = aws.Int64(data.ToPort.Value)
	}

	return apiObject
}

type resourceSecurityGroupIngressRuleData struct {
	ARN             types.String `tfsdk:"arn"`
	CIDRIPv4        types.String `tfsdk:"cidr_ipv4"`
	CIDRIPv6        types.String `tfsdk:"cidr_ipv6"`
	Description     types.String `tfsdk:"description"`
	FromPort        types.Int64  `tfsdk:"from_port"`
	ID              types.String `tfsdk:"id"`
	IPProtocol      types.String `tfsdk:"ip_protocol"`
	SecurityGroupID types.String `tfsdk:"security_group_id"`
	Tags            types.Map    `tfsdk:"tags"`
	TagsAll         types.Map    `tfsdk:"tags_all"`
	ToPort          types.Int64  `tfsdk:"to_port"`
}

// TODO
// * PrefixListId
// * ReferencedGroupId
// * Ensure at least one "target" is specified
// * ForceNew if target type changes
// * All protocol => No FromPort/ToPort
// Add security_group_rule_id attribute; ID = SGID/SGRID
