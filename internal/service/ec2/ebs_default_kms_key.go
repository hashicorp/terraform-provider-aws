package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEBSDefaultKMSKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSDefaultKMSKeyCreate,
		ReadWithoutTimeout:   resourceEBSDefaultKMSKeyRead,
		DeleteWithoutTimeout: resourceEBSDefaultKMSKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"key_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceEBSDefaultKMSKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	resp, err := conn.ModifyEbsDefaultKmsKeyIdWithContext(ctx, &ec2.ModifyEbsDefaultKmsKeyIdInput{
		KmsKeyId: aws.String(d.Get("key_arn").(string)),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EBS default KMS key: %s", err)
	}

	d.SetId(aws.StringValue(resp.KmsKeyId))

	return append(diags, resourceEBSDefaultKMSKeyRead(ctx, d, meta)...)
}

func resourceEBSDefaultKMSKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	resp, err := conn.GetEbsDefaultKmsKeyIdWithContext(ctx, &ec2.GetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS default KMS key: %s", err)
	}

	d.Set("key_arn", resp.KmsKeyId)

	return diags
}

func resourceEBSDefaultKMSKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	_, err := conn.ResetEbsDefaultKmsKeyIdWithContext(ctx, &ec2.ResetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EBS default KMS key: %s", err)
	}

	return diags
}
