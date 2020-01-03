package aws

import (
	"bytes"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2TransitGatewayMulticastDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayMulticastDomainCreate,
		Read:   resourceAwsEc2TransitGatewayMulticastDomainRead,
		Update: resourceAwsEc2TransitGatewayMulticastDomainUpdate,
		Delete: resourceAwsEc2TransitGatewayMulticastDomainDelete,

		Schema: map[string]*schema.Schema{
			"transit_gateway_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"tags": tagsSchema(),
			"association": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transit_gateway_attachment_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
						},
					},
				},
				Set: resourceAwsEc2TransitGatewayMulticastDomainAssociationHash,
			},
			"members": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"network_interface_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
						},
					},
				},
				Set: resourceAwsEc2TransitGatewayMulticastDomainGroupsHash,
			},
			"sources": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"network_interface_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
						},
					},
				},
				Set: resourceAwsEc2TransitGatewayMulticastDomainGroupsHash,
			},
		},
	}
}

func resourceAwsEc2TransitGatewayMulticastDomainCreate(d *schema.ResourceData, meta interface{}) error {
	// create the domain
	conn := meta.(*AWSClient).ec2conn
	input := &ec2.CreateTransitGatewayMulticastDomainInput{
		TransitGatewayId: aws.String(d.Get("transit_gateway_id").(string)),
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}),
			ec2.ResourceTypeTransitGatewayMulticastDomain),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Multicast Domain: %s", input)
	output, err := conn.CreateTransitGatewayMulticastDomain(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Multicast Domain: %s", err)
	}

	id := aws.StringValue(output.TransitGatewayMulticastDomain.TransitGatewayMulticastDomainId)
	d.SetId(id)

	// wait for the domain to become available
	if err := waitForEc2TransitGatewayMulticastDomainCreation(conn, id); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Multicast Domain (%s) availability: %s", id, err)
	}

	return resourceAwsEc2TransitGatewayMulticastDomainUpdate(d, meta)
}

