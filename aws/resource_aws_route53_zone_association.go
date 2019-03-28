package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

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
	var err error
	resp, err := r53.AssociateVPCWithHostedZone(req)
	if err != nil {
		return err
	}

	// Store association id
	d.SetId(fmt.Sprintf("%s:%s", *req.HostedZoneId, *req.VPC.VPCId))

	// Wait until we are done initializing
	wait := resource.StateChangeConf{
		Delay:      30 * time.Second,
		Pending:    []string{"PENDING"},
		Target:     []string{"INSYNC"},
		Timeout:    10 * time.Minute,
		MinTimeout: 2 * time.Second,
		Refresh: func() (result interface{}, state string, err error) {
			changeRequest := &route53.GetChangeInput{
				Id: aws.String(cleanChangeID(*resp.ChangeInfo.Id)),
			}
			return resourceAwsGoRoute53Wait(r53, changeRequest)
		},
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

	if err != nil {
		return err
	}

	vpc, err := route53GetZoneAssociation(conn, zoneID, vpcID)

	if isAWSErr(err, route53.ErrCodeNoSuchHostedZone, "") {
		log.Printf("[WARN] Route 53 Hosted Zone (%s) not found, removing from state", zoneID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Route 53 Hosted Zone (%s): %s", zoneID, err)
	}

	if vpc == nil {
		log.Printf("[WARN] Route 53 Hosted Zone (%s) Association (%s) not found, removing from state", zoneID, vpcID)
		d.SetId("")
		return nil
	}

	d.Set("vpc_id", vpc.VPCId)
	d.Set("vpc_region", vpc.VPCRegion)
	d.Set("zone_id", zoneID)

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

func route53GetZoneAssociation(conn *route53.Route53, zoneID, vpcID string) (*route53.VPC, error) {
	input := &route53.GetHostedZoneInput{
		Id: aws.String(zoneID),
	}

	output, err := conn.GetHostedZone(input)

	if err != nil {
		return nil, err
	}

	var vpc *route53.VPC
	for _, zoneVPC := range output.VPCs {
		if vpcID == aws.StringValue(zoneVPC.VPCId) {
			vpc = zoneVPC
			break
		}
	}

	return vpc, nil
}
