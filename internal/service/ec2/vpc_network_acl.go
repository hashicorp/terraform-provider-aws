// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_network_acl", name="Network ACL")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceNetworkACL() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkACLCreate,
		ReadWithoutTimeout:   resourceNetworkACLRead,
		UpdateWithoutTimeout: resourceNetworkACLUpdate,
		DeleteWithoutTimeout: resourceNetworkACLDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				conn := meta.(*conns.AWSClient).EC2Client(ctx)

				nacl, err := findNetworkACLByID(ctx, conn, d.Id())

				if err != nil {
					return nil, err
				}

				if aws.ToBool(nacl.IsDefault) {
					return nil, fmt.Errorf("use the `aws_default_network_acl` resource instead")
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		// Keep in sync with aws_default_network_acl's schema.
		// See notes in default_network_acl.go.
		SchemaFunc: func() map[string]*schema.Schema {
			networkACLRuleSetNestedBlock := func() *schema.Schema {
				return &schema.Schema{
					Type:       schema.TypeSet,
					Optional:   true,
					Computed:   true,
					ConfigMode: schema.SchemaConfigModeAttr,
					Elem:       networkACLRuleNestedBlock(),
					Set:        networkACLRuleHash,
				}
			}

			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"egress":  networkACLRuleSetNestedBlock(),
				"ingress": networkACLRuleSetNestedBlock(),
				names.AttrOwnerID: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrSubnetIDs: {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				names.AttrVPCID: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

// NACL rule nested block definition.
// Used in aws_network_acl and aws_default_network_acl ingress and egress rule sets.
func networkACLRuleNestedBlock() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrAction: {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
				// Accept pascal case for backwards compatibility reasons, See: TestAccVPCNetworkACL_caseSensitivityNoChanges
				ValidateFunc: validation.StringInSlice(enum.Slice(awstypes.RuleAction.Values("")...), true),
			},
			names.AttrCIDRBlock: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
			},
			"from_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumberOrZero,
			},
			"icmp_code": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"icmp_type": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ipv6_cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidIPv6CIDRNetworkAddress,
			},
			names.AttrProtocol: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					_, err := networkACLProtocolNumber(v.(string))

					if err != nil {
						errors = append(errors, fmt.Errorf("%q : %w", k, err))
					}

					return
				},
			},
			"rule_no": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 32766),
			},
			"to_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumberOrZero,
			},
		},
	}
}

func resourceNetworkACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateNetworkAclInput{
		ClientToken:       aws.String(id.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeNetworkAcl),
		VpcId:             aws.String(d.Get(names.AttrVPCID).(string)),
	}

	output, err := conn.CreateNetworkAcl(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Network ACL: %s", err)
	}

	d.SetId(aws.ToString(output.NetworkAcl.NetworkAclId))

	if err := modifyNetworkACLAttributesOnCreate(ctx, conn, d); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Network ACL: %s", err)
	}

	return append(diags, resourceNetworkACLRead(ctx, d, meta)...)
}

func resourceNetworkACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findNetworkACLByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network ACL %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network ACL (%s): %s", d.Id(), err)
	}

	nacl := outputRaw.(*awstypes.NetworkAcl)

	ownerID := aws.ToString(nacl.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("network-acl/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrOwnerID, ownerID)

	var subnetIDs []string
	for _, v := range nacl.Associations {
		subnetIDs = append(subnetIDs, aws.ToString(v.SubnetId))
	}
	d.Set(names.AttrSubnetIDs, subnetIDs)

	d.Set(names.AttrVPCID, nacl.VpcId)

	var egressEntries []awstypes.NetworkAclEntry
	var ingressEntries []awstypes.NetworkAclEntry
	for _, v := range nacl.Entries {
		// Skip the default rules added by AWS. They can be neither
		// configured or deleted by users.
		if v := aws.ToInt32(v.RuleNumber); v == defaultACLRuleNumberIPv4 || v == defaultACLRuleNumberIPv6 {
			continue
		}

		if aws.ToBool(v.Egress) {
			egressEntries = append(egressEntries, v)
		} else {
			ingressEntries = append(ingressEntries, v)
		}
	}
	if err := d.Set("egress", flattenNetworkACLEntries(egressEntries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting egress: %s", err)
	}
	if err := d.Set("ingress", flattenNetworkACLEntries(ingressEntries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ingress: %s", err)
	}

	setTagsOut(ctx, nacl.Tags)

	return diags
}

func resourceNetworkACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if err := modifyNetworkACLAttributesOnUpdate(ctx, conn, d, true); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EC2 Network ACL (%s): %s", d.Id(), err)
	}

	return append(diags, resourceNetworkACLRead(ctx, d, meta)...)
}

func resourceNetworkACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// Delete all NACL/Subnet associations, even if they are managed via aws_network_acl_association resources.
	nacl, err := findNetworkACLByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network ACL (%s): %s", d.Id(), err)
	}

	var subnetIDs []interface{}
	for _, v := range nacl.Associations {
		subnetIDs = append(subnetIDs, aws.ToString(v.SubnetId))
	}
	if len(subnetIDs) > 0 {
		if err := networkACLAssociationsDelete(ctx, conn, d.Get(names.AttrVPCID).(string), subnetIDs); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting EC2 Network ACL (%s): %s", d.Id(), err)
		}
	}

	input := &ec2.DeleteNetworkAclInput{
		NetworkAclId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting EC2 Network ACL: %s", d.Id())
	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return conn.DeleteNetworkAcl(ctx, input)
	}, errCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkACLIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Network ACL (%s): %s", d.Id(), err)
	}

	return diags
}

// modifyNetworkACLAttributesOnCreate sets NACL attributes on resource Create.
// Called after new NACL creation or existing default NACL adoption.
// Tags are not configured.
func modifyNetworkACLAttributesOnCreate(ctx context.Context, conn *ec2.Client, d *schema.ResourceData) error {
	if v, ok := d.GetOk("egress"); ok && v.(*schema.Set).Len() > 0 {
		if err := createNetworkACLEntries(ctx, conn, d.Id(), v.(*schema.Set).List(), true); err != nil {
			return err
		}
	}

	if v, ok := d.GetOk("ingress"); ok && v.(*schema.Set).Len() > 0 {
		if err := createNetworkACLEntries(ctx, conn, d.Id(), v.(*schema.Set).List(), false); err != nil {
			return err
		}
	}

	if v, ok := d.GetOk(names.AttrSubnetIDs); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			if _, err := networkACLAssociationCreate(ctx, conn, d.Id(), v.(string)); err != nil {
				return err
			}
		}
	}

	return nil
}

// modifyNetworkACLAttributesOnUpdate sets NACL attributes on resource Update.
// Tags are configured.
func modifyNetworkACLAttributesOnUpdate(ctx context.Context, conn *ec2.Client, d *schema.ResourceData, deleteAssociations bool) error {
	if d.HasChange("ingress") {
		o, n := d.GetChange("ingress")
		os, ns := o.(*schema.Set), n.(*schema.Set)

		if err := updateNetworkACLEntries(ctx, conn, d.Id(), os, ns, false); err != nil {
			return err
		}
	}

	if d.HasChange("egress") {
		o, n := d.GetChange("egress")
		os, ns := o.(*schema.Set), n.(*schema.Set)

		if err := updateNetworkACLEntries(ctx, conn, d.Id(), os, ns, true); err != nil {
			return err
		}
	}

	if d.HasChange(names.AttrSubnetIDs) {
		o, n := d.GetChange(names.AttrSubnetIDs)
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := ns.Difference(os).List(), os.Difference(ns).List()

		if len(del) > 0 && deleteAssociations {
			if err := networkACLAssociationsDelete(ctx, conn, d.Get(names.AttrVPCID).(string), del); err != nil {
				return err
			}
		}

		if len(add) > 0 {
			if err := networkACLAssociationsCreate(ctx, conn, d.Id(), add); err != nil {
				return err
			}
		}
	}

	return nil
}

