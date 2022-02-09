package ec2

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceTransitGatewayMulticastDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayMulticastDomainCreate,
		Read:   resourceTransitGatewayMulticastDomainRead,
		Update: resourceTransitGatewayMulticastDomainUpdate,
		Delete: resourceTransitGatewayMulticastDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
							MinItems: 1,
							Set:      schema.HashString,
						},
					},
				},
				Set: resourceTransitGatewayMulticastDomainAssociationsHash,
			},
			"auto_accept_shared_associations": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ec2.AutoAcceptSharedAssociationsValueDisable,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.AutoAcceptSharedAssociationsValueEnable,
					ec2.AutoAcceptSharedAssociationsValueDisable,
				}, false),
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"igmpv2_support": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "disable",
				ConflictsWith: []string{"static_source_support"},
				ValidateFunc:  validation.StringInSlice([]string{"enable", "disable"}, false),
			},
			"members": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: IsMulticastAddress,
						},
						"network_interface_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
							MinItems: 1,
							Set:      schema.HashString,
						},
					},
				},
				Set: resourceTransitGatewayMulticastDomainGroupsHash,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sources": {
				Type:       schema.TypeSet,
				Computed:   true,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: IsMulticastAddress,
						},
						"network_interface_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
							MinItems: 1,
							Set:      schema.HashString,
						},
					},
				},
				Set: resourceTransitGatewayMulticastDomainGroupsHash,
			},

			"static_source_support": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "disable",
				ConflictsWith: []string{"igmpv2_support"},
				ValidateFunc:  validation.StringInSlice([]string{"enable", "disable"}, false),
			},
			"tags": tftags.TagsSchema(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceTransitGatewayMulticastDomainCreate(d *schema.ResourceData, meta interface{}) error {
	// create the domain
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateTransitGatewayMulticastDomainInput{
		TransitGatewayId: aws.String(d.Get("transit_gateway_id").(string)),
		Options:          &ec2.CreateTransitGatewayMulticastDomainRequestOptions{},

		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayMulticastDomain),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Multicast Domain: %s", input)

	if v, ok := d.GetOk("igmpv2_support"); ok {
		input.Options.Igmpv2Support = aws.String(v.(string))
	}

	if v, ok := d.GetOk("static_source_support"); ok {
		input.Options.StaticSourcesSupport = aws.String(v.(string))
	}

	if v, ok := d.GetOk("auto_accept_shared_associations"); ok {
		input.Options.AutoAcceptSharedAssociations = aws.String(v.(string))
	}
	log.Printf("[WARN] %v", input)

	output, err := conn.CreateTransitGatewayMulticastDomain(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Multicast Domain: %s", err)
	}

	id := aws.StringValue(output.TransitGatewayMulticastDomain.TransitGatewayMulticastDomainId)
	d.SetId(id)

	// wait for the domain to become available
	if err := WaitForTransitGatewayMulticastDomainCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Multicast Domain (%s) availability: %s", id, err)
	}

	return resourceTransitGatewayMulticastDomainUpdate(d, meta)
}

func resourceTransitGatewayMulticastDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	id := d.Id()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	multicastDomain, err := DescribeTransitGatewayMulticastDomain(conn, id)
	if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayMulticastDomainId.NotFound", "") {
		return resourceTransitGatewayMulticastDomainNotFound(d)
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Multicast Domain: %s", err)
	}

	if multicastDomain == nil {
		return resourceTransitGatewayMulticastDomainNotFound(d)
	}

	if err := resourceTransitGatewayMulticastDomainAssociationsRead(d, meta); err != nil {
		return err
	}
	if err := resourceTransitGatewayMulticastDomainGroupsRead(d, meta, true); err != nil {
		return err
	}
	if err := resourceTransitGatewayMulticastDomainGroupsRead(d, meta, false); err != nil {
		return err
	}

	if aws.StringValue(multicastDomain.State) == ec2.TransitGatewayStateDeleting ||
		aws.StringValue(multicastDomain.State) == ec2.TransitGatewayStateDeleted {
		log.Printf(
			"[WARN] EC2 Transit Gateway Multicast Domain (%s) in deleted state (%s), removing from state",
			d.Id(), aws.StringValue(multicastDomain.State))
		d.SetId("")
		return nil
	}

	d.Set("igmpv2_support", multicastDomain.Options.Igmpv2Support)
	d.Set("static_source_support", multicastDomain.Options.StaticSourcesSupport)
	d.Set("auto_accept_shared_associations", multicastDomain.Options.AutoAcceptSharedAssociations)

	tags := KeyValueTags(multicastDomain.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	//lintignore:AWSR002
	if err := d.Set("tags", tags.Map()); err != nil {
		return fmt.Errorf("error settings tags: %s", err)
	}

	return nil
}

func resourceTransitGatewayMulticastDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	id := d.Id()

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Multicast Domain (%s) tags: %s", id, err)
		}
	}

	if err := resourceTransitGatewayMulticastDomainAssociationsUpdate(d, meta); err != nil {
		return err
	}
	if err := resourceTransitGatewayMulticastDomainGroupsUpdate(d, meta, true); err != nil {
		return err
	}
	if err := resourceTransitGatewayMulticastDomainGroupsUpdate(d, meta, false); err != nil {
		return err
	}

	return resourceTransitGatewayMulticastDomainRead(d, meta)
}

func resourceTransitGatewayMulticastDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	id := d.Id()
	if err := resourceTransitGatewayMulticastDomainRead(d, meta); err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Multicast Domain (%s): %s", id, err)
	}

	if err := resourceTransitGatewayMulticastDomainGroupsDeregisterAll(d, meta, true); err != nil {
		return err
	}
	if err := resourceTransitGatewayMulticastDomainGroupsDeregisterAll(d, meta, false); err != nil {
		return err
	}
	if err := resourceTransitGatewayMulticastDomainDisassociateAll(d, meta); err != nil {
		return err
	}

	input := &ec2.DeleteTransitGatewayMulticastDomainInput{
		TransitGatewayMulticastDomainId: aws.String(id),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
	_, err := conn.DeleteTransitGatewayMulticastDomain(input)
	if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayMulticastDomainId.NotFound", "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Multicast Domain: %s", err)
	}

	if err := WaitForTransitGatewayMulticastDomainDeletion(conn, id); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Multicast Domain (%s) deletion: %s", id, err)
	}

	return nil
}

func resourceTransitGatewayMulticastDomainNotFound(d *schema.ResourceData) error {
	log.Printf("[WARN] EC2 Transit Gateway (%s) not found, removing from state", d.Id())
	d.SetId("")
	return nil
}

func resourceTransitGatewayMulticastDomainAssociationsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	id := d.Id()

	groupedRemoteAssocSet, err := resourceTransitGatewayMulticastDomainAssociationsReadRaw(conn, id)
	if err != nil {
		return err
	}

	localAssocSet := d.Get("association").(*schema.Set).List()

	// match local associations to remote
	assocSet := &schema.Set{F: resourceTransitGatewayMulticastDomainAssociationsHash}
	for _, localAssoc := range localAssocSet {
		assocData := localAssoc.(map[string]interface{})
		attachmentID := assocData["transit_gateway_attachment_id"].(string)
		subnetIDs := assocData["subnet_ids"].(*schema.Set)

		c, exists := groupedRemoteAssocSet[attachmentID]
		if !exists {
			continue
		}

		container := c.(map[string]interface{})
		remoteSubnetIDs := container["subnet_ids"].(*schema.Set)
		if len(remoteSubnetIDs.List()) == 0 {
			continue
		}

		newSubnetIDs := subnetIDs.Intersection(remoteSubnetIDs)
		if len(newSubnetIDs.List()) == 0 {
			continue
		}

		container["subnet_ids"] = remoteSubnetIDs.Difference(newSubnetIDs)
		groupedRemoteAssocSet[attachmentID] = container
		assocData["subnet_ids"] = newSubnetIDs
		assocSet.Add(assocData)
	}

	// add remaining remote associations
	for attachmentID, elem := range groupedRemoteAssocSet {
		subnetIDs := elem.(map[string]interface{})["subnet_ids"].(*schema.Set)
		if len(subnetIDs.List()) == 0 {
			continue
		}
		assocSet.Add(resourceTransitGatewayMulticastDomainAssociationMake(attachmentID, subnetIDs))
	}

	d.Set("association", assocSet)

	return nil
}

