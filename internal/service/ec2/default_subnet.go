package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDefaultSubnet() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceDefaultSubnetCreate,
		Read:   resourceSubnetRead,
		Update: resourceSubnetUpdate,
		Delete: resourceDefaultSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		// Keep in sync with aws_subnet's schema with the following changes:
		//   - availability_zone is Required/ForceNew
		//   - availability_zone_id is Computed-only
		//   - cidr_block is Computed-only
		//   - outpost_arn is Computed-only
		//   - vpc_id is Computed-only
		// and additions:
		//   - existing_default_subnet Computed-only, set in resourceDefaultSubnetCreate
		//   - force_destroy Optional/Computed, default calculated via CustomizeDiff
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assign_ipv6_address_on_creation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_owned_ipv4_pool": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"map_customer_owned_ip_on_launch", "outpost_arn"},
			},
			"enable_dns64": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_resource_name_dns_aaaa_record_on_launch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_resource_name_dns_a_record_on_launch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"existing_default_subnet": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"ipv6_cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidIPv6CIDRNetworkAddress,
			},
			"ipv6_cidr_block_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_native": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"map_customer_owned_ip_on_launch": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"customer_owned_ipv4_pool", "outpost_arn"},
			},
			"map_public_ip_on_launch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_hostname_type_on_launch": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ec2.HostnameType_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDefaultSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	availabilityZone := d.Get("availability_zone").(string)
	input := &ec2.DescribeSubnetsInput{
		Filters: BuildAttributeFilterList(
			map[string]string{
				"availabilityZone": availabilityZone,
				"defaultForAz":     "true",
			},
		),
	}

	subnet, err := FindSubnet(conn, input)

	if err == nil {
		log.Printf("[INFO] Found existing EC2 Default Subnet (%s)", availabilityZone)
		d.SetId(aws.StringValue(subnet.SubnetId))
		d.Set("existing_default_subnet", true)
	} else if tfresource.NotFound(err) {
		input := &ec2.CreateDefaultSubnetInput{
			AvailabilityZone: aws.String(availabilityZone),
		}

		if v, ok := d.GetOk("ipv6_native"); ok {
			input.Ipv6Native = aws.Bool(v.(bool))
		}

		log.Printf("[DEBUG] Creating EC2 Default Subnet: %s", input)
		output, err := conn.CreateDefaultSubnet(input)

		if err != nil {
			return fmt.Errorf("error creating EC2 Default Subnet (%s): %w", availabilityZone, err)
		}

		subnet = output.Subnet

		d.SetId(aws.StringValue(subnet.SubnetId))
		d.Set("existing_default_subnet", false)

		_, err = WaitSubnetAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return fmt.Errorf("error waiting for EC2 Default Subnet (%s) to become available: %w", d.Id(), err)
		}
	} else {
		return fmt.Errorf("error reading EC2 Default Subnet (%s): %w", d.Id(), err)
	}

	// TODO Compare Subnet with desired configuration and update.

	if err := modifySubnetAttributesOnCreate(conn, d, subnet); err != nil {
		return err
	}

	return resourceSubnetRead(d, meta)
}

func resourceDefaultSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	if d.Get("force_destroy").(bool) {
		return resourceSubnetDelete(d, meta)
	}

	log.Printf("[WARN] EC2 Default Subnet (%s) not deleted, removing from state", d.Id())

	return nil
}
