package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceEBSEncryptionByDefault() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSEncryptionByDefaultCreate,
		ReadWithoutTimeout:   resourceEBSEncryptionByDefaultRead,
		UpdateWithoutTimeout: resourceEBSEncryptionByDefaultUpdate,
		DeleteWithoutTimeout: resourceEBSEncryptionByDefaultDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceEBSEncryptionByDefaultCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	enabled := d.Get("enabled").(bool)
	if err := setEBSEncryptionByDefault(ctx, conn, enabled); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EBS encryption by default (%t): %s", enabled, err)
	}

	//lintignore:R015 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.UniqueId())

	return append(diags, resourceEBSEncryptionByDefaultRead(ctx, d, meta)...)
}

func resourceEBSEncryptionByDefaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	resp, err := conn.GetEbsEncryptionByDefaultWithContext(ctx, &ec2.GetEbsEncryptionByDefaultInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS encryption by default: %s", err)
	}

	d.Set("enabled", resp.EbsEncryptionByDefault)

	return diags
}

func resourceEBSEncryptionByDefaultUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	enabled := d.Get("enabled").(bool)
	if err := setEBSEncryptionByDefault(ctx, conn, enabled); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EBS encryption by default (%t): %s", enabled, err)
	}

	return append(diags, resourceEBSEncryptionByDefaultRead(ctx, d, meta)...)
}

func resourceEBSEncryptionByDefaultDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	// Removing the resource disables default encryption.
	if err := setEBSEncryptionByDefault(ctx, conn, false); err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling EBS encryption by default: %s", err)
	}

	return diags
}

func setEBSEncryptionByDefault(ctx context.Context, conn *ec2.EC2, enabled bool) error {
	var err error

	if enabled {
		_, err = conn.EnableEbsEncryptionByDefaultWithContext(ctx, &ec2.EnableEbsEncryptionByDefaultInput{})
	} else {
		_, err = conn.DisableEbsEncryptionByDefaultWithContext(ctx, &ec2.DisableEbsEncryptionByDefaultInput{})
	}

	return err
}
