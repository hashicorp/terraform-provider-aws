package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPAMResourceDiscoveryAssociation() *schema.Resource {
	return &schema.Resource{
		Create:        resourceIPAMResourceDiscoveryAssociationCreate,
		Read:          resourceIPAMResourceDiscoveryAssociationRead,
		Update:        resourceIPAMResourceDiscoveryAssociationUpdate,
		Delete:        resourceIPAMResourceDiscoveryAssociationDelete,
		CustomizeDiff: customdiff.Sequence(verify.SetTagsDiff),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipam_resource_discovery_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	invalidIPAMResourceDiscoveryAssociationIDNotFound = "InvalidIpamResourceDiscoveryAssociationId.NotFound"
	ipamResourceDiscoveryAssociationCreateTimeout     = 3 * time.Minute
	ipamResourceDiscoveryAssociationCreateDelay       = 5 * time.Second
	IPAMResourceDiscoveryAssociationDeleteTimeout     = 3 * time.Minute
	ipamResourceDiscoveryAssociationDeleteDelay       = 5 * time.Second
)

func resourceIPAMResourceDiscoveryAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.AssociateIpamResourceDiscoveryInput{
		ClientToken:             aws.String(resource.UniqueId()),
		TagSpecifications:       tagSpecificationsFromKeyValueTags(tags, "ipam-resource-discovery-association"),
		IpamId:                  aws.String(d.Get("ipam_id").(string)),
		IpamResourceDiscoveryId: aws.String(d.Get("ipam_resource_discovery_id").(string)),
	}

	log.Printf("[DEBUG] Creating IPAM Resource Discovery Association: %s", input)
	output, err := conn.AssociateIpamResourceDiscovery(input)
	if err != nil {
		return fmt.Errorf("Error associating ipam resource discovery: %w", err)
	}
	d.SetId(aws.StringValue(output.IpamResourceDiscoveryAssociation.IpamResourceDiscoveryAssociationId))
	log.Printf("[INFO] IPAM Resource Discovery Association ID: %s", d.Id())

	if _, err = WaitIPAMResourceDiscoveryAssociationAvailable(conn, d.Id(), ipamResourceDiscoveryAssociationCreateTimeout); err != nil {
		return fmt.Errorf("error waiting for IPAM Resource Discovery Association (%s) to be Available: %w", d.Id(), err)
	}

	return resourceIPAMResourceDiscoveryAssociationRead(d, meta)
}

func resourceIPAMResourceDiscoveryAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	rd, err := findIPAMResourceDiscoveryAssociationById(conn, d.Id())

	if err != nil && !tfawserr.ErrCodeEquals(err, invalidIPAMResourceDiscoveryAssociationIDNotFound) {
		return err
	}

	if !d.IsNewResource() && rd == nil {
		log.Printf("[WARN] IPAM Resource Discovery Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", rd.IpamResourceDiscoveryAssociationArn)
	d.Set("owner_id", rd.OwnerId)
	d.Set("status", rd.Status)
	d.Set("ipam_id", rd.IpamId)
	d.Set("ipam_resource_discovery_id", rd.IpamResourceDiscoveryId)
	d.Set("is_default", rd.IsDefault)

	tags := KeyValueTags(rd.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceIPAMResourceDiscoveryAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating IPAM ResourceDiscovery Association (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceIPAMResourceDiscoveryAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DisassociateIpamResourceDiscoveryInput{
		IpamResourceDiscoveryAssociationId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Disassociating IPAM Resource Discovery: %s", d.Id())
	_, err := conn.DisassociateIpamResourceDiscovery(input)
	if err != nil {
		return fmt.Errorf("error disassociating IPAM Resource Discovery: (%s): %w", d.Id(), err)
	}

	if _, err = WaiterIPAMResourceDiscoveryAssociationDeleted(conn, d.Id(), IPAMResourceDiscoveryAssociationDeleteTimeout); err != nil {
		if tfawserr.ErrCodeEquals(err, invalidIPAMResourceDiscoveryAssociationIDNotFound) {
			return nil
		}
		return fmt.Errorf("error waiting for IPAM Resource Discovery Association (%s) to be dissociated: %w", d.Id(), err)
	}

	return nil
}

func findIPAMResourceDiscoveryAssociationById(conn *ec2.EC2, id string) (*ec2.IpamResourceDiscoveryAssociation, error) {
	input := &ec2.DescribeIpamResourceDiscoveryAssociationsInput{
		IpamResourceDiscoveryAssociationIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeIpamResourceDiscoveryAssociations(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.IpamResourceDiscoveryAssociations) == 0 || output.IpamResourceDiscoveryAssociations[0] == nil {
		return nil, nil
	}

	return output.IpamResourceDiscoveryAssociations[0], nil
}

func WaitIPAMResourceDiscoveryAssociationAvailable(conn *ec2.EC2, rdId string, timeout time.Duration) (*ec2.Ipam, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.IpamResourceDiscoveryAssociationStateAssociateInProgress},
		Target:  []string{ec2.IpamResourceDiscoveryAssociationStateAssociateComplete},
		Refresh: statusIPAMResourceDiscoveryAssociationStatus(conn, rdId),
		Timeout: timeout,
		Delay:   ipamResourceDiscoveryAssociationCreateDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Ipam); ok {
		return output, err
	}

	return nil, err
}

func WaiterIPAMResourceDiscoveryAssociationDeleted(conn *ec2.EC2, rdId string, timeout time.Duration) (*ec2.Ipam, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.IpamResourceDiscoveryAssociationStateAssociateComplete, ec2.IpamResourceDiscoveryAssociationStateDisassociateInProgress},
		Target:  []string{invalidIPAMResourceDiscoveryAssociationIDNotFound, ec2.IpamResourceDiscoveryAssociationStateDisassociateComplete},
		Refresh: statusIPAMResourceDiscoveryAssociationStatus(conn, rdId),
		Timeout: timeout,
		Delay:   ipamResourceDiscoveryAssociationDeleteDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Ipam); ok {
		return output, err
	}

	return nil, err
}

func statusIPAMResourceDiscoveryAssociationStatus(conn *ec2.EC2, rdId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		output, err := findIPAMResourceDiscoveryAssociationById(conn, rdId)

		if tfawserr.ErrCodeEquals(err, invalidIPAMResourceDiscoveryAssociationIDNotFound) {
			return output, invalidIPAMResourceDiscoveryAssociationIDNotFound, nil
		}

		// there was an unhandled error in the Finder
		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}
