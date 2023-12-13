// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfrontkeyvaluestore

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudfrontkeyvaluestore"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_cloudfront_keyvaluestore")
func ResourceKeyValueStore() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyValueStoreCreate,
		ReadWithoutTimeout:   resourceKeyValueStoreRead,
		UpdateWithoutTimeout: resourceKeyValueStoreUpdate,
		DeleteWithoutTimeout: resourceKeyValueStoreDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pair": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 512),
							),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 512),
							),
						},
					},
				},
			},
		},
	}
}

func resourceKeyValueStoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Create store.
	cloudfrontClient := meta.(*conns.AWSClient).CloudFrontConn(ctx)
	keyValueStoreName := d.Get("name").(string)
	input := &cloudfront.CreateKeyValueStoreInput{
		Name: aws.String(keyValueStoreName),
	}
	keyValueStoreOutput, _ := cloudfrontClient.CreateKeyValueStoreWithContext(ctx, input)
	d.SetId(keyValueStoreName)
	log.Printf("[INFO] Jeremy ETAG: %v", keyValueStoreOutput)

	// Add Key(s) / Value(s).
	cloudfrontKeyValueStoreClient := meta.(*conns.AWSClient).CloudFrontKeyValueStoreConn(ctx)
	keyValueInput := &cloudfrontkeyvaluestore.PutKeyInput{
		IfMatch: keyValueStoreOutput.ETag,
		Key:     aws.String("PouetKey"),
		Value:   aws.String("PouetValue"),
	}
	_, _ = cloudfrontKeyValueStoreClient.PutKeyWithContext(ctx, keyValueInput)

	return append(diags, resourceKeyValueStoreRead(ctx, d, meta)...)
}

func resourceKeyValueStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	output, err := findKeyValueStoreByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Key/Value store (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Key/Value store (%s): %s", d.Id(), err)
	}
	d.Set("id", output.KeyValueStore.Id)
	d.Set("name", output.KeyValueStore.Name)
	d.Set("arn", output.KeyValueStore.ARN)
	d.Set("etag", output.ETag)
	return diags
}

func resourceKeyValueStoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	print(conn)

	s := d.Get("name").(string)
	print(s)

	return diags
}

func resourceKeyValueStoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	log.Printf("[INFO] Deleting Key/Value store: %s", d.Id())

	etag := d.Get("etag").(string)
	log.Printf("[INFO] Jeremy: %s", etag)

	_, err := conn.DeleteKeyValueStoreWithContext(ctx, &cloudfront.DeleteKeyValueStoreInput{
		Name:    aws.String(d.Id()),
		IfMatch: aws.String(etag),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeEntityNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Key/Value store (%s): %s", d.Id(), err)
	}

	return diags
}

func findKeyValueStoreByID(ctx context.Context, conn *cloudfront.CloudFront, id string) (*cloudfront.DescribeKeyValueStoreOutput, error) {
	input := &cloudfront.DescribeKeyValueStoreInput{
		Name: aws.String(id),
	}

	output, err := conn.DescribeKeyValueStoreWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeEntityNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.KeyValueStore == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
