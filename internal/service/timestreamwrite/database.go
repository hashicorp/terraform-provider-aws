// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_timestreamwrite_database", name="Database")
// @Tags(identifierAttribute="arn")
func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDatabaseCreate,
		ReadWithoutTimeout:   resourceDatabaseRead,
		UpdateWithoutTimeout: resourceDatabaseUpdate,
		DeleteWithoutTimeout: resourceDatabaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},
			names.AttrKMSKeyID: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TimestreamWriteClient(ctx)

	name := d.Get(names.AttrDatabaseName).(string)
	input := &timestreamwrite.CreateDatabaseInput{
		DatabaseName: aws.String(name),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	output, err := conn.CreateDatabase(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Timestream Database (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Database.DatabaseName))

	return append(diags, resourceDatabaseRead(ctx, d, meta)...)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TimestreamWriteClient(ctx)

	db, err := findDatabaseByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Timestream Database %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Timestream Database (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, db.Arn)
	d.Set(names.AttrDatabaseName, db.DatabaseName)
	d.Set(names.AttrKMSKeyID, db.KmsKeyId)
	d.Set("table_count", db.TableCount)

	return diags
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TimestreamWriteClient(ctx)

	if d.HasChange(names.AttrKMSKeyID) {
		input := &timestreamwrite.UpdateDatabaseInput{
			DatabaseName: aws.String(d.Id()),
			KmsKeyId:     aws.String(d.Get(names.AttrKMSKeyID).(string)),
		}

		_, err := conn.UpdateDatabase(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Timestream Database (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDatabaseRead(ctx, d, meta)...)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TimestreamWriteClient(ctx)

	log.Printf("[INFO] Deleting Timestream Database: %s", d.Id())
	_, err := conn.DeleteDatabase(ctx, &timestreamwrite.DeleteDatabaseInput{
		DatabaseName: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Timestream Database (%s): %s", d.Id(), err)
	}

	return diags
}

func findDatabaseByName(ctx context.Context, conn *timestreamwrite.Client, name string) (*types.Database, error) {
	input := &timestreamwrite.DescribeDatabaseInput{
		DatabaseName: aws.String(name),
	}

	output, err := conn.DescribeDatabase(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Database == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Database, nil
}
