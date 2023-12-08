// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceCollection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCollectionCreate,
		ReadWithoutTimeout:   resourceCollectionRead,
		DeleteWithoutTimeout: resourceCollectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
			"collection_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9_.\-]+$`), "must conform to: ^[a-zA-Z0-9_.\\-]+$"),
			},
			"face_model_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(), // TIP: Many, but not all, resources have `tags` and `tags_all` attributes.
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameCollection = "Collection"
)

func resourceCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).RekognitionClient(ctx)

	collectionId := d.Get("collection_id").(string)

	in := &rekognition.CreateCollectionInput{
		CollectionId: aws.String(collectionId),
		Tags:         getTagsIn(ctx),
	}

	out, err := conn.CreateCollection(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.Rekognition, create.ErrActionCreating, ResNameCollection, collectionId, err)
	}

	d.SetId(collectionId)
	d.Set("arn", out.CollectionArn)
	d.Set("face_model_version", out.FaceModelVersion)

	return append(diags, resourceCollectionRead(ctx, d, meta)...)
}

func resourceCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).RekognitionClient(ctx)

	out, err := findCollectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Rekognition Collection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Rekognition, create.ErrActionReading, ResNameCollection, d.Id(), err)
	}

	d.Set("arn", out.CollectionARN)
	d.Set("face_model_version", out.FaceModelVersion)

	return diags
}

func resourceCollectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).RekognitionClient(ctx)

	log.Printf("[INFO] Deleting Rekognition Collection %s", d.Id())

	_, err := conn.DeleteCollection(ctx, &rekognition.DeleteCollectionInput{
		CollectionId: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.Rekognition, create.ErrActionDeleting, ResNameCollection, d.Id(), err)
	}

	return diags
}

func findCollectionByID(ctx context.Context, conn *rekognition.Client, id string) (*rekognition.DescribeCollectionOutput, error) {
	in := &rekognition.DescribeCollectionInput{
		CollectionId: aws.String(id),
	}

	out, err := conn.DescribeCollection(ctx, in)
	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
