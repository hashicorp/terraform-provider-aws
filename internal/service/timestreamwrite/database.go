// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_timestreamwrite_database", name="Database")
// @Tags(identifierAttribute="arn")
func ResourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDatabaseCreate,
		ReadWithoutTimeout:   resourceDatabaseRead,
		UpdateWithoutTimeout: resourceDatabaseUpdate,
		DeleteWithoutTimeout: resourceDatabaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				// The Timestream API accepts the KmsKeyId as an ID, ARN, alias, or alias ARN but always returns the ARN of the key.
				// The ARN is of the format 'arn:aws:kms:REGION:ACCOUNT_ID:key/KMS_KEY_ID'. Appropriate diff suppression
				// would require an extra API call to the kms service's DescribeKey method to decipher aliases.
				// To avoid importing an extra service in this resource, input here is restricted to only ARNs.
				ValidateFunc: verify.ValidARN,
			},

			"table_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TimestreamWriteConn(ctx)

	dbName := d.Get("database_name").(string)
	input := &timestreamwrite.CreateDatabaseInput{
		DatabaseName: aws.String(dbName),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	resp, err := conn.CreateDatabaseWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Timestream Database (%s): %s", dbName, err)
	}

	if resp == nil || resp.Database == nil {
		return diag.Errorf("creating Timestream Database (%s): empty output", dbName)
	}

	d.SetId(aws.StringValue(resp.Database.DatabaseName))

	return resourceDatabaseRead(ctx, d, meta)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TimestreamWriteConn(ctx)

	input := &timestreamwrite.DescribeDatabaseInput{
		DatabaseName: aws.String(d.Id()),
	}

	resp, err := conn.DescribeDatabaseWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, timestreamwrite.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Timestream Database %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Timestream Database (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.Database == nil {
		return diag.Errorf("reading Timestream Database (%s): empty output", d.Id())
	}

	db := resp.Database
	arn := aws.StringValue(db.Arn)

	d.Set("arn", arn)
	d.Set("database_name", db.DatabaseName)
	d.Set("kms_key_id", db.KmsKeyId)
	d.Set("table_count", db.TableCount)

	return nil
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TimestreamWriteConn(ctx)

	if d.HasChange("kms_key_id") {
		input := &timestreamwrite.UpdateDatabaseInput{
			DatabaseName: aws.String(d.Id()),
			KmsKeyId:     aws.String(d.Get("kms_key_id").(string)),
		}

		_, err := conn.UpdateDatabaseWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Timestream Database (%s): %s", d.Id(), err)
		}
	}

	return resourceDatabaseRead(ctx, d, meta)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TimestreamWriteConn(ctx)

	log.Printf("[INFO] Deleting Timestream Database: %s", d.Id())
	_, err := conn.DeleteDatabaseWithContext(ctx, &timestreamwrite.DeleteDatabaseInput{
		DatabaseName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, timestreamwrite.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Timestream Database (%s): %s", d.Id(), err)
	}

	return nil
}
