// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	BucketAccessKeyIdPartsCount = 2
)

// @SDKResource("aws_lightsail_bucket_access_key")
func ResourceBucketAccessKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketAccessKeyCreate,
		ReadWithoutTimeout:   resourceBucketAccessKeyRead,
		DeleteWithoutTimeout: resourceBucketAccessKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrBucketName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9a-z][0-9a-z-]{1,52}[0-9a-z]$`), "Invalid Bucket name. Must match regex: ^[0-9a-z][0-9a-z-]{1,52}[0-9a-z]$"),
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBucketAccessKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	in := lightsail.CreateBucketAccessKeyInput{
		BucketName: aws.String(d.Get(names.AttrBucketName).(string)),
	}

	out, err := conn.CreateBucketAccessKey(ctx, &in)

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeCreateBucketAccessKey), ResBucketAccessKey, d.Get(names.AttrBucketName).(string), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeCreateBucketAccessKey, ResBucketAccessKey, d.Get(names.AttrBucketName).(string))

	if diag != nil {
		return diag
	}

	idParts := []string{d.Get(names.AttrBucketName).(string), *out.AccessKey.AccessKeyId}
	id, err := flex.FlattenResourceId(idParts, BucketAccessKeyIdPartsCount, false)

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionFlatteningResourceId, ResBucketAccessKey, d.Get(names.AttrBucketName).(string), err)
	}

	d.SetId(id)
	d.Set("secret_access_key", out.AccessKey.SecretAccessKey)

	return append(diags, resourceBucketAccessKeyRead(ctx, d, meta)...)
}

func resourceBucketAccessKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	out, err := FindBucketAccessKeyById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResBucketAccessKey, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionReading, ResBucketAccessKey, d.Id(), err)
	}

	d.Set("access_key_id", out.AccessKeyId)
	d.Set(names.AttrBucketName, d.Get(names.AttrBucketName).(string))
	d.Set(names.AttrCreatedAt, out.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrStatus, out.Status)

	return diags
}

func resourceBucketAccessKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	parts, err := flex.ExpandResourceId(d.Id(), BucketAccessKeyIdPartsCount, false)

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionExpandingResourceId, ResBucketAccessKey, d.Id(), err)
	}

	out, err := conn.DeleteBucketAccessKey(ctx, &lightsail.DeleteBucketAccessKeyInput{
		BucketName:  aws.String(parts[0]),
		AccessKeyId: aws.String(parts[1]),
	})

	if err != nil && errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionDeleting, ResBucketAccessKey, d.Id(), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeDeleteBucketAccessKey, ResBucketAccessKey, d.Id())

	if diag != nil {
		return diag
	}

	return diags
}

func FindBucketAccessKeyById(ctx context.Context, conn *lightsail.Client, id string) (*types.AccessKey, error) {
	parts, err := flex.ExpandResourceId(id, BucketAccessKeyIdPartsCount, false)

	if err != nil {
		return nil, err
	}

	in := &lightsail.GetBucketAccessKeysInput{BucketName: aws.String(parts[0])}
	out, err := conn.GetBucketAccessKeys(ctx, in)

	if IsANotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	var entry types.AccessKey
	entryExists := false

	for _, n := range out.AccessKeys {
		if parts[1] == aws.ToString(n.AccessKeyId) {
			entry = n
			entryExists = true
			break
		}
	}

	if !entryExists {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &entry, nil
}
