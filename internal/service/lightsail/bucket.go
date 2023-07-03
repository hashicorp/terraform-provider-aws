// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_bucket", name="Bucket")
// @Tags(identifierAttribute="id")
func ResourceBucket() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketCreate,
		ReadWithoutTimeout:   resourceBucketRead,
		UpdateWithoutTimeout: resourceBucketUpdate,
		DeleteWithoutTimeout: resourceBucketDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	in := lightsail.CreateBucketInput{
		BucketName: aws.String(d.Get("name").(string)),
		BundleId:   aws.String(d.Get("bundle_id").(string)),
		Tags:       getTagsIn(ctx),
	}

	out, err := conn.CreateBucket(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, string(types.OperationTypeCreateBucket), ResBucket, d.Get("name").(string), err)
	}

	id := d.Get("name").(string)
	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeCreateBucket, ResBucket, id)

	if diag != nil {
		return diag
	}

	d.SetId(id)

	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	out, err := FindBucketById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResBucket, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResBucket, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("availability_zone", out.Location.AvailabilityZone)
	d.Set("bundle_id", out.BundleId)
	d.Set("created_at", out.CreatedAt.Format(time.RFC3339))
	d.Set("name", out.Name)
	d.Set("region", out.Location.RegionName)
	d.Set("support_code", out.SupportCode)
	d.Set("url", out.Url)

	setTagsOut(ctx, out.Tags)

	return nil
}

func resourceBucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	if d.HasChange("bundle_id") {
		in := lightsail.UpdateBucketBundleInput{
			BucketName: aws.String(d.Id()),
			BundleId:   aws.String(d.Get("bundle_id").(string)),
		}
		out, err := conn.UpdateBucketBundle(ctx, &in)

		if err != nil {
			return create.DiagError(names.Lightsail, string(types.OperationTypeUpdateBucket), ResBucket, d.Get("name").(string), err)
		}

		diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeUpdateBucket, ResBucket, d.Get("name").(string))

		if diag != nil {
			return diag
		}
	}

	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	out, err := conn.DeleteBucket(ctx, &lightsail.DeleteBucketInput{
		BucketName: aws.String(d.Id()),
	})

	if err != nil && errs.IsA[*types.NotFoundException](err) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionDeleting, ResBucket, d.Id(), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeDeleteBucket, ResBucket, d.Id())

	if diag != nil {
		return diag
	}

	return nil
}

func FindBucketById(ctx context.Context, conn *lightsail.Client, id string) (*types.Bucket, error) {
	in := &lightsail.GetBucketsInput{BucketName: aws.String(id)}
	out, err := conn.GetBuckets(ctx, in)

	if IsANotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.Buckets) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.Buckets[0], nil
}
