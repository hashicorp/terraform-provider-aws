package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

func resourceAwsRoute53ZoneAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ZoneAssociationCreate,
		Read:   resourceAwsRoute53ZoneAssociationRead,
		Delete: resourceAwsRoute53ZoneAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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

			"owning_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsRoute53ZoneAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	r53 := meta.(*AWSClient).r53conn

	req := &route53.AssociateVPCWithHostedZoneInput{
		HostedZoneId: aws.String(d.Get("zone_id").(string)),
		VPC: &route53.VPC{
			VPCId:     aws.String(d.Get("vpc_id").(string)),
			VPCRegion: aws.String(meta.(*AWSClient).region),
		},
		Comment: aws.String("Managed by Terraform"),
	}
	if w := d.Get("vpc_region"); w != "" {
		req.VPC.VPCRegion = aws.String(w.(string))
	}

	log.Printf("[DEBUG] Associating Route53 Private Zone %s with VPC %s with region %s", *req.HostedZoneId, *req.VPC.VPCId, *req.VPC.VPCRegion)

	resp, err := r53.AssociateVPCWithHostedZone(req)
	if err != nil {
		return err
	}

	// Store association id
	d.SetId(fmt.Sprintf("%s:%s", *req.HostedZoneId, *req.VPC.VPCId))

	// Wait until we are done initializing
	wait := resource.StateChangeConf{
		Delay:      30 * time.Second,
		Pending:    []string{route53.ChangeStatusPending},
		Target:     []string{route53.ChangeStatusInsync},
		Timeout:    10 * time.Minute,
		MinTimeout: 2 * time.Second,
		Refresh:    resourceAwsRoute53ZoneAssociationRefreshFunc(r53, cleanChangeID(*resp.ChangeInfo.Id), d.Id()),
	}
	_, err = wait.WaitForState()
	if err != nil {
		return err
	}

	return resourceAwsRoute53ZoneAssociationRead(d, meta)
}

func resourceAwsRoute53ZoneAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	zoneID, vpcID, err := resourceAwsRoute53ZoneAssociationParseId(d.Id())
	vpcRegion := meta.(*AWSClient).region

	if err != nil {
		return err
	}

	hostedZoneSummary, err := route53GetZoneAssociation(conn, zoneID, vpcID, vpcRegion)

	if err != nil {
		return fmt.Errorf("error getting Route 53 Hosted Zone (%s): %s", zoneID, err)
	}

	if hostedZoneSummary == nil {
		log.Printf("[WARN] Route 53 Hosted Zone (%s) Association (%s) not found, removing from state", zoneID, vpcID)
		d.SetId("")
		return nil
	}

	d.Set("vpc_id", vpcID)
	d.Set("vpc_region", vpcRegion)
	d.Set("zone_id", zoneID)
	d.Set("owning_account", hostedZoneSummary.Owner.OwningAccount)

	return nil
}

func resourceAwsRoute53ZoneAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	zoneID, vpcID, err := resourceAwsRoute53ZoneAssociationParseId(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Disassociating Route 53 Hosted Zone (%s) Association: %s", zoneID, vpcID)

	req := &route53.DisassociateVPCFromHostedZoneInput{
		HostedZoneId: aws.String(zoneID),
		VPC: &route53.VPC{
			VPCId:     aws.String(vpcID),
			VPCRegion: aws.String(d.Get("vpc_region").(string)),
		},
		Comment: aws.String("Managed by Terraform"),
	}

	_, err = conn.DisassociateVPCFromHostedZone(req)

	if err != nil {
		return fmt.Errorf("error disassociating Route 53 Hosted Zone (%s) Association (%s): %s", zoneID, vpcID, err)
	}

	return nil
}

func resourceAwsRoute53ZoneAssociationParseId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected ZONEID:VPCID", id)
	}
	return parts[0], parts[1], nil
}

func resourceAwsRoute53ZoneAssociationRefreshFunc(conn *route53.Route53, changeId, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		changeRequest := &route53.GetChangeInput{
			Id: aws.String(changeId),
		}
		result, state, err := resourceAwsGoRoute53Wait(conn, changeRequest)
		if isAWSErr(err, "AccessDenied", "") {
			log.Printf("[WARN] AccessDenied when trying to get Route 53 change progress for %s - ignoring due to likely cross account issue", id)
			return true, route53.ChangeStatusInsync, nil
		}
		return result, state, err
	}
}

func route53GetZoneAssociation(conn *route53.Route53, zoneID, vpcID, vpcRegion string) (*route53.HostedZoneSummary, error) {
	input := &route53.ListHostedZonesByVPCInput{
		VPCId:     aws.String(vpcID),
		VPCRegion: aws.String(vpcRegion),
	}

	output, err := conn.ListHostedZonesByVPC(input)

	if err != nil {
		return nil, err
	}

	var associatedHostedZoneSummary *route53.HostedZoneSummary
	for _, hostedZoneSummary := range output.HostedZoneSummaries {
		if zoneID == aws.StringValue(hostedZoneSummary.HostedZoneId) {
			associatedHostedZoneSummary = hostedZoneSummary
			break
		}
	}

	return associatedHostedZoneSummary, nil
}