func networkACLRuleHash(v interface{}) int {
	var buf bytes.Buffer

	tfMap := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", tfMap["from_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", tfMap["to_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", tfMap["rule_no"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(tfMap[names.AttrAction].(string))))

	// The AWS network ACL API only speaks protocol numbers, and that's
	// all we store. Never hash a protocol name.
	protocolNumber, _ := networkACLProtocolNumber(tfMap[names.AttrProtocol].(string))
	buf.WriteString(fmt.Sprintf("%d-", protocolNumber))

	if v, ok := tfMap[names.AttrCIDRBlock]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := tfMap["ipv6_cidr_block"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := tfMap["icmp_type"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := tfMap["icmp_code"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}

	return create.StringHashcode(buf.String())
}

func createNetworkACLEntries(ctx context.Context, conn *ec2.Client, naclID string, tfList []interface{}, egress bool) error {
	naclEntries := expandNetworkACLEntries(tfList, egress)

	for _, naclEntry := range naclEntries {
		if aws.ToString(naclEntry.Protocol) == "-1" {
			// Protocol -1 rules don't store ports in AWS. Thus, they'll always
			// hash differently when being read out of the API. Force the user
			// to set from_port and to_port to 0 for these rules, to keep the
			// hashing consistent.
			if from, to := aws.ToInt32(naclEntry.PortRange.From), aws.ToInt32(naclEntry.PortRange.To); from != 0 || to != 0 {
				return fmt.Errorf("to_port (%d) and from_port (%d) must both be 0 to use the 'all' \"-1\" protocol!", to, from)
			}
		}

		input := &ec2.CreateNetworkAclEntryInput{
			CidrBlock:     naclEntry.CidrBlock,
			Egress:        naclEntry.Egress,
			IcmpTypeCode:  naclEntry.IcmpTypeCode,
			Ipv6CidrBlock: naclEntry.Ipv6CidrBlock,
			NetworkAclId:  aws.String(naclID),
			PortRange:     naclEntry.PortRange,
			Protocol:      naclEntry.Protocol,
			RuleAction:    naclEntry.RuleAction,
			RuleNumber:    naclEntry.RuleNumber,
		}

		log.Printf("[INFO] Creating EC2 Network ACL Entry: %#v", input)
		_, err := conn.CreateNetworkAclEntry(ctx, input)

		if err != nil {
			return fmt.Errorf("creating EC2 Network ACL (%s) Entry: %w", naclID, err)
		}
	}

	return nil
}

func deleteNetworkACLEntriesList(ctx context.Context, conn *ec2.Client, naclID string, tfList []interface{}, egress bool) error {
	return deleteNetworkACLEntries(ctx, conn, naclID, expandNetworkACLEntries(tfList, egress))
}

func deleteNetworkACLEntries(ctx context.Context, conn *ec2.Client, naclID string, naclEntries []awstypes.NetworkAclEntry) error {
	for _, naclEntry := range naclEntries {
		// AWS includes default rules with all network ACLs that can be
		// neither modified nor destroyed. They have a custom rule
		// number that is out of bounds for any other rule. If we
		// encounter it, just continue. There's no work to be done.
		if v := aws.ToInt32(naclEntry.RuleNumber); v == defaultACLRuleNumberIPv4 || v == defaultACLRuleNumberIPv6 {
			continue
		}

		input := &ec2.DeleteNetworkAclEntryInput{
			Egress:       naclEntry.Egress,
			NetworkAclId: aws.String(naclID),
			RuleNumber:   naclEntry.RuleNumber,
		}

		log.Printf("[INFO] Deleting EC2 Network ACL Entry: %#v", input)
		_, err := conn.DeleteNetworkAclEntry(ctx, input)

		if err != nil {
			return fmt.Errorf("deleting EC2 Network ACL (%s) Entry: %w", naclID, err)
		}
	}

	return nil
}

func updateNetworkACLEntries(ctx context.Context, conn *ec2.Client, naclID string, os, ns *schema.Set, egress bool) error {
	if err := deleteNetworkACLEntriesList(ctx, conn, naclID, os.Difference(ns).List(), egress); err != nil {
		return err
	}

	if err := createNetworkACLEntries(ctx, conn, naclID, ns.Difference(os).List(), egress); err != nil {
		return err
	}

	return nil
}

func expandNetworkACLEntry(tfMap map[string]interface{}, egress bool) *awstypes.NetworkAclEntry {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NetworkAclEntry{
		Egress:    aws.Bool(egress),
		PortRange: &awstypes.PortRange{},
	}

	if v, ok := tfMap["rule_no"].(int); ok {
		apiObject.RuleNumber = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrAction].(string); ok && v != "" {
		apiObject.RuleAction = awstypes.RuleAction(v)
	}

	if v, ok := tfMap[names.AttrCIDRBlock].(string); ok && v != "" {
		apiObject.CidrBlock = aws.String(v)
	}

	if v, ok := tfMap["ipv6_cidr_block"].(string); ok && v != "" {
		apiObject.Ipv6CidrBlock = aws.String(v)
	}

	if v, ok := tfMap["from_port"].(int); ok {
		apiObject.PortRange.From = aws.Int32(int32(v))
	}

	if v, ok := tfMap["to_port"].(int); ok {
		apiObject.PortRange.To = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		protocolNumber, err := networkACLProtocolNumber(v)

		if err != nil {
			log.Printf("[WARN] %s", err)
			return nil
		}

		apiObject.Protocol = aws.String(strconv.Itoa(protocolNumber))

		// Specify additional required fields for ICMP.
		if protocolNumber == 1 || protocolNumber == 58 {
			apiObject.IcmpTypeCode = &awstypes.IcmpTypeCode{}

			if v, ok := tfMap["icmp_code"].(int); ok {
				apiObject.IcmpTypeCode.Code = aws.Int32(int32(v))
			}

			if v, ok := tfMap["icmp_type"].(int); ok {
				apiObject.IcmpTypeCode.Type = aws.Int32(int32(v))
			}
		}
	}

	return apiObject
}

func expandNetworkACLEntries(tfList []interface{}, egress bool) []awstypes.NetworkAclEntry {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.NetworkAclEntry

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandNetworkACLEntry(tfMap, egress)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenNetworkACLEntry(apiObject awstypes.NetworkAclEntry) map[string]interface{} {
	tfMap := map[string]interface{}{
		names.AttrAction: apiObject.RuleAction,
	}

	if v := apiObject.RuleNumber; v != nil {
		tfMap["rule_no"] = aws.ToInt32(v)
	}

	if v := apiObject.CidrBlock; v != nil {
		tfMap[names.AttrCIDRBlock] = aws.ToString(v)
	}

	if v := apiObject.Ipv6CidrBlock; v != nil {
		tfMap["ipv6_cidr_block"] = aws.ToString(v)
	}

	if apiObject := apiObject.PortRange; apiObject != nil {
		if v := apiObject.From; v != nil {
			tfMap["from_port"] = aws.ToInt32(v)
		}

		if v := apiObject.To; v != nil {
			tfMap["to_port"] = aws.ToInt32(v)
		}
	}

	if v := aws.ToString(apiObject.Protocol); v != "" {
		// The AWS network ACL API only speaks protocol numbers, and
		// that's all we record.
		protocolNumber, err := networkACLProtocolNumber(v)

		if err != nil {
			log.Printf("[WARN] %s", err)
			return nil
		}

		tfMap[names.AttrProtocol] = strconv.Itoa(protocolNumber)
	}

	if apiObject := apiObject.IcmpTypeCode; apiObject != nil {
		if v := apiObject.Code; v != nil {
			tfMap["icmp_code"] = aws.ToInt32(v)
		}

		if v := apiObject.Type; v != nil {
			tfMap["icmp_type"] = aws.ToInt32(v)
		}
	}

	return tfMap
}

func flattenNetworkACLEntries(apiObjects []awstypes.NetworkAclEntry) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenNetworkACLEntry(apiObject))
	}

	return tfList
}

