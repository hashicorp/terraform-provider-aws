package finspace

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_finspace_kx_user")
// @Tags(identifierAttribute="arn")
func ResourceKxUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKxUserCreate,
		ReadWithoutTimeout:   resourceKxUserRead,
		UpdateWithoutTimeout: resourceKxUserUpdate,
		DeleteWithoutTimeout: resourceKxUserDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"iam_role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameKxUser = "Kx User"
)

func resourceKxUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).FinSpaceClient()

	in := &finspace.CreateKxUserInput{
		UserName:      aws.String(d.Get("name").(string)),
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		IamRole:       aws.String(d.Get("iam_role").(string)),
		Tags:          GetTagsIn(ctx),
	}

	out, err := client.CreateKxUser(ctx, in)
	if err != nil {
		return create.DiagError(names.FinSpace, create.ErrActionCreating, ResNameKxUser, d.Get("name").(string), err)
	}

	if out == nil {
		return create.DiagError(names.FinSpace, create.ErrActionCreating, ResNameKxUser, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.EnvironmentId) + "," + aws.ToString(out.UserName))

	return resourceKxUserRead(ctx, d, meta)
}

func resourceKxUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FinSpaceClient()

	out, err := findKxUserByName(ctx, conn, d.Get("name").(string), d.Get("environment_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FinSpace KxUser (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.FinSpace, create.ErrActionReading, ResNameKxUser, d.Id(), err)
	}

	d.Set("arn", out.UserArn)
	d.Set("name", out.UserName)
	d.Set("iam_role", out.IamRole)
	d.Set("environment_id", out.EnvironmentId)

	return nil
}

func resourceKxUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FinSpaceClient()

	update := false

	in := &finspace.UpdateKxUserInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		UserName:      aws.String(d.Get("name").(string)),
	}

	if d.HasChanges("iam_role") {
		in.IamRole = aws.String(d.Get("iam_role").(string))
	}

	if d.HasChanges("iam_role") {
		update = true
		log.Printf("[DEBUG] Updating FinSpace KxUser (%s): %#v", d.Id(), in)
		_, err := conn.UpdateKxUser(ctx, in)
		if err != nil {
			return create.DiagError(names.FinSpace, create.ErrActionUpdating, ResNameKxUser, d.Id(), err)
		}
	}

	if !update {
		return nil
	}
	return resourceKxUserRead(ctx, d, meta)
}

func resourceKxUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FinSpaceClient()

	log.Printf("[INFO] Deleting FinSpace KxUser %s", d.Id())

	_, err := conn.DeleteKxUser(ctx, &finspace.DeleteKxUserInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		UserName:      aws.String(d.Get("name").(string)),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.FinSpace, create.ErrActionDeleting, ResNameKxUser, d.Id(), err)
	}

	return nil
}

func findKxUserByName(ctx context.Context, conn *finspace.Client, name string, environmentId string) (*finspace.GetKxUserOutput, error) {
	in := &finspace.GetKxUserInput{
		UserName:      aws.String(name),
		EnvironmentId: aws.String(environmentId),
	}
	out, err := conn.GetKxUser(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.UserArn == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
