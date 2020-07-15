package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

func resourceAwsRoute53VPCAssociationAuthorization() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53VPCAssociationAuthorizationCreate,
		Read:   resourceAwsRoute53VPCAssociationAuthorizationRead,
		Delete: resourceAwsRoute53VPCAssociationAuthorizationDelete,

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpc_region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsRoute53VPCAssociationAuthorizationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	req := &route53.CreateVPCAssociationAuthorizationInput{
		HostedZoneId: aws.String(d.Get("zone_id").(string)),
		VPC: &route53.VPC{
			VPCId:     aws.String(d.Get("vpc_id").(string)),
			VPCRegion: aws.String(meta.(*AWSClient).region),
		},
	}
	if w := d.Get("vpc_region"); w != "" {
		req.VPC.VPCRegion = aws.String(w.(string))
	}

	log.Printf("[DEBUG] Creating Route53 VPC Association Authorization for hosted zone %s with VPC %s and region %s", *req.HostedZoneId, *req.VPC.VPCId, *req.VPC.VPCRegion)
	_, err := conn.CreateVPCAssociationAuthorization(req)
	if err != nil {
		return fmt.Errorf("Error creating Route53 VPC Association Authorization: %s", err)
	}

	// Store association id
	d.SetId(fmt.Sprintf("%s:%s", *req.HostedZoneId, *req.VPC.VPCId))
	d.Set("vpc_region", req.VPC.VPCRegion)

	return resourceAwsRoute53VPCAssociationAuthorizationRead(d, meta)
}

func resourceAwsRoute53VPCAssociationAuthorizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn
	zone_id, vpc_id := resourceAwsRoute53VPCAssociationAuthorizationParseId(d.Id())

	req := route53.ListVPCAssociationAuthorizationsInput{
		HostedZoneId: aws.String(zone_id),
	}
	for {
		log.Printf("[DEBUG] Listing Route53 VPC Association Authorizations for hosted zone %s", zone_id)
		res, err := conn.ListVPCAssociationAuthorizations(&req)

		if isAWSErr(err, route53.ErrCodeNoSuchHostedZone, "") {
			d.SetId("")
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error listing Route53 VPC Association Authorizations: %s", err)
		}

		for _, vpc := range res.VPCs {
			if vpc_id == *vpc.VPCId {
				return nil
			}
		}

		// Loop till we find our authorization or we reach the end
		if res.NextToken != nil {
			req.NextToken = res.NextToken
		} else {
			break
		}
	}

	// no association found
	d.SetId("")
	return nil
}

func resourceAwsRoute53VPCAssociationAuthorizationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn
	zone_id, vpc_id := resourceAwsRoute53VPCAssociationAuthorizationParseId(d.Id())

	req := route53.DeleteVPCAssociationAuthorizationInput{
		HostedZoneId: aws.String(zone_id),
		VPC: &route53.VPC{
			VPCId:     aws.String(vpc_id),
			VPCRegion: aws.String(d.Get("vpc_region").(string)),
		},
	}

	log.Printf("[DEBUG] Deleting Route53 Assocatiation Authorization for hosted zone %s for VPC %s", zone_id, vpc_id)
	_, err := conn.DeleteVPCAssociationAuthorization(&req)
	if err != nil {
		return fmt.Errorf("Error deleting Route53 VPC Association Authorization: %s", err)
	}

	return nil
}

func resourceAwsRoute53VPCAssociationAuthorizationParseId(id string) (zone_id, vpc_id string) {
	parts := strings.SplitN(id, ":", 2)
	zone_id = parts[0]
	vpc_id = parts[1]
	return
}
