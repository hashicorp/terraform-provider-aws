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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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

// @SDKResource("aws_finspace_kx_database", name="Kx Database")
// @Tags(identifierAttribute="arn")
func ResourceKxDatabase() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKxDatabaseCreate,
		ReadWithoutTimeout:   resourceKxDatabaseRead,
		UpdateWithoutTimeout: resourceKxDatabaseUpdate,
		DeleteWithoutTimeout: resourceKxDatabaseDelete,

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
			"created_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"last_modified_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameKxDatabase = "Kx Database"

	kxDatabaseIDPartCount = 2
)

func resourceKxDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	in := &finspace.CreateKxDatabaseInput{
		DatabaseName:  aws.String(d.Get(names.AttrName).(string)),
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		ClientToken:   aws.String(id.UniqueId()),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		in.Description = aws.String(v.(string))
	}

	out, err := conn.CreateKxDatabase(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxDatabase, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.DatabaseArn == nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxDatabase, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	idParts := []string{
		aws.ToString(out.EnvironmentId),
		aws.ToString(out.DatabaseName),
	}
	id, err := flex.FlattenResourceId(idParts, kxDatabaseIDPartCount, false)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionFlatteningResourceId, ResNameKxDatabase, d.Get(names.AttrName).(string), err)
	}

	d.SetId(id)

	return append(diags, resourceKxDatabaseRead(ctx, d, meta)...)
}

func resourceKxDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	out, err := findKxDatabaseByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FinSpace KxDatabase (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionReading, ResNameKxDatabase, d.Id(), err)
	}

	d.Set(names.AttrARN, out.DatabaseArn)
	d.Set(names.AttrName, out.DatabaseName)
	d.Set("environment_id", out.EnvironmentId)
	d.Set(names.AttrDescription, out.Description)
	d.Set("created_timestamp", out.CreatedTimestamp.String())
	d.Set("last_modified_timestamp", out.LastModifiedTimestamp.String())

	return diags
}

func resourceKxDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	if d.HasChanges(names.AttrDescription) {
		in := &finspace.UpdateKxDatabaseInput{
			EnvironmentId: aws.String(d.Get("environment_id").(string)),
			DatabaseName:  aws.String(d.Get(names.AttrName).(string)),
			Description:   aws.String(d.Get(names.AttrDescription).(string)),
		}

		_, err := conn.UpdateKxDatabase(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.FinSpace, create.ErrActionUpdating, ResNameKxDatabase, d.Id(), err)
		}
	}

	return append(diags, resourceKxDatabaseRead(ctx, d, meta)...)
}

func resourceKxDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	log.Printf("[INFO] Deleting FinSpace KxDatabase %s", d.Id())

	_, err := conn.DeleteKxDatabase(ctx, &finspace.DeleteKxDatabaseInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		DatabaseName:  aws.String(d.Get(names.AttrName).(string)),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionDeleting, ResNameKxDatabase, d.Id(), err)
	}

	return diags
}

func findKxDatabaseByID(ctx context.Context, conn *finspace.Client, id string) (*finspace.GetKxDatabaseOutput, error) {
	parts, err := flex.ExpandResourceId(id, kxDatabaseIDPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &finspace.GetKxDatabaseInput{
		EnvironmentId: aws.String(parts[0]),
		DatabaseName:  aws.String(parts[1]),
	}

	out, err := conn.GetKxDatabase(ctx, in)
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

	if out == nil || out.DatabaseArn == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