func resourceTransitGatewayMulticastDomainGroupsRead(d *schema.ResourceData, meta interface{}, member bool) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	id := d.Id()

	groupedRemoteGroups, err := resourceTransitGatewayMulticastDomainGroupsReadRaw(conn, id, member)
	if err != nil {
		return err
	}

	groupType := ResourceTransitGatewayMulticastDomainGroupType(member)
	localGroups := d.Get(groupType).(*schema.Set).List()

	groups := &schema.Set{F: resourceTransitGatewayMulticastDomainGroupsHash}
	for _, localGroup := range localGroups {
		groupData := localGroup.(map[string]interface{})
		groupIP := groupData["group_ip_address"].(string)
		netIDs := groupData["network_interface_ids"].(*schema.Set)

		c, exists := groupedRemoteGroups[groupIP]
		if !exists {
			continue
		}

		container := c.(map[string]interface{})
		remoteNetIDs := container["network_interface_ids"].(*schema.Set)
		if len(remoteNetIDs.List()) == 0 {
			continue
		}

		newNetIDs := netIDs.Intersection(remoteNetIDs)
		if len(newNetIDs.List()) == 0 {
			continue
		}

		container["network_interface_ids"] = remoteNetIDs.Difference(newNetIDs)
		groupedRemoteGroups[groupIP] = container
		groupData["network_interface_ids"] = newNetIDs
		groups.Add(groupData)
	}

	for groupIP, elem := range groupedRemoteGroups {
		netIDs := elem.(map[string]interface{})["network_interface_ids"].(*schema.Set)
		if len(netIDs.List()) == 0 {
			continue
		}
		groups.Add(resourceTransitGatewayMulticastDomainGroupMake(groupIP, netIDs))
	}

	d.Set(groupType, groups)

	return nil
}

func resourceTransitGatewayMulticastDomainGroupMake(groupIP string, netIDs *schema.Set) map[string]interface{} {
	group := make(map[string]interface{})
	group["group_ip_address"] = groupIP
	group["network_interface_ids"] = netIDs
	return group
}

func resourceTransitGatewayMulticastDomainAssociationMake(attachID string, subnetIDs *schema.Set) map[string]interface{} {
	assoc := make(map[string]interface{})
	assoc["transit_gateway_attachment_id"] = attachID
	assoc["subnet_ids"] = subnetIDs
	return assoc
}

func resourceTransitGatewayMulticastDomainGroupsReadRaw(conn *ec2.EC2, domainID string, member bool) (map[string]interface{}, error) {
	remoteGroups, err := SearchTransitGatewayMulticastDomainGroupsByType(conn, domainID, member)
	if err != nil {
		return nil, fmt.Errorf(
			"error getting EC2 Transit Gateway Multicast Domain (%s) groups: %s", domainID, err)
	}

	groupedRemoteGroups := make(map[string]interface{})
	for _, remoteGroup := range remoteGroups {
		groupIP := aws.StringValue(remoteGroup.GroupIpAddress)
		netID := aws.StringValue(remoteGroup.NetworkInterfaceId)
		if c, exists := groupedRemoteGroups[groupIP]; exists {
			container := c.(map[string]interface{})
			container["network_interface_ids"].(*schema.Set).Add(netID)
			continue
		}
		newSet := &schema.Set{F: schema.HashString}
		newSet.Add(netID)
		groupedRemoteGroups[groupIP] = resourceTransitGatewayMulticastDomainGroupMake(groupIP, newSet)
	}

	return groupedRemoteGroups, nil
}

