package directconnect

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceGatewayCreate,
		Read:   resourceGatewayRead,
		Delete: resourceGatewayDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validAmazonSideASN,
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
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	name := d.Get("name").(string)
	input := &directconnect.CreateDirectConnectGatewayInput{
		DirectConnectGatewayName: aws.String(name),
	}

	if v, ok := d.Get("amazon_side_asn").(string); ok && v != "" {
		if v, err := strconv.ParseInt(v, 10, 64); err == nil {
			input.AmazonSideAsn = aws.Int64(v)
		}
	}

	log.Printf("[DEBUG] Creating Direct Connect Gateway: %s", input)
	resp, err := conn.CreateDirectConnectGateway(input)

	if err != nil {
		return fmt.Errorf("error creating Direct Connect Gateway (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(resp.DirectConnectGateway.DirectConnectGatewayId))

	if _, err := waitGatewayCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Direct Connect Gateway (%s) to create: %w", d.Id(), err)
	}

	return resourceGatewayRead(d, meta)
}

func resourceGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	output, err := FindGatewayByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Gateway (%s): %w", d.Id(), err)
	}

	d.Set("amazon_side_asn", strconv.FormatInt(aws.Int64Value(output.AmazonSideAsn), 10))
	d.Set("name", output.DirectConnectGatewayName)
	d.Set("owner_account_id", output.OwnerAccount)

	return nil
}

func resourceGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	log.Printf("[DEBUG] Deleting Direct Connect Gateway: %s", d.Id())
	_, err := conn.DeleteDirectConnectGateway(&directconnect.DeleteDirectConnectGatewayInput{
		DirectConnectGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "does not exist") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Direct Connect Gateway (%s): %w", d.Id(), err)
	}

	if _, err := waitGatewayDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Direct Connect Gateway (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