func resourceAwsEc2TransitGatewayMulticastDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()

	if err := resourceAwsEc2TransitGatewayMulticastDomainAssociationsRead(d, meta); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsRead(d, meta, true); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsRead(d, meta, false); err != nil {
		return err
	}

	multicastDomain, err := ec2DescribeTransitGatewayMulticastDomain(conn, id)
	if multicastDomain == nil || isAWSErr(err, "InvalidTransitGatewayMulticastDomainId.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Multicast Domain: %s", err)
	}

	if aws.StringValue(multicastDomain.State) == ec2.TransitGatewayStateDeleting ||
		aws.StringValue(multicastDomain.State) == ec2.TransitGatewayStateDeleted {
		log.Printf(
			"[WARN] EC2 Transit Gateway Multicast Domain (%s) in deleted state (%s), removing from state",
			d.Id(), aws.StringValue(multicastDomain.State))
		d.SetId("")
		return nil
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(multicastDomain.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error settings tags: %s", err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainAssociationsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()

	// get associations by grouping by the attachment ID
	associations, err := ec2GetTransitGatewayMulticastDomainAssociations(conn, id)
	if err != nil {
		return fmt.Errorf("error getting EC2 Transit Gateway Multicast Domain (%s) associations: %s", id, err)
	}

	tgwAttachmentMap := make(map[string][]*string)
	for _, assoc := range associations {
		attachmentID := aws.StringValue(assoc.TransitGatewayAttachmentId)
		subnetID := assoc.Subnet.SubnetId
		if lst, exists := tgwAttachmentMap[attachmentID]; exists {
			tgwAttachmentMap[attachmentID] = append(lst, subnetID)
			continue
		}
		tgwAttachmentMap[attachmentID] = []*string{subnetID}
	}

	// flatten data so that each association is to 1 tgw attachment
	assocSet := &schema.Set{F: resourceAwsEc2TransitGatewayMulticastDomainAssociationHash}
	for attachmentID := range tgwAttachmentMap {
		assocData := make(map[string]interface{})
		assocData["transit_gateway_attachment_id"] = attachmentID
		assocData["subnet_ids"] = flattenStringSet(tgwAttachmentMap[attachmentID])
		assocSet.Add(assocData)
	}
	d.Set("association", assocSet)

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupsRead(d *schema.ResourceData, meta interface{}, member bool) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()

	// get groups by grouping by the group IP address
	groups, err := ec2GetTransitGatewayMulticastDomainGroups(conn, id, member)
	if err != nil {
		return fmt.Errorf("error getting EC2 Transit Gateway Multicast Domain (%s) groups: %s", id, err)
	}

	groupIpMap := make(map[string][]*string)
	for _, group := range groups {
		groupIP := aws.StringValue(group.GroupIpAddress)
		netID := group.NetworkInterfaceId
		if lst, exists := groupIpMap[groupIP]; exists {
			groupIpMap[groupIP] = append(lst, netID)
			continue
		}
		groupIpMap[groupIP] = []*string{netID}
	}

	groupSet := &schema.Set{F: resourceAwsEc2TransitGatewayMulticastDomainGroupsHash}
	for groupIP := range groupIpMap {
		groupData := make(map[string]interface{})
		groupData["group_ip_address"] = groupIP
		groupData["network_interface_ids"] = flattenStringSet(groupIpMap[groupIP])
		groupSet.Add(groupData)
	}

	if member {
		d.Set("members", groupSet)
		return nil
	}
	d.Set("sources", groupSet)

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Ec2UpdateTags(conn, id, o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Multicast Domain (%s) tags: %s", id, err)
		}
	}

	if err := resourceAwsEc2TransitGatewayMulticastDomainAssociationsUpdate(d, meta); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsUpdate(d, meta, true); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsUpdate(d, meta, false); err != nil {
		return err
	}

	return resourceAwsEc2TransitGatewayMulticastDomainRead(d, meta)
}

