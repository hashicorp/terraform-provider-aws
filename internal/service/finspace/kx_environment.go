// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_finspace_kx_environment", name="Kx Environment")
// @Tags(identifierAttribute="arn")
func ResourceKxEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKxEnvironmentCreate,
		ReadWithoutTimeout:   resourceKxEnvironmentRead,
		UpdateWithoutTimeout: resourceKxEnvironmentUpdate,
		DeleteWithoutTimeout: resourceKxEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(75 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"created_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_dns_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_dns_server_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(3, 255),
						},
						"custom_dns_server_ip": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"infrastructure_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"last_modified_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"transit_gateway_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attachment_network_acl_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 100,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrCIDRBlock: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.IsCIDR,
									},
									"icmp_type_code": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrType: {
													Type:     schema.TypeInt,
													Required: true,
												},
												"code": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									"port_range": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"from": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IsPortNumber,
												},
												"to": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IsPortNumber,
												},
											},
										},
									},
									names.AttrProtocol: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 5),
									},
									"rule_action": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.RuleAction](),
									},
									"rule_number": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 32766),
									},
								},
							},
						},
						"routable_cidr_space": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
						names.AttrTransitGatewayID: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 32),
						},
					},
				},
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameKxEnvironment = "Kx Environment"
)

func resourceKxEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	in := &finspace.CreateKxEnvironmentInput{
		Name:        aws.String(d.Get(names.AttrName).(string)),
		ClientToken: aws.String(id.UniqueId()),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		in.KmsKeyId = aws.String(v.(string))
	}

	out, err := conn.CreateKxEnvironment(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxEnvironment, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.EnvironmentId == nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxEnvironment, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.EnvironmentId))

	if _, err := waitKxEnvironmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForCreation, ResNameKxEnvironment, d.Id(), err)
	}

	if err := updateKxEnvironmentNetwork(ctx, d, conn); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxEnvironment, d.Id(), err)
	}

	// The CreateKxEnvironment API currently fails to tag the environment when the
	// Tags field is set. Until the API is fixed, tag after creation instead.
	if err := createTags(ctx, conn, aws.ToString(out.EnvironmentArn), getTagsIn(ctx)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxEnvironment, d.Id(), err)
	}

	return append(diags, resourceKxEnvironmentRead(ctx, d, meta)...)
}

func resourceKxEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	out, err := findKxEnvironmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FinSpace KxEnvironment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionReading, ResNameKxEnvironment, d.Id(), err)
	}

	d.Set(names.AttrID, out.EnvironmentId)
	d.Set(names.AttrARN, out.EnvironmentArn)
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrDescription, out.Description)
	d.Set(names.AttrKMSKeyID, out.KmsKeyId)
	d.Set(names.AttrStatus, out.Status)
	d.Set(names.AttrAvailabilityZones, out.AvailabilityZoneIds)
	d.Set("infrastructure_account_id", out.DedicatedServiceAccountId)
	d.Set("created_timestamp", out.CreationTimestamp.String())
	d.Set("last_modified_timestamp", out.UpdateTimestamp.String())

	if err := d.Set("transit_gateway_configuration", flattenTransitGatewayConfiguration(out.TransitGatewayConfiguration)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionSetting, ResNameKxEnvironment, d.Id(), err)
	}

	if err := d.Set("custom_dns_configuration", flattenCustomDNSConfigurations(out.CustomDNSConfiguration)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionSetting, ResNameKxEnvironment, d.Id(), err)
	}

	return diags
}

func resourceKxEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	update := false

	in := &finspace.UpdateKxEnvironmentInput{
		EnvironmentId: aws.String(d.Id()),
		Name:          aws.String(d.Get(names.AttrName).(string)),
	}

	if d.HasChanges(names.AttrDescription) {
		in.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChanges(names.AttrName) || d.HasChanges(names.AttrDescription) {
		update = true
		log.Printf("[DEBUG] Updating FinSpace KxEnvironment (%s): %#v", d.Id(), in)
		_, err := conn.UpdateKxEnvironment(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.FinSpace, create.ErrActionUpdating, ResNameKxEnvironment, d.Id(), err)
		}
	}

	if d.HasChanges("transit_gateway_configuration") || d.HasChanges("custom_dns_configuration") {
		update = true
		if err := updateKxEnvironmentNetwork(ctx, d, conn); err != nil {
			return create.AppendDiagError(diags, names.FinSpace, create.ErrActionUpdating, ResNameKxEnvironment, d.Id(), err)
		}
	}

	if !update {
		return diags
	}
	return append(diags, resourceKxEnvironmentRead(ctx, d, meta)...)
}

func resourceKxEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	log.Printf("[INFO] Deleting FinSpace Kx Environment: %s", d.Id())
	_, err := conn.DeleteKxEnvironment(ctx, &finspace.DeleteKxEnvironmentInput{
		EnvironmentId: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) ||
		errs.IsAErrorMessageContains[*types.ValidationException](err, "The Environment is in DELETED state") {
		log.Printf("[DEBUG] FinSpace KxEnvironment %s already deleted. Nothing to delete.", d.Id())
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionDeleting, ResNameKxEnvironment, d.Id(), err)
	}

	if _, err := waitKxEnvironmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForDeletion, ResNameKxEnvironment, d.Id(), err)
	}

	return diags
}

// As of 2023-02-09, updating network configuration requires 2 separate requests if both DNS
// and transit gateway configurationtions are set.
func updateKxEnvironmentNetwork(ctx context.Context, d *schema.ResourceData, client *finspace.Client) error {
	transitGatewayConfigIn := &finspace.UpdateKxEnvironmentNetworkInput{
		EnvironmentId: aws.String(d.Id()),
		ClientToken:   aws.String(id.UniqueId()),
	}

	customDnsConfigIn := &finspace.UpdateKxEnvironmentNetworkInput{
		EnvironmentId: aws.String(d.Id()),
		ClientToken:   aws.String(id.UniqueId()),
	}

	updateTransitGatewayConfig := false
	updateCustomDnsConfig := false

	if v, ok := d.GetOk("transit_gateway_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil &&
		d.HasChanges("transit_gateway_configuration") {
		transitGatewayConfigIn.TransitGatewayConfiguration = expandTransitGatewayConfiguration(v.([]interface{}))
		updateTransitGatewayConfig = true
	}

	if v, ok := d.GetOk("custom_dns_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil &&
		d.HasChanges("custom_dns_configuration") {
		customDnsConfigIn.CustomDNSConfiguration = expandCustomDNSConfigurations(v.([]interface{}))
		updateCustomDnsConfig = true
	}

	if updateTransitGatewayConfig {
		if _, err := client.UpdateKxEnvironmentNetwork(ctx, transitGatewayConfigIn); err != nil {
			return err
		}

		if _, err := waitTransitGatewayConfigurationUpdated(ctx, client, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return err
		}
	}

	if updateCustomDnsConfig {
		if _, err := client.UpdateKxEnvironmentNetwork(ctx, customDnsConfigIn); err != nil {
			return err
		}

		if _, err := waitCustomDNSConfigurationUpdated(ctx, client, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return err
		}
	}

	return nil
}

func waitKxEnvironmentCreated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.EnvironmentStatusCreateRequested, types.EnvironmentStatusCreating),
		Target:                    enum.Slice(types.EnvironmentStatusCreated),
		Refresh:                   statusKxEnvironment(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func waitTransitGatewayConfigurationUpdated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.TgwStatusUpdateRequested, types.TgwStatusUpdating),
		Target:  enum.Slice(types.TgwStatusSuccessfullyUpdated),
		Refresh: statusTransitGatewayConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func waitCustomDNSConfigurationUpdated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DnsStatusUpdateRequested, types.DnsStatusUpdating),
		Target:  enum.Slice(types.DnsStatusSuccessfullyUpdated),
		Refresh: statusCustomDNSConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func waitKxEnvironmentDeleted(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EnvironmentStatusDeleteRequested, types.EnvironmentStatusDeleting),
		Target:  []string{},
		Refresh: statusKxEnvironment(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func statusKxEnvironment(ctx context.Context, conn *finspace.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findKxEnvironmentByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func statusTransitGatewayConfiguration(ctx context.Context, conn *finspace.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findKxEnvironmentByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.TgwStatus), nil
	}
}

func statusCustomDNSConfiguration(ctx context.Context, conn *finspace.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findKxEnvironmentByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.DnsStatus), nil
	}
}

func findKxEnvironmentByID(ctx context.Context, conn *finspace.Client, id string) (*finspace.GetKxEnvironmentOutput, error) {
	in := &finspace.GetKxEnvironmentInput{
		EnvironmentId: aws.String(id),
	}
	out, err := conn.GetKxEnvironment(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}
	// Treat DELETED status as NotFound
	if out != nil && out.Status == types.EnvironmentStatusDeleted {
		return nil, &retry.NotFoundError{
			LastError:   errors.New("status is deleted"),
			LastRequest: in,
		}
	}

	if out == nil || out.EnvironmentArn == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func expandTransitGatewayConfiguration(tfList []interface{}) *types.TransitGatewayConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.TransitGatewayConfiguration{}

	if v, ok := tfMap[names.AttrTransitGatewayID].(string); ok && v != "" {
		a.TransitGatewayID = aws.String(v)
	}

	if v, ok := tfMap["routable_cidr_space"].(string); ok && v != "" {
		a.RoutableCIDRSpace = aws.String(v)
	}

	if v, ok := tfMap["attachment_network_acl_configuration"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		a.AttachmentNetworkAclConfiguration = expandAttachmentNetworkACLConfigurations(v.([]interface{}))
	}

	return a
}

