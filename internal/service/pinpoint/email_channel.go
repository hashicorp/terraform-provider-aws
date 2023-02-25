package pinpoint

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEmailChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEmailChannelUpsert,
		ReadWithoutTimeout:   resourceEmailChannelRead,
		UpdateWithoutTimeout: resourceEmailChannelUpsert,
		DeleteWithoutTimeout: resourceEmailChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"configuration_set": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"from_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"identity": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"messages_per_second": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEmailChannelUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn()

	applicationId := d.Get("application_id").(string)

	params := &pinpoint.EmailChannelRequest{}

	params.Enabled = aws.Bool(d.Get("enabled").(bool))
	params.FromAddress = aws.String(d.Get("from_address").(string))
	params.Identity = aws.String(d.Get("identity").(string))

	if v, ok := d.GetOk("role_arn"); ok {
		params.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("configuration_set"); ok {
		params.ConfigurationSet = aws.String(v.(string))
	}

	req := pinpoint.UpdateEmailChannelInput{
		ApplicationId:       aws.String(applicationId),
		EmailChannelRequest: params,
	}

	_, err := conn.UpdateEmailChannelWithContext(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Pinpoint Email Channel for application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return append(diags, resourceEmailChannelRead(ctx, d, meta)...)
}

func resourceEmailChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn()

	log.Printf("[INFO] Reading Pinpoint Email Channel for application %s", d.Id())

	output, err := conn.GetEmailChannelWithContext(ctx, &pinpoint.GetEmailChannelInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
			log.Printf("[WARN] Pinpoint Email Channel for application %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "getting Pinpoint Email Channel for application %s: %s", d.Id(), err)
	}

	res := output.EmailChannelResponse
	d.Set("application_id", res.ApplicationId)
	d.Set("enabled", res.Enabled)
	d.Set("from_address", res.FromAddress)
	d.Set("identity", res.Identity)
	d.Set("role_arn", res.RoleArn)
	d.Set("configuration_set", res.ConfigurationSet)
	d.Set("messages_per_second", res.MessagesPerSecond)

	return diags
}

func resourceEmailChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn()

	log.Printf("[DEBUG] Deleting Pinpoint Email Channel for application %s", d.Id())
	_, err := conn.DeleteEmailChannelWithContext(ctx, &pinpoint.DeleteEmailChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint Email Channel for application %s: %s", d.Id(), err)
	}
	return diags
}