func resourceAwsEc2TransitGatewayMulticastDomainAssociationsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	// check if association set has changed
	if d.HasChange("association") {
		o, n := d.GetChange("association")
		old := o.(*schema.Set).Difference(n.(*schema.Set))
		nw := n.(*schema.Set).Difference(o.(*schema.Set))

		// disassociate old associations
		for _, assoc := range old.List() {
			assocData := assoc.(map[string]interface{})
			if err := resourceAwsEc2TransitGatewayMulticastDomainDisassociate(conn, id, assocData); err != nil {
				return err
			}
		}

		// save current state
		associations := o.(*schema.Set).Intersection(n.(*schema.Set))
		d.Set("association", associations)

		// associate new subnets
		for _, assoc := range nw.List() {
			assocData := assoc.(map[string]interface{})
			if err := resourceAwsEc2TransitGatewayMulticastDomainAssociate(conn, id, assocData); err != nil {
				return err
			}
			associations.Add(assoc)
			d.Set("association", associations)
		}
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupsUpdate(d *schema.ResourceData, meta interface{}, member bool) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	key := groupType(member)
	if d.HasChange(key) {
		o, n := d.GetChange(key)
		old := o.(*schema.Set).Difference(n.(*schema.Set))
		nw := n.(*schema.Set).Difference(o.(*schema.Set))

		// remove old groups
		for _, group := range old.List() {
			groupData := group.(map[string]interface{})
			if err := resourceAwsEc2TransitGatewayMulticastDomainGroupDeregister(conn, id, groupData, member); err != nil {
				return err
			}
		}

		// save current state
		groups := o.(*schema.Set).Intersection(n.(*schema.Set))
		d.Set(key, groups)

		// register new groups
		for _, group := range nw.List() {
			groupData := group.(map[string]interface{})
			if err := resourceAwsEc2TransitGatewayMulticastDomainGroupRegister(conn, id, groupData, member); err != nil {
				return err
			}
			groups.Add(group)
			d.Set(key, groups)
		}
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainDisassociate(conn *ec2.EC2, id string, assocData map[string]interface{}) error {
	tgwAttachmentID := assocData["transit_gateway_attachment_id"].(string)
	subnetIDs := expandStringSet(assocData["subnet_ids"].(*schema.Set))
	input := &ec2.DisassociateTransitGatewayMulticastDomainInput{
		SubnetIds:                       subnetIDs,
		TransitGatewayAttachmentId:      aws.String(tgwAttachmentID),
		TransitGatewayMulticastDomainId: aws.String(id),
	}

	log.Printf(
		"[DEBUG] Disassociating subnets from EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
	if _, err := conn.DisassociateTransitGatewayMulticastDomain(input); err != nil {
		return fmt.Errorf(
			"error disassociating EC2 Transit Gateway Multicast Domain (%s) subnets: %s", id, err)
	}

	// wait for subnets to be disassociated
	if err := waitForEc2TransitGatewayMulticastDomainDisassociation(conn, id, subnetIDs); err != nil {
		return fmt.Errorf(
			"error waiting for EC2 Transit Gateway Multicast Domain (%s) to disassociate subnets: %s",
			id, err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainAssociate(conn *ec2.EC2, id string, assocData map[string]interface{}) error {
	tgwAttachmentID := assocData["transit_gateway_attachment_id"].(string)
	subnetIDs := expandStringSet(assocData["subnet_ids"].(*schema.Set))
	input := &ec2.AssociateTransitGatewayMulticastDomainInput{
		SubnetIds:                       subnetIDs,
		TransitGatewayAttachmentId:      aws.String(tgwAttachmentID),
		TransitGatewayMulticastDomainId: aws.String(id),
	}

	log.Printf(
		"[DEBUG] Associating subnets to EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
	if _, err := conn.AssociateTransitGatewayMulticastDomain(input); err != nil {
		return fmt.Errorf(
			"error associating EC2 Transit Gateway Multicast Domain (%s) subnets: %s", id, err)
	}

	// wait for associations
	if err := waitForEc2TransitGatewayMulticastDomainAssociation(conn, id, subnetIDs); err != nil {
		return fmt.Errorf(
			"error waiting for EC2 Transit Gateway Multicast Domain (%s) associations: %s", id, err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupDeregister(conn *ec2.EC2, id string, groupData map[string]interface{}, member bool) error {
	// Note: for some reason AWS made the decision to logically separate "members" from "sources" in
	// register/deregister; however, they are unified in "search"
	if member {
		input := &ec2.DeregisterTransitGatewayMulticastGroupMembersInput{
			GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
			NetworkInterfaceIds:             expandStringSet(groupData["network_interface_ids"].(*schema.Set)),
			TransitGatewayMulticastDomainId: aws.String(id),
		}

		log.Printf(
			"[DEBUG] Deregistering members from EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
		if _, err := conn.DeregisterTransitGatewayMulticastGroupMembers(input); err != nil {
			return fmt.Errorf(
				"error removing member from EC2 Transit Gateway Multicast Domain (%s): %s", id, err)
		}
		return nil
	}

	input := &ec2.DeregisterTransitGatewayMulticastGroupSourcesInput{
		GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
		NetworkInterfaceIds:             expandStringSet(groupData["network_interface_ids"].(*schema.Set)),
		TransitGatewayMulticastDomainId: aws.String(id),
	}

	log.Printf(
		"[DEBUG] Deregistering sources from EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
	if _, err := conn.DeregisterTransitGatewayMulticastGroupSources(input); err != nil {
		return fmt.Errorf(
			"error removing source from EC2 Transit Gateway Multicast Domain (%s): %s", id, err)
	}
	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupRegister(conn *ec2.EC2, id string, groupData map[string]interface{}, member bool) error {
	// Note: for some reason AWS made the decision to logically separate "members" from "sources" in
	// register/deregister; however, they are unified in "search"
	if member {
		input := &ec2.RegisterTransitGatewayMulticastGroupMembersInput{
			GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
			NetworkInterfaceIds:             expandStringSet(groupData["network_interface_ids"].(*schema.Set)),
			TransitGatewayMulticastDomainId: aws.String(id),
		}

		log.Printf(
			"[DEBUG] Registering members to EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
		if _, err := conn.RegisterTransitGatewayMulticastGroupMembers(input); err != nil {
			return fmt.Errorf(
				"error registering EC2 Transit Gateway Multicast Domain (%s) members: %s", id, err)
		}
		return nil
	}

	input := &ec2.RegisterTransitGatewayMulticastGroupSourcesInput{
		GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
		NetworkInterfaceIds:             expandStringSet(groupData["network_interface_ids"].(*schema.Set)),
		TransitGatewayMulticastDomainId: aws.String(id),
	}

	log.Printf(
		"[DEBUG] Registering sources to EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
	if _, err := conn.RegisterTransitGatewayMulticastGroupSources(input); err != nil {
		return fmt.Errorf(
			"error registering EC2 Transit Gateway Multicast Domain (%s) sources: %s", id, err)
	}
	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	if err := resourceAwsEc2TransitGatewayMulticastDomainRead(d, meta); err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Multicast Domain (%s): %s", id, err)
	}

	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsDeregisterAll(d, meta, true); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainGroupsDeregisterAll(d, meta, false); err != nil {
		return err
	}
	if err := resourceAwsEc2TransitGatewayMulticastDomainDisassociateAll(d, meta); err != nil {
		return err
	}

	input := &ec2.DeleteTransitGatewayMulticastDomainInput{
		TransitGatewayMulticastDomainId: aws.String(id),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
	_, err := conn.DeleteTransitGatewayMulticastDomain(input)
	if isAWSErr(err, "InvalidTransitGatewayMulticastDomainId.NotFound", "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Multicast Domain: %s", err)
	}

	if err := waitForEc2TransitGatewayMulticastDomainDeletion(conn, id); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Multicast Domain (%s) deletion: %s", id, err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainDisassociateAll(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	// disassociate subnets
	assocSet := d.Get("association").(*schema.Set)
	for _, assoc := range assocSet.List() {
		assocData := assoc.(map[string]interface{})
		if err := resourceAwsEc2TransitGatewayMulticastDomainDisassociate(conn, id, assocData); err != nil {
			return err
		}
	}
	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupsDeregisterAll(d *schema.ResourceData, meta interface{}, member bool) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	key := groupType(member)
	// deregister groups
	groups := d.Get(key).(*schema.Set)
	for _, group := range groups.List() {
		groupData := group.(map[string]interface{})
		if err := resourceAwsEc2TransitGatewayMulticastDomainGroupDeregister(conn, id, groupData, member); err != nil {
			return err
		}
	}
	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainAssociationHash(meta interface{}) int {
	m, castOk := meta.(map[string]interface{})
	if !castOk {
		return 0
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", m["transit_gateway_attachment_id"].(string)))
	for _, subnetID := range m["subnet_ids"].(*schema.Set).List() {
		buf.WriteString(fmt.Sprintf("%s,", subnetID.(string)))
	}
	return hashcode.String(buf.String())
}

func resourceAwsEc2TransitGatewayMulticastDomainGroupsHash(meta interface{}) int {
	m, castOk := meta.(map[string]interface{})
	if !castOk {
		return 0
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", m["group_ip_address"].(string)))
	for _, networkID := range m["network_interface_ids"].(*schema.Set).List() {
		buf.WriteString(fmt.Sprintf("%s,", networkID.(string)))
	}
	return hashcode.String(buf.String())
}

func groupType(member bool) string {
	if member {
		return "members"
	}
	return "sources"
}