func resourceTransitGatewayMulticastDomainAssociationsReadRaw(conn *ec2.EC2, domainID string) (map[string]interface{}, error) {
	remoteAssocSet, err := GetTransitGatewayMulticastDomainAssociations(conn, domainID)
	if err != nil {
		return nil, fmt.Errorf(
			"error getting EC2 Transit Gateway Multicast Domain (%s) associations: %s", domainID, err)
	}

	groupedAssocSet := make(map[string]interface{})
	for _, remoteAssoc := range remoteAssocSet {
		attachID := aws.StringValue(remoteAssoc.TransitGatewayAttachmentId)
		subnetID := aws.StringValue(remoteAssoc.Subnet.SubnetId)
		if c, exists := groupedAssocSet[attachID]; exists {
			container := c.(map[string]interface{})
			container["subnet_ids"].(*schema.Set).Add(subnetID)
			continue
		}
		newSet := &schema.Set{F: schema.HashString}
		newSet.Add(subnetID)
		groupedAssocSet[attachID] = resourceTransitGatewayMulticastDomainAssociationMake(attachID, newSet)
	}

	return groupedAssocSet, nil
}

func resourceTransitGatewayMulticastDomainAssociationsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	id := d.Id()

	if !d.HasChange("association") {
		return nil
	}

	o, n := d.GetChange("association")
	old := resourceTransitGatewayMulticastDomainAssociationsCompress(o.(*schema.Set).Difference(n.(*schema.Set)))
	nw := resourceTransitGatewayMulticastDomainAssociationsCompress(n.(*schema.Set).Difference(o.(*schema.Set)))

	if old.HashEqual(nw) {
		log.Printf(
			"[DEBUG] Flattened EC2 Transit Gateway Multicast Domain assoctiation configuration was " +
				"determined to be equivalent")
		d.Set("association", n)
		return nil
	}

	// disassociate old associations
	for _, assoc := range old.List() {
		assocData := assoc.(map[string]interface{})
		if err := resourceTransitGatewayMulticastDomainDisassociate(conn, id, assocData); err != nil {
			return err
		}
	}

	// save current state
	associations := o.(*schema.Set).Intersection(n.(*schema.Set))
	d.Set("association", associations)

	// associate new subnets
	for _, assoc := range nw.List() {
		assocData := assoc.(map[string]interface{})
		if err := resourceTransitGatewayMulticastDomainAssociate(conn, id, assocData); err != nil {
			return err
		}
		associations.Add(assoc)
		d.Set("association", associations)
	}

	return nil
}

func resourceTransitGatewayMulticastDomainGroupsUpdate(d *schema.ResourceData, meta interface{}, member bool) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	id := d.Id()
	key := ResourceTransitGatewayMulticastDomainGroupType(member)

	if !d.HasChange(key) {
		return nil
	}

	o, n := d.GetChange(key)
	old := resourceTransitGatewayMulticastDomainGroupsCompress(o.(*schema.Set).Difference(n.(*schema.Set)))
	nw := resourceTransitGatewayMulticastDomainGroupsCompress(n.(*schema.Set).Difference(o.(*schema.Set)))

	if old.HashEqual(nw) {
		log.Printf(
			"[DEBUG] Flattened EC2 Transit Gateway Multicast Domain group configuration was determined to be " +
				"equivalent")
		d.Set(key, n)
		return nil
	}

	// remove old groups
	for _, group := range old.List() {
		groupData := group.(map[string]interface{})
		if err := resourceTransitGatewayMulticastDomainGroupDeregister(conn, id, groupData, member); err != nil {
			return err
		}
	}

	// save current state
	groups := o.(*schema.Set).Intersection(n.(*schema.Set))
	d.Set(key, groups)

	// register new groups
	for _, group := range nw.List() {
		groupData := group.(map[string]interface{})
		if err := resourceTransitGatewayMulticastDomainGroupRegister(conn, id, groupData, member); err != nil {
			return err
		}
		groups.Add(group)
		d.Set(key, groups)
	}

	return nil
}

func resourceTransitGatewayMulticastDomainDisassociate(conn *ec2.EC2, id string, assocData map[string]interface{}) error {
	tgwAttachmentID := assocData["transit_gateway_attachment_id"].(string)
	subnetIDs := flex.ExpandStringSet(assocData["subnet_ids"].(*schema.Set))
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
	if err := WaitForTransitGatewayMulticastDomainDisassociation(conn, id, subnetIDs); err != nil {
		return fmt.Errorf(
			"error waiting for EC2 Transit Gateway Multicast Domain (%s) to disassociate subnets: %s",
			id, err)
	}

	return nil
}