func expandAttachmentNetworkACLConfigurations(tfList []interface{}) []types.NetworkACLEntry {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.NetworkACLEntry
	for _, r := range tfList {
		m, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		a := expandAttachmentNetworkACLConfiguration(m)
		if a == nil {
			continue
		}

		s = append(s, *a)
	}
	return s
}

func expandAttachmentNetworkACLConfiguration(tfMap map[string]interface{}) *types.NetworkACLEntry {
	if tfMap == nil {
		return nil
	}

	a := &types.NetworkACLEntry{}
	if v, ok := tfMap["rule_number"].(int); ok && v > 0 {
		a.RuleNumber = aws.Int32(int32(v))
	}
	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		a.Protocol = &v
	}
	if v, ok := tfMap["rule_action"].(string); ok && v != "" {
		a.RuleAction = types.RuleAction(v)
	}
	if v, ok := tfMap[names.AttrCIDRBlock].(string); ok && v != "" {
		a.CidrBlock = &v
	}
	if v, ok := tfMap["port_range"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		a.PortRange = expandPortRange(v.([]interface{}))
	}
	if v, ok := tfMap["icmp_type_code"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		a.IcmpTypeCode = expandIcmpTypeCode(v.([]interface{}))
	}

	return a
}

func expandPortRange(tfList []interface{}) *types.PortRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap := tfList[0].(map[string]interface{})

	return &types.PortRange{
		From: int32(tfMap["from"].(int)),
		To:   int32(tfMap["to"].(int)),
	}
}

func expandIcmpTypeCode(tfList []interface{}) *types.IcmpTypeCode {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap := tfList[0].(map[string]interface{})

	return &types.IcmpTypeCode{
		Code: int32(tfMap["code"].(int)),
		Type: int32(tfMap[names.AttrType].(int)),
	}
}

func expandCustomDNSConfiguration(tfMap map[string]interface{}) *types.CustomDNSServer {
	if tfMap == nil {
		return nil
	}

	a := &types.CustomDNSServer{}

	if v, ok := tfMap["custom_dns_server_name"].(string); ok && v != "" {
		a.CustomDNSServerName = aws.String(v)
	}

	if v, ok := tfMap["custom_dns_server_ip"].(string); ok && v != "" {
		a.CustomDNSServerIP = aws.String(v)
	}

	return a
}

func expandCustomDNSConfigurations(tfList []interface{}) []types.CustomDNSServer {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.CustomDNSServer

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandCustomDNSConfiguration(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func flattenTransitGatewayConfiguration(apiObject *types.TransitGatewayConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.TransitGatewayID; v != nil {
		m[names.AttrTransitGatewayID] = aws.ToString(v)
	}

	if v := apiObject.RoutableCIDRSpace; v != nil {
		m["routable_cidr_space"] = aws.ToString(v)
	}

	if v := apiObject.AttachmentNetworkAclConfiguration; v != nil {
		m["attachment_network_acl_configuration"] = flattenAttachmentNetworkACLConfigurations(v)
	}

	return []interface{}{m}
}

func flattenAttachmentNetworkACLConfigurations(apiObjects []types.NetworkACLEntry) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenAttachmentNetworkACLConfiguration(&apiObject))
	}

	return l
}

func flattenAttachmentNetworkACLConfiguration(apiObject *types.NetworkACLEntry) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrCIDRBlock: aws.ToString(apiObject.CidrBlock),
		names.AttrProtocol:  aws.ToString(apiObject.Protocol),
		"rule_action":       apiObject.RuleAction,
		"rule_number":       apiObject.RuleNumber,
	}

	if v := apiObject.PortRange; v != nil {
		m["port_range"] = flattenPortRange(v)
	}
	if v := apiObject.IcmpTypeCode; v != nil {
		m["icmp_type_code"] = flattenIcmpTypeCode(v)
	}

	return m
}

func flattenPortRange(apiObject *types.PortRange) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"from": apiObject.From,
		"to":   apiObject.To,
	}

	return []interface{}{m}
}

func flattenIcmpTypeCode(apiObject *types.IcmpTypeCode) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrType: apiObject.Type,
		"code":         apiObject.Code,
	}

	return []interface{}{m}
}

func flattenCustomDNSConfiguration(apiObject *types.CustomDNSServer) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.CustomDNSServerName; v != nil {
		m["custom_dns_server_name"] = aws.ToString(v)
	}

	if v := apiObject.CustomDNSServerIP; v != nil {
		m["custom_dns_server_ip"] = aws.ToString(v)
	}

	return m
}

func flattenCustomDNSConfigurations(apiObjects []types.CustomDNSServer) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenCustomDNSConfiguration(&apiObject))
	}

	return l
}
