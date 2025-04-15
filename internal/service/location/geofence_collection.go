// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	awstypes "github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_location_geofence_collection", name="Geofence Collection")
// @Tags(identifierAttribute="collection_arn")
func ResourceGeofenceCollection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGeofenceCollectionCreate,
		ReadWithoutTimeout:   resourceGeofenceCollectionRead,
		UpdateWithoutTimeout: resourceGeofenceCollectionUpdate,
		DeleteWithoutTimeout: resourceGeofenceCollectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"collection_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"collection_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	ResNameGeofenceCollection = "Geofence Collection"
)

func resourceGeofenceCollectionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	in := &location.CreateGeofenceCollectionInput{
		CollectionName: aws.String(d.Get("collection_name").(string)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok && v != "" {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok && v != "" {
		in.KmsKeyId = aws.String(v.(string))
	}

	out, err := conn.CreateGeofenceCollection(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionCreating, ResNameGeofenceCollection, d.Get("collection_name").(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionCreating, ResNameGeofenceCollection, d.Get("collection_name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.CollectionName))

	return append(diags, resourceGeofenceCollectionRead(ctx, d, meta)...)
}

func resourceGeofenceCollectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	out, err := findGeofenceCollectionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Location GeofenceCollection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionReading, ResNameGeofenceCollection, d.Id(), err)
	}

	d.Set("collection_arn", out.CollectionArn)
	d.Set("collection_name", out.CollectionName)
	d.Set(names.AttrCreateTime, aws.ToTime(out.CreateTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, out.Description)
	d.Set(names.AttrKMSKeyID, out.KmsKeyId)
	d.Set("update_time", aws.ToTime(out.UpdateTime).Format(time.RFC3339))

	return diags
}

func resourceGeofenceCollectionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	update := false

	in := &location.UpdateGeofenceCollectionInput{
		CollectionName: aws.String(d.Id()),
	}

	if d.HasChange(names.AttrDescription) {
		in.Description = aws.String(d.Get(names.AttrDescription).(string))
		update = true
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating Location GeofenceCollection (%s): %#v", d.Id(), in)
	_, err := conn.UpdateGeofenceCollection(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionUpdating, ResNameGeofenceCollection, d.Id(), err)
	}

	return append(diags, resourceGeofenceCollectionRead(ctx, d, meta)...)
}

func resourceGeofenceCollectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	log.Printf("[INFO] Deleting Location GeofenceCollection %s", d.Id())

	_, err := conn.DeleteGeofenceCollection(ctx, &location.DeleteGeofenceCollectionInput{
		CollectionName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionDeleting, ResNameGeofenceCollection, d.Id(), err)
	}

	return diags
}

func findGeofenceCollectionByName(ctx context.Context, conn *location.Client, name string) (*location.DescribeGeofenceCollectionOutput, error) {
	in := &location.DescribeGeofenceCollectionInput{
		CollectionName: aws.String(name),
	}

	out, err := conn.DescribeGeofenceCollection(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