func resourceTransitGatewayMulticastDomainAssociate(conn *ec2.EC2, id string, assocData map[string]interface{}) error {
	tgwAttachmentID := assocData["transit_gateway_attachment_id"].(string)
	subnetIDs := flex.ExpandStringSet(assocData["subnet_ids"].(*schema.Set))
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
	if err := WaitForTransitGatewayMulticastDomainAssociation(conn, id, subnetIDs); err != nil {
		return fmt.Errorf(
			"error waiting for EC2 Transit Gateway Multicast Domain (%s) associations: %s", id, err)
	}

	return nil
}

func resourceTransitGatewayMulticastDomainGroupDeregister(conn *ec2.EC2, id string, groupData map[string]interface{}, member bool) error {
	// Note: Search function returns both "members" and "sources"
	if member {
		input := &ec2.DeregisterTransitGatewayMulticastGroupMembersInput{
			GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
			NetworkInterfaceIds:             flex.ExpandStringSet(groupData["network_interface_ids"].(*schema.Set)),
			TransitGatewayMulticastDomainId: aws.String(id),
		}

		log.Printf(
			"[DEBUG] Deregistering members from EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
		if _, err := conn.DeregisterTransitGatewayMulticastGroupMembers(input); err != nil {
			return fmt.Errorf(
				"error removing member from EC2 Transit Gateway Multicast Domain (%s): %s", id, err)
		}
	} else {
		input := &ec2.DeregisterTransitGatewayMulticastGroupSourcesInput{
			GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
			NetworkInterfaceIds:             flex.ExpandStringSet(groupData["network_interface_ids"].(*schema.Set)),
			TransitGatewayMulticastDomainId: aws.String(id),
		}

		log.Printf(
			"[DEBUG] Deregistering sources from EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
		if _, err := conn.DeregisterTransitGatewayMulticastGroupSources(input); err != nil {
			return fmt.Errorf(
				"error removing source from EC2 Transit Gateway Multicast Domain (%s): %s", id, err)
		}
	}

	if err := WaitForTransitGatewayMulticastDomainGroupDeregister(conn, id, groupData, member); err != nil {
		return fmt.Errorf(
			"error validating EC2 Transit Gateway Multicast Domain (%s) group deregistration: %s", id, err)
	}

	return nil
}

func resourceTransitGatewayMulticastDomainGroupRegister(conn *ec2.EC2, id string, groupData map[string]interface{}, member bool) error {
	// Note: Search function returns both "members" and "sources"
	if member {
		input := &ec2.RegisterTransitGatewayMulticastGroupMembersInput{
			GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
			NetworkInterfaceIds:             flex.ExpandStringSet(groupData["network_interface_ids"].(*schema.Set)),
			TransitGatewayMulticastDomainId: aws.String(id),
		}

		log.Printf(
			"[DEBUG] Registering members to EC2 Transit Gateway Multifvcast Domain (%s): %s", id, input)
		if _, err := conn.RegisterTransitGatewayMulticastGroupMembers(input); err != nil {
			return fmt.Errorf(
				"error registering EC2 Transit Gateway Multicast Domain (%s) members: %s", id, err)
		}
	} else {
		multicastDomain, err := DescribeTransitGatewayMulticastDomain(conn, id)
		staticSourcesSupport := aws.StringValue(multicastDomain.Options.StaticSourcesSupport)

		if err != nil {
			return fmt.Errorf("error reading EC2 Transit Gateway Multicast Domain (%s): %s", id, err)
		}

		if staticSourcesSupport == "disable" {
			return fmt.Errorf("multicast domain %s does not have static sources - resource %v", id, multicastDomain)
		}

		input := &ec2.RegisterTransitGatewayMulticastGroupSourcesInput{
			GroupIpAddress:                  aws.String(groupData["group_ip_address"].(string)),
			NetworkInterfaceIds:             flex.ExpandStringSet(groupData["network_interface_ids"].(*schema.Set)),
			TransitGatewayMulticastDomainId: aws.String(id),
		}

		log.Printf(
			"[DEBUG] Registering sources to EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
		if _, err := conn.RegisterTransitGatewayMulticastGroupSources(input); err != nil {
			return fmt.Errorf(
				"error registering EC2 Transit Gateway Multicast Domain (%s) sources: %s", id, err)
		}
	}

	if err := WaitForTransitGatewayMulticastDomainGroupDeregister(conn, id, groupData, member); err != nil {
		return fmt.Errorf(
			"error validating EC2 Transit Gateway Multicast Domain (%s) group registration: %s", id, err)
	}

	return nil
}

func resourceTransitGatewayMulticastDomainDisassociateAll(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	id := d.Id()
	// disassociate subnets
	assocSet := d.Get("association").(*schema.Set)
	for _, assoc := range assocSet.List() {
		assocData := assoc.(map[string]interface{})
		if err := resourceTransitGatewayMulticastDomainDisassociate(conn, id, assocData); err != nil {
			return err
		}
	}
	return nil
}

func resourceTransitGatewayMulticastDomainGroupsDeregisterAll(d *schema.ResourceData, meta interface{}, member bool) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	id := d.Id()
	key := ResourceTransitGatewayMulticastDomainGroupType(member)
	// deregister groups
	groups := d.Get(key).(*schema.Set)
	for _, group := range groups.List() {
		groupData := group.(map[string]interface{})
		if err := resourceTransitGatewayMulticastDomainGroupDeregister(conn, id, groupData, member); err != nil {
			return err
		}
	}
	return nil
}

func resourceTransitGatewayMulticastDomainAssociationsHash(meta interface{}) int {
	m, castOk := meta.(map[string]interface{})
	if !castOk {
		return 0
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", m["transit_gateway_attachment_id"].(string)))
	for _, subnetID := range m["subnet_ids"].(*schema.Set).List() {
		buf.WriteString(fmt.Sprintf("%s,", subnetID.(string)))
	}
	return create.StringHashcode(buf.String())
}

func resourceTransitGatewayMulticastDomainGroupsHash(meta interface{}) int {
	m, castOk := meta.(map[string]interface{})
	if !castOk {
		return 0
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", m["group_ip_address"].(string)))
	for _, networkID := range m["network_interface_ids"].(*schema.Set).List() {
		buf.WriteString(fmt.Sprintf("%s,", networkID.(string)))
	}
	return create.StringHashcode(buf.String())
}

func ResourceTransitGatewayMulticastDomainGroupType(member bool) string {
	if member {
		return "members"
	}
	return "sources"
}

func resourceTransitGatewayMulticastDomainAssociationsCompress(assocSet *schema.Set) *schema.Set {
	return resourceTransitGatewayMulticastDomainCompress(
		assocSet, "transit_gateway_attachment_id", "subnet_ids",
		resourceTransitGatewayMulticastDomainAssociationsHash)
}

func resourceTransitGatewayMulticastDomainGroupsCompress(groupSet *schema.Set) *schema.Set {
	return resourceTransitGatewayMulticastDomainCompress(
		groupSet, "group_ip_address", "network_interface_ids",
		resourceTransitGatewayMulticastDomainGroupsHash)
}

func resourceTransitGatewayMulticastDomainCompress(set *schema.Set, groupingKey string, compressKey string, setFunc schema.SchemaSetFunc) *schema.Set {
	groups := make(map[string][]*string)
	for _, elem := range set.List() {
		data := elem.(map[string]interface{})
		groupVal := data[groupingKey].(string)
		val := flex.ExpandStringSet(data[compressKey].(*schema.Set))
		if existing, exists := groups[groupVal]; exists {
			groups[groupVal] = append(existing, val...)
			continue
		}
		groups[groupVal] = val
	}

	compressed := &schema.Set{F: setFunc}
	for group, values := range groups {
		data := make(map[string]interface{})
		data[groupingKey] = group
		data[compressKey] = flex.FlattenStringSet(values)
		compressed.Add(data)
	}

	return compressed
}

// IsMulticastAddress is a SchemaValidateFunc which tests if the provided value is of type string and a valid Multicast address
func IsMulticastAddress(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return warnings, errors
	}

	ip := net.ParseIP(v)
	if multicast := ip.IsMulticast(); multicast == false {
		errors = append(errors, fmt.Errorf("expected %s to contain a valid Multicast address, got: %s", k, v))
	}

	return warnings, errors
}
