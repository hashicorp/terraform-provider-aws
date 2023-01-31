package ssm

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceResourceDataSync() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceDataSyncCreate,
		ReadWithoutTimeout:   resourceResourceDataSyncRead,
		DeleteWithoutTimeout: resourceResourceDataSyncDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"s3_destination": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_arn": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"bucket_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"sync_format": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  ssm.ResourceDataSyncS3FormatJsonSerDe,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceResourceDataSyncCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	name := d.Get("name").(string)

	input := &ssm.CreateResourceDataSyncInput{
		S3Destination: expandResourceDataSyncS3Destination(d),
		SyncName:      aws.String(name),
	}

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateResourceDataSyncWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, ssm.ErrCodeResourceDataSyncInvalidConfigurationException, "S3 write failed for bucket") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.CreateResourceDataSyncWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Resource Data Sync (%s): %s", name, err)
	}

	d.SetId(name)
	return append(diags, resourceResourceDataSyncRead(ctx, d, meta)...)
}

func resourceResourceDataSyncRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	syncItem, err := FindResourceDataSyncItem(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Resource Data Sync (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Resource Data Sync (%s): %s", d.Id(), err)
	}

	d.Set("name", syncItem.SyncName)
	d.Set("s3_destination", flattenResourceDataSyncS3Destination(syncItem.S3Destination))
	return diags
}

func resourceResourceDataSyncDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	input := &ssm.DeleteResourceDataSyncInput{
		SyncName: aws.String(d.Id()),
	}

	_, err := conn.DeleteResourceDataSyncWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ssm.ErrCodeResourceDataSyncNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Resource Data Sync (%s): %s", d.Id(), err)
	}
	return diags
}

func FindResourceDataSyncItem(ctx context.Context, conn *ssm.SSM, name string) (*ssm.ResourceDataSyncItem, error) {
	var result *ssm.ResourceDataSyncItem
	input := &ssm.ListResourceDataSyncInput{}

	err := conn.ListResourceDataSyncPagesWithContext(ctx, input, func(page *ssm.ListResourceDataSyncOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.ResourceDataSyncItems {
			if aws.StringValue(item.SyncName) == name {
				result = item
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, &resource.NotFoundError{}
	}
	return result, nil
}

func flattenResourceDataSyncS3Destination(dest *ssm.ResourceDataSyncS3Destination) []interface{} {
	result := make(map[string]interface{})
	result["bucket_name"] = aws.StringValue(dest.BucketName)
	result["region"] = aws.StringValue(dest.Region)
	result["sync_format"] = aws.StringValue(dest.SyncFormat)
	if dest.AWSKMSKeyARN != nil {
		result["kms_key_arn"] = aws.StringValue(dest.AWSKMSKeyARN)
	}
	if dest.Prefix != nil {
		result["prefix"] = aws.StringValue(dest.Prefix)
	}
	return []interface{}{result}
}

func expandResourceDataSyncS3Destination(d *schema.ResourceData) *ssm.ResourceDataSyncS3Destination {
	raw := d.Get("s3_destination").([]interface{})[0].(map[string]interface{})
	s3dest := &ssm.ResourceDataSyncS3Destination{
		BucketName: aws.String(raw["bucket_name"].(string)),
		Region:     aws.String(raw["region"].(string)),
		SyncFormat: aws.String(raw["sync_format"].(string)),
	}
	if v, ok := raw["kms_key_arn"].(string); ok && v != "" {
		s3dest.AWSKMSKeyARN = aws.String(v)
	}
	if v, ok := raw["prefix"].(string); ok && v != "" {
		s3dest.Prefix = aws.String(v)
	}
	return s3dest
}
