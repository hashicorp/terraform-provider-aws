package lightsail

import (
	"context"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
			"bucket_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,52}[a-z0-9]$`), "Invalid Bucket name. Must match regex: ^[a-z0-9][a-z0-9-]{1,52}[a-z0-9]$"),
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBucketAccessKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	in := lightsail.CreateBucketAccessKeyInput{
		BucketName: aws.String(d.Get("bucket_name").(string)),
	}

	out, err := conn.CreateBucketAccessKeyWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateBucketAccessKey, ResBucketAccessKey, d.Get("bucket_name").(string), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeCreateBucketAccessKey, ResBucketAccessKey, d.Get("bucket_name").(string))

	if diag != nil {
		return diag
	}

	idParts := []string{d.Get("bucket_name").(string), *out.AccessKey.AccessKeyId}
	id, err := flex.FlattenResourceId(idParts, BucketAccessKeyIdPartsCount)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionFlatteningResourceId, ResBucketAccessKey, d.Get("bucket_name").(string), err)
	}

	d.SetId(id)
	d.Set("secret_access_key", out.AccessKey.SecretAccessKey)

	return resourceBucketAccessKeyRead(ctx, d, meta)
}

func resourceBucketAccessKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	out, err := FindBucketAccessKeyById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResBucketAccessKey, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResBucketAccessKey, d.Id(), err)
	}

	d.Set("access_key_id", out.AccessKeyId)
	d.Set("bucket_name", d.Get("bucket_name").(string))
	d.Set("created_at", out.CreatedAt.Format(time.RFC3339))
	d.Set("status", out.Status)

	return nil
}

func resourceBucketAccessKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	parts, err := flex.ExpandResourceId(d.Id(), BucketAccessKeyIdPartsCount)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionExpandingResourceId, ResBucketAccessKey, d.Id(), err)
	}

	out, err := conn.DeleteBucketAccessKeyWithContext(ctx, &lightsail.DeleteBucketAccessKeyInput{
		BucketName:  aws.String(parts[0]),
		AccessKeyId: aws.String(parts[1]),
	})

	if err != nil && tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionDeleting, ResBucketAccessKey, d.Id(), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeDeleteBucketAccessKey, ResBucketAccessKey, d.Id())

	if diag != nil {
		return diag
	}

	return nil
}