func networkACLProtocolNumber(v string) (int, error) {
	i, err := strconv.Atoi(v)

	if err != nil {
		// Lookup number by name.
		var ok bool
		i, ok = ianaProtocolAToI[v]

		if !ok {
			return 0, fmt.Errorf("unsupported NACL protocol: %s", v)
		}
	} else {
		_, ok := ianaProtocolIToA[i]

		if !ok {
			return 0, fmt.Errorf("unsupported NACL protocol: %d", i)
		}
	}

	return i, nil
}

type ianaProtocolAToIType map[string]int
type ianaProtocolIToAType map[int]string

func (m ianaProtocolAToIType) invert() ianaProtocolIToAType {
	output := make(map[int]string, len(m))

	for k, v := range m {
		output[v] = k
	}

	return output
}

var (
	ianaProtocolAToI = ianaProtocolAToIType(map[string]int{
		// Defined at https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml.
		"all":             -1,
		"hopopt":          0,
		"icmp":            1,
		"igmp":            2,
		"ggp":             3,
		"ipv4":            4,
		"st":              5,
		"tcp":             6,
		"cbt":             7,
		"egp":             8,
		"igp":             9,
		"bbn-rcc-mon":     10,
		"nvp-ii":          11,
		"pup":             12,
		"argus":           13,
		"emcon":           14,
		"xnet":            15,
		"chaos":           16,
		"udp":             17,
		"mux":             18,
		"dcn-meas":        19,
		"hmp":             20,
		"prm":             21,
		"xns-idp":         22,
		"trunk-1":         23,
		"trunk-2":         24,
		"leaf-1":          25,
		"leaf-2":          26,
		"rdp":             27,
		"irtp":            28,
		"iso-tp4":         29,
		"netblt":          30,
		"mfe-nsp":         31,
		"merit-inp":       32,
		"dccp":            33,
		"3pc":             34,
		"idpr":            35,
		"xtp":             36,
		"ddp":             37,
		"idpr-cmtp":       38,
		"tp++":            39,
		"il":              40,
		"ipv6":            41,
		"sdrp":            42,
		"ipv6-route":      43,
		"ipv6-frag":       44,
		"idrp":            45,
		"rsvp":            46,
		"gre":             47,
		"dsr":             48,
		"bna":             49,
		"esp":             50,
		"ah":              51,
		"i-nlsp":          52,
		"swipe":           53,
		"narp":            54,
		"mobile":          55,
		"tlsp":            56,
		"ipv6-icmp":       58,
		"ipv6-nonxt":      59,
		"ipv6-opts":       60,
		"61":              61,
		"cftp":            62,
		"63":              63,
		"sat-expak":       64,
		"kryptolan":       65,
		"rvd":             66,
		"ippc":            67,
		"68":              68,
		"sat-mon":         69,
		"visa":            70,
		"ipcv":            71,
		"cpnx":            72,
		"cphb":            73,
		"wsn":             74,
		"pvp":             75,
		"br-sat-mon":      76,
		"sun-nd":          77,
		"wb-mon":          78,
		"wb-expak":        79,
		"iso-ip":          80,
		"vmtp":            81,
		"secure-vmtp":     82,
		"vines":           83,
		"ttp":             84,
		"nsfnet-igp":      85,
		"dgp":             86,
		"tcf":             87,
		"eigrp":           88,
		"ospfigp":         89,
		"sprite-rpc":      90,
		"larp":            91,
		"mtp":             92,
		"ax.25":           93,
		"ipip":            94,
		"micp":            95,
		"scc-sp":          96,
		"etherip":         97,
		"encap":           98,
		"99":              99,
		"gmtp":            100,
		"ifmp":            101,
		"pnni":            102,
		"pim":             103,
		"aris":            104,
		"scps":            105,
		"qnx":             106,
		"a/n":             107,
		"ipcomp":          108,
		"snp":             109,
		"compaq-peer":     110,
		"ipx-in-ip":       111,
		"vrrp":            112,
		"pgm":             113,
		"114":             114,
		"l2tp":            115,
		"dd":              116,
		"iatp":            117,
		"stp":             118,
		"srp":             119,
		"uti":             120,
		"smp":             121,
		"sm":              122,
		"ptp":             123,
		"isis-over-ipv4":  124,
		"fire":            125,
		"crtp":            126,
		"crudp":           127,
		"sscopmce":        128,
		"iplt":            129,
		"sps":             130,
		"pipe":            131,
		"sctp":            132,
		"fc":              133,
		"rsvp-e2e-ignore": 134,
		"mobility-header": 135,
		"udplite":         136,
		"mpls-in-ip":      137,
		"manet":           138,
		"hip":             139,
		"shim6":           140,
		"wesp":            141,
		"rohc":            142,
		"253":             253,
		"254":             254,
	})
	ianaProtocolIToA = ianaProtocolAToI.invert()
)
