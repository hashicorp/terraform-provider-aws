package ec2

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceNetworkACL() *schema.Resource {
	networkACLRuleSetSchema := &schema.Schema{
		Type:       schema.TypeSet,
		Optional:   true,
		Computed:   true,
		ConfigMode: schema.SchemaConfigModeAttr,
		Elem:       networkACLRuleResource,
		Set:        resourceNetworkACLEntryHash,
	}

	return &schema.Resource{
		Create: resourceNetworkACLCreate,
		Read:   resourceNetworkACLRead,
		Update: resourceNetworkACLUpdate,
		Delete: resourceNetworkACLDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				conn := meta.(*conns.AWSClient).EC2Conn

				nacl, err := FindNetworkACLByID(conn, d.Id())

				if err != nil {
					return nil, err
				}

				if aws.BoolValue(nacl.IsDefault) {
					return nil, fmt.Errorf("use the `aws_default_network_acl` resource instead")
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		// Keep in sync with aws_default_network_acl's schema.
		// See notes in default_network_acl.go.
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"egress":  networkACLRuleSetSchema,
			"ingress": networkACLRuleSetSchema,
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

// NACL rule Resource definition.
// Used in aws_network_acl and aws_default_network_acl ingress and egress rule sets.
var networkACLRuleResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"action": {
			Type:     schema.TypeString,
			Required: true,
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				return strings.EqualFold(old, new)
			},
			ValidateFunc: validation.StringInSlice(ec2.RuleAction_Values(), true),
		},
		"cidr_block": {
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
		"protocol": {
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

func resourceNetworkACLCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateNetworkAclInput{
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeNetworkAcl),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Network ACL: %s", input)
	output, err := conn.CreateNetworkAcl(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Network ACL: %w", err)
	}

	d.SetId(aws.StringValue(output.NetworkAcl.NetworkAclId))

	if err := modifyNetworkACLAttributesOnCreate(conn, d); err != nil {
		return err
	}

	return resourceNetworkACLRead(d, meta)
}

func resourceNetworkACLRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(PropagationTimeout, func() (interface{}, error) {
		return FindNetworkACLByID(conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network ACL %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Network ACL (%s): %w", d.Id(), err)
	}

	nacl := outputRaw.(*ec2.NetworkAcl)

	ownerID := aws.StringValue(nacl.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("network-acl/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("owner_id", ownerID)

	var subnetIDs []string
	for _, v := range nacl.Associations {
		subnetIDs = append(subnetIDs, aws.StringValue(v.SubnetId))
	}
	d.Set("subnet_ids", subnetIDs)

	d.Set("vpc_id", nacl.VpcId)

	var egressEntries []*ec2.NetworkAclEntry
	var ingressEntries []*ec2.NetworkAclEntry
	for _, v := range nacl.Entries {
		// Skip the default rules added by AWS. They can be neither
		// configured or deleted by users.
		if v := aws.Int64Value(v.RuleNumber); v == defaultACLRuleNumberIPv4 || v == defaultACLRuleNumberIPv6 {
			continue
		}

		if aws.BoolValue(v.Egress) {
			egressEntries = append(egressEntries, v)
		} else {
			ingressEntries = append(ingressEntries, v)
		}
	}
	if err := d.Set("egress", flattenNetworkAclEntries(egressEntries)); err != nil {
		return fmt.Errorf("error setting egress: %w", err)
	}
	if err := d.Set("ingress", flattenNetworkAclEntries(ingressEntries)); err != nil {
		return fmt.Errorf("error setting ingress: %w", err)
	}

	tags := KeyValueTags(nacl.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceNetworkACLUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if err := modifyNetworkACLAttributesOnUpdate(conn, d, true); err != nil {
		return err
	}

	return resourceNetworkACLRead(d, meta)
}

func resourceNetworkACLDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	// Delete all NACL/Subnet associations, even if they are managed via aws_network_acl_association resources.
	nacl, err := FindNetworkACLByID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading EC2 Network ACL (%s): %w", d.Id(), err)
	}

	var subnetIDs []interface{}
	for _, v := range nacl.Associations {
		subnetIDs = append(subnetIDs, aws.StringValue(v.SubnetId))
	}
	if len(subnetIDs) > 0 {
		if err := networkACLAssociationsDelete(conn, d.Get("vpc_id").(string), subnetIDs); err != nil {
			return err
		}
	}

	input := &ec2.DeleteNetworkAclInput{
		NetworkAclId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting EC2 Network ACL: %s", d.Id())
	_, err = tfresource.RetryWhenAWSErrCodeEquals(PropagationTimeout, func() (interface{}, error) {
		return conn.DeleteNetworkAcl(input)
	}, ErrCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidNetworkAclIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Network ACL (%s): %w", d.Id(), err)
	}

	return nil
}

// modifyNetworkACLAttributesOnCreate sets NACL attributes on resource Create.
// Called after new NACL creation or existing default NACL adoption.
// Tags are not configured.
func modifyNetworkACLAttributesOnCreate(conn *ec2.EC2, d *schema.ResourceData) error {
	if v, ok := d.GetOk("egress"); ok && v.(*schema.Set).Len() > 0 {
		if err := createNetworkACLEntries(conn, d.Id(), v.(*schema.Set).List(), true); err != nil {
			return err
		}
	}

	if v, ok := d.GetOk("ingress"); ok && v.(*schema.Set).Len() > 0 {
		if err := createNetworkACLEntries(conn, d.Id(), v.(*schema.Set).List(), false); err != nil {
			return err
		}
	}

	if v, ok := d.GetOk("subnet_ids"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			if _, err := networkACLAssociationCreate(conn, d.Id(), v.(string)); err != nil {
				return err
			}
		}
	}

	return nil
}

// modifyNetworkACLAttributesOnUpdate sets NACL attributes on resource Update.
// Tags are configured.
func modifyNetworkACLAttributesOnUpdate(conn *ec2.EC2, d *schema.ResourceData, deleteEntries bool) error {
	if d.HasChange("ingress") {
		o, n := d.GetChange("ingress")
		os, ns := o.(*schema.Set), n.(*schema.Set)

		if err := updateNetworkACLEntries(conn, d.Id(), os, ns, false, deleteEntries); err != nil {
			return err
		}
	}

	if d.HasChange("egress") {
		o, n := d.GetChange("egress")
		os, ns := o.(*schema.Set), n.(*schema.Set)

		if err := updateNetworkACLEntries(conn, d.Id(), os, ns, true, deleteEntries); err != nil {
			return err
		}
	}

	if d.HasChange("subnet_ids") {
		o, n := d.GetChange("subnet_ids")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := ns.Difference(os).List(), os.Difference(ns).List()

		if len(del) > 0 {
			if err := networkACLAssociationsDelete(conn, d.Get("vpc_id").(string), del); err != nil {
				return err
			}
		}

		if len(add) > 0 {
			if err := networkACLAssociationsCreate(conn, d.Id(), add); err != nil {
				return err
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Network ACL (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceNetworkACLEntryHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["from_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["to_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["rule_no"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["action"].(string))))

	// The AWS network ACL API only speaks protocol numbers, and that's
	// all we store. Never hash a protocol name.
	protocol := m["protocol"].(string)
	if _, err := strconv.Atoi(m["protocol"].(string)); err != nil {
		// We're a protocol name. Look up the number.
		buf.WriteString(fmt.Sprintf("%d-", protocolIntegers()[protocol]))
	} else {
		// We're a protocol number. Pass the value through.
		buf.WriteString(fmt.Sprintf("%s-", protocol))
	}

	if v, ok := m["cidr_block"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["ipv6_cidr_block"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["ssl_certificate_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["icmp_type"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["icmp_code"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}

	return create.StringHashcode(buf.String())
}

func createNetworkACLEntries(conn *ec2.EC2, naclID string, tfList []interface{}, egress bool) error {
	naclEntries := expandNetworkAclEntries(tfList, egress)

	for _, naclEntry := range naclEntries {
		if naclEntry == nil {
			continue
		}

		if aws.StringValue(naclEntry.Protocol) == "-1" {
			// Protocol -1 rules don't store ports in AWS. Thus, they'll always
			// hash differently when being read out of the API. Force the user
			// to set from_port and to_port to 0 for these rules, to keep the
			// hashing consistent.
			if from, to := aws.Int64Value(naclEntry.PortRange.From), aws.Int64Value(naclEntry.PortRange.To); from != 0 || to != 0 {
				return fmt.Errorf("to_port (%d) and from_port (%d) must both be 0 to use the the 'all' \"-1\" protocol!", to, from)
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

		log.Printf("[INFO] Creating EC2 Network ACL Entry: %s", input)
		_, err := conn.CreateNetworkAclEntry(input)

		if err != nil {
			return fmt.Errorf("error creating EC2 Network ACL (%s) Entry: %w", naclID, err)
		}
	}

	return nil
}

func deleteNetworkACLEntries(conn *ec2.EC2, naclID string, tfList []interface{}, egress bool) error {
	return deleteNetworkAclEntries(conn, naclID, expandNetworkAclEntries(tfList, egress))
}

func deleteNetworkAclEntries(conn *ec2.EC2, naclID string, naclEntries []*ec2.NetworkAclEntry) error {
	for _, naclEntry := range naclEntries {
		if naclEntry == nil {
			continue
		}

		// AWS includes default rules with all network ACLs that can be
		// neither modified nor destroyed. They have a custom rule
		// number that is out of bounds for any other rule. If we
		// encounter it, just continue. There's no work to be done.
		if v := aws.Int64Value(naclEntry.RuleNumber); v == defaultACLRuleNumberIPv4 || v == defaultACLRuleNumberIPv6 {
			continue
		}

		input := &ec2.DeleteNetworkAclEntryInput{
			Egress:       naclEntry.Egress,
			NetworkAclId: aws.String(naclID),
			RuleNumber:   naclEntry.RuleNumber,
		}

		log.Printf("[INFO] Deleting EC2 Network ACL Entry: %s", input)
		_, err := conn.DeleteNetworkAclEntry(input)

		if err != nil {
			return fmt.Errorf("error deleting EC2 Network ACL (%s) Entry: %w", naclID, err)
		}
	}

	return nil
}

func updateNetworkACLEntries(conn *ec2.EC2, naclID string, os, ns *schema.Set, egress bool, deleteEntries bool) error {
	if deleteEntries {
		if err := deleteNetworkACLEntries(conn, naclID, os.Difference(ns).List(), egress); err != nil {
			return err
		}
	}

	if err := createNetworkACLEntries(conn, naclID, ns.Difference(os).List(), egress); err != nil {
		return err
	}

	return nil
}

func expandNetworkAclEntry(tfMap map[string]interface{}, egress bool) *ec2.NetworkAclEntry {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.NetworkAclEntry{
		Egress:    aws.Bool(egress),
		PortRange: &ec2.PortRange{},
	}

	if v, ok := tfMap["rule_no"].(int); ok {
		apiObject.RuleNumber = aws.Int64(int64(v))
	}

	if v, ok := tfMap["action"].(string); ok && v != "" {
		apiObject.RuleAction = aws.String(v)
	}

	if v, ok := tfMap["cidr_block"].(string); ok && v != "" {
		apiObject.CidrBlock = aws.String(v)
	}

	if v, ok := tfMap["ipv6_cidr_block"].(string); ok && v != "" {
		apiObject.Ipv6CidrBlock = aws.String(v)
	}

	if v, ok := tfMap["from_port"].(int); ok {
		apiObject.PortRange.From = aws.Int64(int64(v))
	}

	if v, ok := tfMap["to_port"].(int); ok {
		apiObject.PortRange.To = aws.Int64(int64(v))
	}

	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		i, err := networkACLProtocolNumber(v)

		if err != nil {
			log.Printf("[WARN] %s", err)
			return nil
		}

		apiObject.Protocol = aws.String(strconv.Itoa(i))

		// Specify additional required fields for ICMP.
		if i == 1 || i == 58 {
			apiObject.IcmpTypeCode = &ec2.IcmpTypeCode{}

			if v, ok := tfMap["icmp_code"].(int); ok {
				apiObject.IcmpTypeCode.Code = aws.Int64(int64(v))
			}

			if v, ok := tfMap["icmp_type"].(int); ok {
				apiObject.IcmpTypeCode.Type = aws.Int64(int64(v))
			}
		}
	}

	return apiObject
}

func expandNetworkAclEntries(tfList []interface{}, egress bool) []*ec2.NetworkAclEntry {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.NetworkAclEntry

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandNetworkAclEntry(tfMap, egress)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenNetworkAclEntry(apiObject *ec2.NetworkAclEntry) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RuleNumber; v != nil {
		tfMap["rule_no"] = aws.Int64Value(v)
	}

	if v := apiObject.RuleAction; v != nil {
		tfMap["action"] = aws.StringValue(v)
	}

	if v := apiObject.CidrBlock; v != nil {
		tfMap["cidr_block"] = aws.StringValue(v)
	}

	if v := apiObject.Ipv6CidrBlock; v != nil {
		tfMap["ipv6_cidr_block"] = aws.StringValue(v)
	}

	if apiObject := apiObject.PortRange; apiObject != nil {
		if v := apiObject.From; v != nil {
			tfMap["from_port"] = aws.Int64Value(v)
		}

		if v := apiObject.To; v != nil {
			tfMap["to_port"] = aws.Int64Value(v)
		}
	}

	if v := aws.StringValue(apiObject.Protocol); v != "" {
		// The AWS network ACL API only speaks protocol numbers, and
		// that's all we record.
		i, err := networkACLProtocolNumber(v)

		if err != nil {
			log.Printf("[WARN] %s", err)
			return nil
		}

		tfMap["protocol"] = strconv.Itoa(i)
	}

	if apiObject := apiObject.IcmpTypeCode; apiObject != nil {
		if v := apiObject.Code; v != nil {
			tfMap["icmp_code"] = aws.Int64Value(v)
		}

		if v := apiObject.Type; v != nil {
			tfMap["icmp_type"] = aws.Int64Value(v)
		}
	}

	return tfMap
}

func flattenNetworkAclEntries(apiObjects []*ec2.NetworkAclEntry) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenNetworkAclEntry(apiObject))
	}

	return tfList
}

func networkACLProtocolNumber(v string) (int, error) {
	i, err := strconv.Atoi(v)

	if err != nil {
		// Lookup number by name.
		var ok bool
		i, ok = protocolIntegers()[v]

		if !ok {
			return 0, fmt.Errorf("unsupported NACL protocol: %s", v)
		}
	} else {
		_, ok := protocolStrings(protocolIntegers())[i]

		if !ok {
			return 0, fmt.Errorf("unsupported NACL protocol: %d", i)
		}
	}

	return i, nil
}
