package aws

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/route53"
)

func resourceAwsRoute53ZoneAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ZoneAssociationCreate,
		Read:   resourceAwsRoute53ZoneAssociationRead,
		Update: resourceAwsRoute53ZoneAssociationUpdate,
		Delete: resourceAwsRoute53ZoneAssociationDelete,

		Schema: map[string]*schema.Schema{
			"zone_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"vpc_region": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceAwsRoute53ZoneAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	zoneId := d.Get("zone_id").(string)
	vpcId := d.Get("vpc_id").(string)
	region := meta.(*AWSClient).region
	if w := d.Get("vpc_region"); w != "" {
		region = w.(string)
	}

	if err := associateRoute53ZoneWithVpc(zoneId, vpcId, region, meta); err != nil {
		return err
	}

	// Store association id
	d.SetId(fmt.Sprintf("%s:%s", zoneId, vpcId))
	d.Set("vpc_region", region)

	return resourceAwsRoute53ZoneAssociationUpdate(d, meta)
}

func resourceAwsRoute53ZoneAssociationRead(d *schema.ResourceData, meta interface{}) error {
	r53 := meta.(*AWSClient).r53conn
	zone_id, vpc_id := resourceAwsRoute53ZoneAssociationParseId(d.Id())
	zone, err := r53.GetHostedZone(&route53.GetHostedZoneInput{Id: aws.String(zone_id)})
	if err != nil {
		// Handle a deleted zone
		if r53err, ok := err.(awserr.Error); ok && r53err.Code() == "NoSuchHostedZone" {
			d.SetId("")
			return nil
		}
		return err
	}

	for _, vpc := range zone.VPCs {
		if vpc_id == *vpc.VPCId {
			// association is there, return
			return nil
		}
	}

	// no association found
	d.SetId("")
	return nil
}

func resourceAwsRoute53ZoneAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsRoute53ZoneAssociationRead(d, meta)
}

func resourceAwsRoute53ZoneAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	zone_id, vpc_id := resourceAwsRoute53ZoneAssociationParseId(d.Id())

	if err := dissociateVpcFromRoute53Zone(zone_id, vpc_id, d.Get("vpc_region").(string), meta); err != nil {
		return err
	}

	return nil
}

func resourceAwsRoute53ZoneAssociationParseId(id string) (zone_id, vpc_id string) {
	parts := strings.SplitN(id, ":", 2)
	zone_id = parts[0]
	vpc_id = parts[1]
	return
}

func associateRoute53ZoneWithVpc(zoneId string, vpcId string, vpcRegion string, meta interface{}) error {
	r53 := meta.(*AWSClient).r53conn

	req := &route53.AssociateVPCWithHostedZoneInput{
		HostedZoneId: aws.String(zoneId),
		VPC: &route53.VPC{
			VPCId:     aws.String(vpcId),
			VPCRegion: aws.String(vpcRegion),
		},
		Comment: aws.String("Managed by Terraform"),
	}

	log.Printf("[DEBUG] Associating Route53 Private Zone %s with VPC %s with region %s", *req.HostedZoneId, *req.VPC.VPCId, *req.VPC.VPCRegion)
	var err error
	resp, err := r53.AssociateVPCWithHostedZone(req)
	if err != nil {
		return err
	}

	if err := waitForRoute53RecordSetToSync(r53, cleanChangeID(*resp.ChangeInfo.Id)); err != nil {
		return err
	}

	return nil
}

func dissociateVpcFromRoute53Zone(zoneId string, vpcId string, vpcRegion string, meta interface{}) error {
	r53 := meta.(*AWSClient).r53conn

	req := &route53.DisassociateVPCFromHostedZoneInput{
		HostedZoneId: aws.String(zoneId),
		VPC: &route53.VPC{
			VPCId:     aws.String(vpcId),
			VPCRegion: aws.String(vpcRegion),
		},
		Comment: aws.String("Managed by Terraform"),
	}

	log.Printf("[DEBUG] Dissociating VPC %s with region %s from Private Zone %s", *req.VPC.VPCId, *req.VPC.VPCRegion, *req.HostedZoneId)
	var err error
	resp, err := r53.DisassociateVPCFromHostedZone(req)
	if err != nil {
		return err
	}

	if err := waitForRoute53RecordSetToSync(r53, cleanChangeID(*resp.ChangeInfo.Id)); err != nil {
		return err
	}

	return nil
}
