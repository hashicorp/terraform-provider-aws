// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediastore

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\w+$`), "must contain alphanumeric characters or underscores"),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
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
	conn := meta.(*conns.AWSClient).MediaStoreConn(ctx)

	input := &mediastore.CreateContainerInput{
		ContainerName: aws.String(d.Get("name").(string)),
		Tags:          getTagsIn(ctx),
	}

	resp, err := conn.CreateContainerWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MediaStore Container: %s", err)
	}

	d.SetId(aws.StringValue(resp.Container.Name))

	stateConf := &retry.StateChangeConf{
		Pending:    []string{mediastore.ContainerStatusCreating},
		Target:     []string{mediastore.ContainerStatusActive},
		Refresh:    containerRefreshStatusFunc(ctx, conn, d.Id()),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MediaStore Container (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceContainerRead(ctx, d, meta)...)
}

func resourceContainerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaStoreConn(ctx)

	input := &mediastore.DescribeContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	resp, err := conn.DescribeContainerWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, mediastore.ErrCodeContainerNotFoundException) {
		log.Printf("[WARN] No Container found: %s, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MediaStore Container %s: %s", d.Id(), err)
	}

	arn := aws.StringValue(resp.Container.ARN)
	d.Set("arn", arn)
	d.Set("name", resp.Container.Name)
	d.Set("endpoint", resp.Container.Endpoint)

	return diags
}

func resourceContainerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceContainerRead(ctx, d, meta)...)
}

func resourceContainerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaStoreConn(ctx)

	input := &mediastore.DeleteContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	_, err := conn.DeleteContainerWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, mediastore.ErrCodeContainerNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting MediaStore Container (%s): %s", d.Id(), err)
	}

	dcinput := &mediastore.DescribeContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		_, err := conn.DescribeContainerWithContext(ctx, dcinput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, mediastore.ErrCodeContainerNotFoundException) {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("Media Store Container (%s) still exists", d.Id()))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DescribeContainerWithContext(ctx, dcinput)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MediaStore Container (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func containerRefreshStatusFunc(ctx context.Context, conn *mediastore.MediaStore, cn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &mediastore.DescribeContainerInput{
			ContainerName: aws.String(cn),
		}
		resp, err := conn.DescribeContainerWithContext(ctx, input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, *resp.Container.Status, nil
	}
}
