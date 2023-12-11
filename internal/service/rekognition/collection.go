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

// @SDKResource("aws_rekognition_collection", name="Collection")
func ResourceCollection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: Create,
		ReadWithoutTimeout:   Read,
		DeleteWithoutTimeout: Delete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("collection_id", d.Id())
				return []*schema.ResourceData{d}, nil
			},
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
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameCollection = "Collection"
)

func Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	tags, err := GetResourceTags(ctx, conn, "arn:"+*out.CollectionArn)
	if err != nil {
		return create.AppendDiagError(diags, names.Rekognition, create.ErrActionReading, ResNameCollection, collectionId, err)
	}

	d.Set("tags_all", tags.Tags)

	return append(diags, Read(ctx, d, meta)...)
}

func Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).RekognitionClient(ctx)

	collectionId := d.Id()

	out, err := FindCollectionByID(ctx, conn, collectionId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Rekognition Collection (%s) not found, removing from state", collectionId)
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Rekognition, create.ErrActionReading, ResNameCollection, collectionId, err)
	}

	d.Set("arn", out.CollectionARN)
	d.Set("face_model_version", out.FaceModelVersion)
	d.Set("face_model_version", out.FaceModelVersion)

	tags, err := GetResourceTags(ctx, conn, *out.CollectionARN)
	if err != nil {
		return create.AppendDiagError(diags, names.Rekognition, create.ErrActionReading, ResNameCollection, collectionId, err)
	}

	d.Set("tags_all", tags.Tags)

	return diags
}

func Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	collectionId := d.Id()

	conn := meta.(*conns.AWSClient).RekognitionClient(ctx)

	log.Printf("[INFO] Deleting Rekognition Collection %s", collectionId)

	_, err := conn.DeleteCollection(ctx, &rekognition.DeleteCollectionInput{
		CollectionId: aws.String(collectionId),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.Rekognition, create.ErrActionDeleting, ResNameCollection, collectionId, err)
	}

	return diags
}

// resource tags can only be retrieved with a separate operation
func GetResourceTags(ctx context.Context, conn *rekognition.Client, arn string) (*rekognition.ListTagsForResourceOutput, error) {
	in := &rekognition.ListTagsForResourceInput{
		ResourceArn: aws.String(arn),
	}

	out, err := conn.ListTagsForResource(ctx, in)
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

func FindCollectionByID(ctx context.Context, conn *rekognition.Client, id string) (*rekognition.DescribeCollectionOutput, error) {
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
