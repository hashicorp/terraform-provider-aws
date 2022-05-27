package directconnect

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceHostedConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceHostedConnectionCreate,
		Read:   resourceHostedConnectionRead,
		Delete: resourceHostedConnectionDelete,

		Schema: map[string]*schema.Schema{
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
			"connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"has_logical_redundancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"lag_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"loa_issue_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"partner_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vlan": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
			},
		},
	}
}

func resourceHostedConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	name := d.Get("name").(string)
	input := &directconnect.AllocateHostedConnectionInput{
		Bandwidth:      aws.String(d.Get("bandwidth").(string)),
		ConnectionId:   aws.String(d.Get("connection_id").(string)),
		ConnectionName: aws.String(name),
		OwnerAccount:   aws.String(d.Get("owner_account_id").(string)),
		Vlan:           aws.Int64(int64(d.Get("vlan").(int))),
	}

	log.Printf("[DEBUG] Creating Direct Connect Hosted Connection: %s", input)
	output, err := conn.AllocateHostedConnection(input)

	if err != nil {
		return fmt.Errorf("error creating Direct Connect Hosted Connection (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ConnectionId))

	return resourceHostedConnectionRead(d, meta)
}

func resourceHostedConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	connection, err := FindHostedConnectionByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Hosted Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Hosted Connection (%s): %w", d.Id(), err)
	}

	// Cannot set the following attributes from the response:
	// - connection_id: conn.ConnectionId is this resource's ID, not the ID of the interconnect or LAG
	// - tags: conn.Tags seems to always come back empty and DescribeTags needs to be called from the owner account
	d.Set("aws_device", connection.AwsDeviceV2)
	d.Set("has_logical_redundancy", connection.HasLogicalRedundancy)
	d.Set("jumbo_frame_capable", connection.JumboFrameCapable)
	d.Set("lag_id", connection.LagId)
	d.Set("loa_issue_time", aws.TimeValue(connection.LoaIssueTime).Format(time.RFC3339))
	d.Set("location", connection.Location)
	d.Set("name", connection.ConnectionName)
	d.Set("owner_account_id", connection.OwnerAccount)
	d.Set("partner_name", connection.PartnerName)
	d.Set("provider_name", connection.ProviderName)
	d.Set("region", connection.Region)
	d.Set("state", connection.ConnectionState)
	d.Set("vlan", connection.Vlan)

	return nil
}

func resourceHostedConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	return deleteConnection(conn, d.Id(), waitHostedConnectionDeleted)
}
