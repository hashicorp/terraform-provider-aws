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

func ResourceVPCIpam() *schema.Resource {
	return &schema.Resource{
		Create:        resourceVPCIpamCreate,
		Read:          resourceVPCIpamRead,
		Update:        resourceVPCIpamUpdate,
		Delete:        resourceVPCIpamDelete,
		CustomizeDiff: customdiff.Sequence(verify.SetTagsDiff),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"operating_regions": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidRegionName,
						},
					},
				},
			},
			"private_default_scope_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_default_scope_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scope_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cascade": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	IpamStatusAvailable   = "Available"
	InvalidIpamIdNotFound = "InvalidIpamId.NotFound"
	IpamCreateTimeout     = 3 * time.Minute
	IpamCreateDeley       = 5 * time.Second
	IpamDeleteTimeout     = 3 * time.Minute
	IpamDeleteDelay       = 5 * time.Second
)

func resourceVPCIpamCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	current_region := meta.(*conns.AWSClient).Region
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateIpamInput{
		ClientToken:       aws.String(resource.UniqueId()),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, "ipam"),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	operatingRegions := d.Get("operating_regions").(*schema.Set).List()
	if !expandIpamOperatingRegionsContainsCurrentRegion(operatingRegions, current_region) {
		return fmt.Errorf("Must include (%s) as a operating_region", current_region)
	}
	input.OperatingRegions = expandIpamOperatingRegions(operatingRegions)

	log.Printf("[DEBUG] Creating IPAM: %s", input)
	output, err := conn.CreateIpam(input)
	if err != nil {
		return fmt.Errorf("Error creating ipam: %w", err)
	}
	d.SetId(aws.StringValue(output.Ipam.IpamId))
	log.Printf("[INFO] IPAM ID: %s", d.Id())

	if _, err = WaitIpamAvailable(conn, d.Id(), IpamCreateTimeout); err != nil {
		return fmt.Errorf("error waiting for IPAM (%s) to be Available: %w", d.Id(), err)
	}

	return resourceVPCIpamRead(d, meta)
}

func resourceVPCIpamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ipam, err := findIpamById(conn, d.Id())

	if err != nil && !tfawserr.ErrCodeEquals(err, InvalidIpamIdNotFound) {
		return err
	}

	if !d.IsNewResource() && ipam == nil {
		log.Printf("[WARN] IPAM (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", ipam.IpamArn)
	d.Set("description", ipam.Description)
	d.Set("operating_regions", flattenIpamOperatingRegions(ipam.OperatingRegions))
	d.Set("public_default_scope_id", ipam.PublicDefaultScopeId)
	d.Set("private_default_scope_id", ipam.PrivateDefaultScopeId)
	d.Set("scope_count", ipam.ScopeCount)

	tags := KeyValueTags(ipam.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceVPCIpamUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating IPAM (%s) tags: %w", d.Id(), err)
		}
	}

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyIpamInput{
			IpamId: aws.String(d.Id()),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("operating_regions") {
			o, n := d.GetChange("operating_regions")
			if o == nil {
				o = new(schema.Set)
			}
			if n == nil {
				n = new(schema.Set)
			}

			os := o.(*schema.Set)
			ns := n.(*schema.Set)
			operatingRegionUpdateAdd := expandIpamOperatingRegionsUpdateAddRegions(ns.Difference(os).List())
			operatingRegionUpdateRemove := expandIpamOperatingRegionsUpdateDeleteRegions(os.Difference(ns).List())

			if len(operatingRegionUpdateAdd) != 0 {
				input.AddOperatingRegions = operatingRegionUpdateAdd
			}

			if len(operatingRegionUpdateRemove) != 0 {
				input.RemoveOperatingRegions = operatingRegionUpdateRemove
			}
		}

		_, err := conn.ModifyIpam(input)

		if err != nil {
			return fmt.Errorf("error modifying IPAM (%s): %w", d.Id(), err)
		}
	}

	return nil
}

func resourceVPCIpamDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DeleteIpamInput{
		IpamId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("cascade"); ok {
		input.Cascade = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Deleting IPAM: %s", d.Id())
	_, err := conn.DeleteIpam(input)
	if err != nil {
		return fmt.Errorf("error deleting IPAM: (%s): %w", d.Id(), err)
	}

	if _, err = WaiterIpamDeleted(conn, d.Id(), IpamDeleteTimeout); err != nil {
		if tfawserr.ErrCodeEquals(err, InvalidIpamIdNotFound) {
			return nil
		}
		return fmt.Errorf("error waiting for IPAM (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func findIpamById(conn *ec2.EC2, id string) (*ec2.Ipam, error) {
	input := &ec2.DescribeIpamsInput{
		IpamIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeIpams(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Ipams) == 0 || output.Ipams[0] == nil {
		return nil, nil
	}

	return output.Ipams[0], nil
}

func WaitIpamAvailable(conn *ec2.EC2, ipamId string, timeout time.Duration) (*ec2.Ipam, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.IpamStateCreateInProgress},
		Target:  []string{ec2.IpamStateCreateComplete},
		Refresh: statusIpamStatus(conn, ipamId),
		Timeout: timeout,
		Delay:   IpamCreateDeley,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Ipam); ok {
		return output, err
	}

	return nil, err
}

func WaiterIpamDeleted(conn *ec2.EC2, ipamId string, timeout time.Duration) (*ec2.Ipam, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.IpamStateCreateComplete, ec2.IpamStateModifyComplete, ec2.IpamStateDeleteInProgress},
		Target:  []string{InvalidIpamIdNotFound},
		Refresh: statusIpamStatus(conn, ipamId),
		Timeout: timeout,
		Delay:   IpamDeleteDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Ipam); ok {
		return output, err
	}

	return nil, err
}

func statusIpamStatus(conn *ec2.EC2, ipamId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		output, err := findIpamById(conn, ipamId)

		if tfawserr.ErrCodeEquals(err, InvalidIpamIdNotFound) {
			return output, InvalidIpamIdNotFound, nil
		}

		// there was an unhandled error in the Finder
		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func expandIpamOperatingRegions(operatingRegions []interface{}) []*ec2.AddIpamOperatingRegion {
	regions := make([]*ec2.AddIpamOperatingRegion, 0, len(operatingRegions))
	for _, regionRaw := range operatingRegions {
		region := regionRaw.(map[string]interface{})
		regions = append(regions, expandIpamOperatingRegion(region))
	}

	return regions
}

func expandIpamOperatingRegion(operatingRegion map[string]interface{}) *ec2.AddIpamOperatingRegion {
	region := &ec2.AddIpamOperatingRegion{
		RegionName: aws.String(operatingRegion["region_name"].(string)),
	}
	return region
}

func flattenIpamOperatingRegions(operatingRegions []*ec2.IpamOperatingRegion) []interface{} {
	regions := []interface{}{}
	for _, operatingRegion := range operatingRegions {
		regions = append(regions, flattenIpamOperatingRegion(operatingRegion))
	}
	return regions
}

func flattenIpamOperatingRegion(operatingRegion *ec2.IpamOperatingRegion) map[string]interface{} {
	region := make(map[string]interface{})
	region["region_name"] = aws.StringValue(operatingRegion.RegionName)
	return region
}

func expandIpamOperatingRegionsUpdateAddRegions(operatingRegions []interface{}) []*ec2.AddIpamOperatingRegion {
	regionUpdates := make([]*ec2.AddIpamOperatingRegion, 0, len(operatingRegions))
	for _, regionRaw := range operatingRegions {
		region := regionRaw.(map[string]interface{})
		regionUpdates = append(regionUpdates, expandIpamOperatingRegionsUpdateAddRegion(region))
	}
	return regionUpdates
}

func expandIpamOperatingRegionsUpdateAddRegion(operatingRegion map[string]interface{}) *ec2.AddIpamOperatingRegion {
	regionUpdate := &ec2.AddIpamOperatingRegion{
		RegionName: aws.String(operatingRegion["region_name"].(string)),
	}
	return regionUpdate
}

func expandIpamOperatingRegionsUpdateDeleteRegions(operatingRegions []interface{}) []*ec2.RemoveIpamOperatingRegion {
	regionUpdates := make([]*ec2.RemoveIpamOperatingRegion, 0, len(operatingRegions))
	for _, regionRaw := range operatingRegions {
		region := regionRaw.(map[string]interface{})
		regionUpdates = append(regionUpdates, expandIpamOperatingRegionsUpdateDeleteRegion(region))
	}
	return regionUpdates
}

func expandIpamOperatingRegionsUpdateDeleteRegion(operatingRegion map[string]interface{}) *ec2.RemoveIpamOperatingRegion {
	regionUpdate := &ec2.RemoveIpamOperatingRegion{
		RegionName: aws.String(operatingRegion["region_name"].(string)),
	}
	return regionUpdate
}

func expandIpamOperatingRegionsContainsCurrentRegion(operatingRegions []interface{}, current_region string) bool {
	for _, regionRaw := range operatingRegions {
		region := regionRaw.(map[string]interface{})
		if region["region_name"].(string) == current_region {
			return true
		}
	}
	return false
}
