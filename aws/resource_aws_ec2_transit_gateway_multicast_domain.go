package aws

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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
							ForceNew: true,
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

func resourceAwsEc2TransitGatewayMulticastDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Ec2UpdateTags(conn, id, o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Multicast Domain (%s) tags: %s", id, err)
		}
	}

	// check if association set has changed
	if d.HasChange("association") {
		o, n := d.GetChange("association")
		old := o.(*schema.Set).Difference(n.(*schema.Set))
		nw := n.(*schema.Set).Difference(o.(*schema.Set))

		// disassociate old associations
		for _, assoc := range old.List() {
			assocMap := assoc.(map[string]interface{})
			tgwAttachmentID := assocMap["transit_gateway_attachment_id"].(string)
			subnetIDs := expandStringSet(assocMap["subnet_ids"].(*schema.Set))
			input := &ec2.DisassociateTransitGatewayMulticastDomainInput{
				SubnetIds:                       subnetIDs,
				TransitGatewayAttachmentId:      aws.String(tgwAttachmentID),
				TransitGatewayMulticastDomainId: aws.String(id),
			}

			log.Printf("Disassociating subnets from EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
			_, err := conn.DisassociateTransitGatewayMulticastDomain(input)
			if err != nil {
				return fmt.Errorf(
					"error disassociating EC2 Transit Gateway Multicast Domain (%s) subnets: %s", id, err)
			}

			// wait for subnets to be disassociated
			if err := waitForEc2TransitGatewayMulticastDomainDisassociation(conn, id, subnetIDs); err != nil {
				return fmt.Errorf(
					"error waiting for EC2 Transit Gateway Multicast Domain (%s) to disassociate subnets: %s",
					id, err)
			}
		}

		// save current state
		associations := o.(*schema.Set).Intersection(n.(*schema.Set))
		d.Set("association", associations)

		// associate new subnets
		for _, assoc := range nw.List() {
			assocMap := assoc.(map[string]interface{})
			tgwAttachmentID := assocMap["transit_gateway_attachment_id"].(string)
			subnetIDs := expandStringSet(assocMap["subnet_ids"].(*schema.Set))
			input := &ec2.AssociateTransitGatewayMulticastDomainInput{
				SubnetIds:                       subnetIDs,
				TransitGatewayAttachmentId:      aws.String(tgwAttachmentID),
				TransitGatewayMulticastDomainId: aws.String(id),
			}

			log.Printf(
				"[DEBUG] Associating subnets to EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
			_, err := conn.AssociateTransitGatewayMulticastDomain(input)
			if err != nil {
				return fmt.Errorf(
					"error associating EC2 Transit Gateway Multicast Domain (%s) subnets: %s", id, err)
			}

			// wait for associations
			if err := waitForEc2TransitGatewayMulticastDomainAssociation(conn, id, subnetIDs); err != nil {
				return fmt.Errorf(
					"error waiting for EC2 Transit Gateway Multicast Domain (%s) associations: %s", id, err)
			}

			associations.Add(assoc)
			d.Set("association", associations)
		}
	}

	return resourceAwsEc2TransitGatewayMulticastDomainRead(d, meta)
}

func resourceAwsEc2TransitGatewayMulticastDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	id := d.Id()
	err := resourceAwsEc2TransitGatewayMulticastDomainRead(d, meta)
	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Multicast Domain (%s): %s", id, err)
	}

	// disassociate subnets
	assocSet := d.Get("association").(*schema.Set)
	for _, assoc := range assocSet.List() {
		assocData := assoc.(map[string]interface{})
		subnetIDs := expandStringSet(assocData["subnet_ids"].(*schema.Set))
		input := &ec2.DisassociateTransitGatewayMulticastDomainInput{
			SubnetIds:                       subnetIDs,
			TransitGatewayAttachmentId:      aws.String(assocData["transit_gateway_attachment_id"].(string)),
			TransitGatewayMulticastDomainId: aws.String(id),
		}
		log.Printf("[DEBUG] Disassociating EC2 Transit Gateway Multicast Domain (%s) subnets: %s", id, input)
		_, err := conn.DisassociateTransitGatewayMulticastDomain(input)
		if err != nil {
			return fmt.Errorf(
				"error disassociating EC2 Transit Gateway Multicast Domain (%s) subnets: %s", id, err)
		}

		// wait for subnets to disassociate
		if err := waitForEc2TransitGatewayMulticastDomainDisassociation(conn, id, subnetIDs); err != nil {
			return fmt.Errorf(
				"error while waiting for EC2 Transit Gateway Multicast Domain (%s) subnets to disassociate: %s",
				id, err)
		}

	}

	input := &ec2.DeleteTransitGatewayMulticastDomainInput{
		TransitGatewayMulticastDomainId: aws.String(id),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteTransitGatewayMulticastDomain(input)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteTransitGatewayMulticastDomain(input)
	}

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
