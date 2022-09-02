package directconnect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceConnectionCreate,
		Read:   resourceConnectionRead,
		Update: resourceConnectionUpdate,
		Delete: resourceConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validConnectionBandWidth(),
			},
			// The MAC Security (MACsec) connection encryption mode.
			"encryption_mode": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"no_encrypt", "should_encrypt", "must_encrypt"}, false),
			},
			"has_logical_redundancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// Indicates whether the connection supports MAC Security (MACsec).
			"macsec_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			// Enable or disable MAC Security (MACsec) on this connection.
			"macsec_requested": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			// The MAC Security (MACsec) security keys associated with the connection.
			"macsec_keys": {
				// Slice of type MacSecKey
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"secret_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ckn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"start_on": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// The MAC Security (MACsec) port link status of the connection.
			"port_encryption_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &directconnect.CreateConnectionInput{
		Bandwidth:      aws.String(d.Get("bandwidth").(string)),
		ConnectionName: aws.String(name),
		Location:       aws.String(d.Get("location").(string)),
		RequestMACSec:  aws.Bool(d.Get("macsec_requested").(bool)),
	}

	if v, ok := d.GetOk("provider_name"); ok {
		input.ProviderName = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Direct Connect Connection: %s", input)
	output, err := conn.CreateConnection(input)

	if err != nil {
		return fmt.Errorf("error creating Direct Connect Connection (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ConnectionId))
	d.Set("macsec_requested", d.Get("macsec_requested").(bool))

	return resourceConnectionRead(d, meta)
}

func resourceConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	connection, err := FindConnectionByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Connection (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    aws.StringValue(connection.Region),
		Service:   "directconnect",
		AccountID: aws.StringValue(connection.OwnerAccount),
		Resource:  fmt.Sprintf("dxcon/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("aws_device", connection.AwsDeviceV2)
	d.Set("bandwidth", connection.Bandwidth)
	d.Set("encryption_mode", connection.EncryptionMode)
	d.Set("has_logical_redundancy", connection.HasLogicalRedundancy)
	d.Set("jumbo_frame_capable", connection.JumboFrameCapable)
	d.Set("location", connection.Location)
	d.Set("macsec_capable", connection.MacSecCapable)
	d.Set("name", connection.ConnectionName)
	d.Set("owner_account_id", connection.OwnerAccount)
	d.Set("port_encryption_status", connection.PortEncryptionStatus)
	d.Set("provider_name", connection.ProviderName)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Direct Connect Connection (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	// Update encryption mode
	if d.HasChange("encryption_mode") {
		input := &directconnect.UpdateConnectionInput{
			ConnectionId:   aws.String(d.Id()),
			EncryptionMode: aws.String(d.Get("encryption_mode").(string)),
		}
		log.Printf("[DEBUG] Modifying Direct Connect connection attributes: %s", input)
		_, err := conn.UpdateConnection(input)
		if err != nil {
			return fmt.Errorf("error modifying Direct Connect connection (%s) attributes: %s", d.Id(), err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Direct Connect Connection (%s) tags: %w", arn, err)
		}
	}

	return resourceConnectionRead(d, meta)
}

func resourceConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	return deleteConnection(conn, d.Id(), waitConnectionDeleted)
}

func deleteConnection(conn *directconnect.DirectConnect, connectionID string, waiter func(*directconnect.DirectConnect, string) (*directconnect.Connection, error)) error {
	log.Printf("[DEBUG] Deleting Direct Connect Connection: %s", connectionID)
	_, err := conn.DeleteConnection(&directconnect.DeleteConnectionInput{
		ConnectionId: aws.String(connectionID),
	})

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Connection with ID") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Direct Connect Connection (%s): %w", connectionID, err)
	}

	_, err = waiter(conn, connectionID)

	if err != nil {
		return fmt.Errorf("error waiting for Direct Connect Connection (%s) delete: %w", connectionID, err)
	}

	return nil
}

// ValidateMacSecAvailability checks for MAC Security availability given a
// location `l` and desired port speed `p`
func ValidateMacSecAvailability(l string, p string, meta interface{}) bool {
	conn := meta.(*conns.AWSClient).DirectConnectConn
	input := &directconnect.DescribeLocationsInput{}

	response, err := conn.DescribeLocations(input)

	if err != nil {
		fmt.Printf("error checking MACSec availability: %s", err)
		return false
	}

	var available bool

	for _, location := range response.Locations {
		if l == *location.LocationCode {
			for _, port_speed := range location.AvailableMacSecPortSpeeds {
				if *port_speed == p {
					available = true
				} else {
					fmt.Printf("MACSec not available for location %s and desired port speed %s", l, p)
					available = false
				}
			}
		}
	}
	return available
}
