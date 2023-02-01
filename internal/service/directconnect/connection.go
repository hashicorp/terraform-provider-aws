package directconnect

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionCreate,
		ReadWithoutTimeout:   resourceConnectionRead,
		UpdateWithoutTimeout: resourceConnectionUpdate,
		DeleteWithoutTimeout: resourceConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"request_macsec": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
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
			"skip_destroy": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vlan_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &directconnect.CreateConnectionInput{
		Bandwidth:      aws.String(d.Get("bandwidth").(string)),
		ConnectionName: aws.String(name),
		Location:       aws.String(d.Get("location").(string)),
		RequestMACSec:  aws.Bool(d.Get("request_macsec").(bool)),
	}

	if v, ok := d.GetOk("provider_name"); ok {
		input.ProviderName = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Direct Connect Connection: %s", input)
	output, err := conn.CreateConnectionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Connection (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ConnectionId))

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	connection, err := FindConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Connection (%s): %s", d.Id(), err)
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
	d.Set("vlan_id", connection.Vlan)

	// d.Set("request_macsec", d.Get("request_macsec").(bool))

	if !d.IsNewResource() && !d.Get("request_macsec").(bool) {
		d.Set("request_macsec", aws.Bool(false))
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Direct Connect Connection (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()

	// Update encryption mode
	if d.HasChange("encryption_mode") {
		input := &directconnect.UpdateConnectionInput{
			ConnectionId:   aws.String(d.Id()),
			EncryptionMode: aws.String(d.Get("encryption_mode").(string)),
		}
		log.Printf("[DEBUG] Modifying Direct Connect connection attributes: %s", input)
		_, err := conn.UpdateConnectionWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Direct Connect connection (%s) attributes: %s", d.Id(), err)
		}

		if _, err := waitConnectionConfirmed(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect connection (%s) to become available: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Direct Connect Connection (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if v, ok := d.GetOk("skip_destroy"); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Direct Connect Connection: %s", d.Id())
		return diags
	}

	conn := meta.(*conns.AWSClient).DirectConnectConn()

	if err := deleteConnection(ctx, conn, d.Id(), waitConnectionDeleted); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect Connection (%s): %s", d.Id(), err)
	}
	return diags
}

func deleteConnection(ctx context.Context, conn *directconnect.DirectConnect, connectionID string, waiter func(context.Context, *directconnect.DirectConnect, string) (*directconnect.Connection, error)) error {
	log.Printf("[DEBUG] Deleting Direct Connect Connection: %s", connectionID)
	_, err := conn.DeleteConnectionWithContext(ctx, &directconnect.DeleteConnectionInput{
		ConnectionId: aws.String(connectionID),
	})

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Connection with ID") {
		return nil
	}

	if err != nil {
		return err
	}

	_, err = waiter(ctx, conn, connectionID)

	if err != nil {
		return fmt.Errorf("wating for completion: %w", err)
	}

	return nil
}
