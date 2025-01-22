// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediastore

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mediastore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_media_store_container", name="Container")
// @Tags(identifierAttribute="arn")
func ResourceContainer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContainerCreate,
		ReadWithoutTimeout:   resourceContainerRead,
		UpdateWithoutTimeout: resourceContainerUpdate,
		DeleteWithoutTimeout: resourceContainerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^\w+$`), "must contain alphanumeric characters or underscores"),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceContainerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaStoreClient(ctx)

	input := &mediastore.CreateContainerInput{
		ContainerName: aws.String(d.Get(names.AttrName).(string)),
		Tags:          getTagsIn(ctx),
	}

	resp, err := conn.CreateContainer(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MediaStore Container: %s", err)
	}

	d.SetId(aws.ToString(resp.Container.Name))

	_, err = waitContainerActive(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MediaStore Container (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceContainerRead(ctx, d, meta)...)
}

func resourceContainerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaStoreClient(ctx)

	resp, err := findContainerByName(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		log.Printf("[WARN] No Container found: %s, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MediaStore Container %s: %s", d.Id(), err)
	}

	arn := aws.ToString(resp.ARN)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrName, resp.Name)
	d.Set(names.AttrEndpoint, resp.Endpoint)

	return diags
}

func resourceContainerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceContainerRead(ctx, d, meta)...)
}

func resourceContainerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaStoreClient(ctx)

	input := &mediastore.DeleteContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	_, err := conn.DeleteContainer(ctx, input)

	if errs.IsA[*awstypes.ContainerNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MediaStore Container (%s): %s", d.Id(), err)
	}

	_, err = waitContainerDeleted(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MediaStore Container (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func containerRefreshStatusFunc(ctx context.Context, conn *mediastore.Client, cn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := findContainerByName(ctx, conn, cn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return resp, string(resp.Status), nil
	}
}

func findContainerByName(ctx context.Context, conn *mediastore.Client, id string) (*awstypes.Container, error) {
	input := &mediastore.DescribeContainerInput{
		ContainerName: aws.String(id),
	}

	output, err := conn.DescribeContainer(ctx, input)

	if errs.IsA[*awstypes.ContainerNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Container == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Container, nil
}

func waitContainerActive(ctx context.Context, conn *mediastore.Client, id string) (*awstypes.Container, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ContainerStatusCreating),
		Target:     enum.Slice(awstypes.ContainerStatusActive),
		Refresh:    containerRefreshStatusFunc(ctx, conn, id),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*awstypes.Container); ok {
		return v, err
	}

	return nil, err
}

func waitContainerDeleted(ctx context.Context, conn *mediastore.Client, id string) (*awstypes.Container, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ContainerStatusDeleting),
		Target:     []string{},
		Refresh:    containerRefreshStatusFunc(ctx, conn, id),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*awstypes.Container); ok {
		return v, err
	}

	return nil, err
}
