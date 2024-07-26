// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_finspace_kx_user", name="Kx User")
// @Tags(identifierAttribute="arn")
func ResourceKxUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKxUserCreate,
		ReadWithoutTimeout:   resourceKxUserRead,
		UpdateWithoutTimeout: resourceKxUserUpdate,
		DeleteWithoutTimeout: resourceKxUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			names.AttrName: {
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

	kxUserIDPartCount = 2
)

func resourceKxUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	in := &finspace.CreateKxUserInput{
		UserName:      aws.String(d.Get(names.AttrName).(string)),
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		IamRole:       aws.String(d.Get("iam_role").(string)),
		Tags:          getTagsIn(ctx),
	}

	out, err := client.CreateKxUser(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxUser, d.Get(names.AttrName).(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxUser, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	idParts := []string{
		aws.ToString(out.EnvironmentId),
		aws.ToString(out.UserName),
	}
	id, err := flex.FlattenResourceId(idParts, kxUserIDPartCount, false)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionFlatteningResourceId, ResNameKxUser, d.Get(names.AttrName).(string), err)
	}
	d.SetId(id)

	return append(diags, resourceKxUserRead(ctx, d, meta)...)
}

func resourceKxUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	out, err := findKxUserByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FinSpace KxUser (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionReading, ResNameKxUser, d.Id(), err)
	}

	d.Set(names.AttrARN, out.UserArn)
	d.Set(names.AttrName, out.UserName)
	d.Set("iam_role", out.IamRole)
	d.Set("environment_id", out.EnvironmentId)

	return diags
}

func resourceKxUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	if d.HasChange("iam_role") {
		in := &finspace.UpdateKxUserInput{
			EnvironmentId: aws.String(d.Get("environment_id").(string)),
			UserName:      aws.String(d.Get(names.AttrName).(string)),
			IamRole:       aws.String(d.Get("iam_role").(string)),
		}

		_, err := conn.UpdateKxUser(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.FinSpace, create.ErrActionUpdating, ResNameKxUser, d.Id(), err)
		}
	}

	return append(diags, resourceKxUserRead(ctx, d, meta)...)
}

func resourceKxUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	log.Printf("[INFO] Deleting FinSpace KxUser %s", d.Id())

	_, err := conn.DeleteKxUser(ctx, &finspace.DeleteKxUserInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		UserName:      aws.String(d.Get(names.AttrName).(string)),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionDeleting, ResNameKxUser, d.Id(), err)
	}

	return diags
}

func findKxUserByID(ctx context.Context, conn *finspace.Client, id string) (*finspace.GetKxUserOutput, error) {
	parts, err := flex.ExpandResourceId(id, kxUserIDPartCount, false)
	if err != nil {
		return nil, err
	}
	in := &finspace.GetKxUserInput{
		EnvironmentId: aws.String(parts[0]),
		UserName:      aws.String(parts[1]),
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
