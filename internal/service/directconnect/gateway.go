package directconnect

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGatewayCreate,
		ReadWithoutTimeout:   resourceGatewayRead,
		DeleteWithoutTimeout: resourceGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()

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
	resp, err := conn.CreateDirectConnectGatewayWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Gateway (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(resp.DirectConnectGateway.DirectConnectGatewayId))

	if _, err := waitGatewayCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway (%s) to create: %s", d.Id(), err)
	}

	return append(diags, resourceGatewayRead(ctx, d, meta)...)
}

func resourceGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()

	output, err := FindGatewayByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Gateway (%s): %s", d.Id(), err)
	}

	d.Set("amazon_side_asn", strconv.FormatInt(aws.Int64Value(output.AmazonSideAsn), 10))
	d.Set("name", output.DirectConnectGatewayName)
	d.Set("owner_account_id", output.OwnerAccount)

	return diags
}

func resourceGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()

	log.Printf("[DEBUG] Deleting Direct Connect Gateway: %s", d.Id())
	_, err := conn.DeleteDirectConnectGatewayWithContext(ctx, &directconnect.DeleteDirectConnectGatewayInput{
		DirectConnectGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect Gateway (%s): %s", d.Id(), err)
	}

	if _, err := waitGatewayDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway (%s) to delete: %s", d.Id(), err)
	}

	return diags
}
