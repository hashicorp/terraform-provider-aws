// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcegroups

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroups"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroups/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_resourcegroups_resource", name="Resource")
func resourceResource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceCreate,
		ReadWithoutTimeout:   resourceResourceRead,
		DeleteWithoutTimeout: resourceResourceDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"group_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrResourceType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ResourceGroupsClient(ctx)

	groupARN := d.Get("group_arn").(string)
	resourceARN := d.Get(names.AttrResourceARN).(string)
	id := strings.Join([]string{strings.Split(strings.ToLower(groupARN), "/")[1], strings.Split(resourceARN, "/")[1]}, "_")
	input := &resourcegroups.GroupResourcesInput{
		Group:        aws.String(groupARN),
		ResourceArns: []string{resourceARN},
	}

	output, err := conn.GroupResources(ctx, input)

	if err == nil {
		err = failedResourcesError(output.Failed)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Resource Groups Resource (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitResourceCreated(ctx, conn, groupARN, resourceARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Resource Groups Resource (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ResourceGroupsClient(ctx)

	output, err := findResourceByTwoPartKey(ctx, conn, d.Get("group_arn").(string), d.Get(names.AttrResourceARN).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ResourceGroups Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Resource Groups Resource (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrResourceARN, output.Identifier.ResourceArn)
	d.Set(names.AttrResourceType, output.Identifier.ResourceType)

	return diags
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ResourceGroupsClient(ctx)

	groupARN := d.Get("group_arn").(string)
	resourceARN := d.Get(names.AttrResourceARN).(string)

	log.Printf("[INFO] Deleting Resource Groups Resource: %s", d.Id())
	output, err := conn.UngroupResources(ctx, &resourcegroups.UngroupResourcesInput{
		Group:        aws.String(groupARN),
		ResourceArns: []string{resourceARN},
	})

	if err == nil {
		err = failedResourcesError(output.Failed)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Resource Groups Resource (%s): %s", d.Id(), err)
	}

	if _, err := waitResourceDeleted(ctx, conn, groupARN, resourceARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Resource Groups Resource (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findResourceByTwoPartKey(ctx context.Context, conn *resourcegroups.Client, groupARN, resourceARN string) (*types.ListGroupResourcesItem, error) {
	input := &resourcegroups.ListGroupResourcesInput{
		Group: aws.String(groupARN),
	}
	var output []types.ListGroupResourcesItem

	pages := resourcegroups.NewListGroupResourcesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Resources...)
	}

	output = slices.Filter(output, func(v types.ListGroupResourcesItem) bool {
		return v.Identifier != nil && aws.ToString(v.Identifier.ResourceArn) == resourceARN
	})

	return tfresource.AssertSingleValueResult(output)
}

func statusResource(ctx context.Context, conn *resourcegroups.Client, groupARN, resourceARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findResourceByTwoPartKey(ctx, conn, groupARN, resourceARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.Status == nil {
			return output, "", nil
		}

		return output, string(output.Status.Name), nil
	}
}

func waitResourceCreated(ctx context.Context, conn *resourcegroups.Client, groupARN, resourceARN string, timeout time.Duration) (*types.ListGroupResourcesItem, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ResourceStatusValuePending),
		Target:  []string{""},
		Refresh: statusResource(ctx, conn, groupARN, resourceARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ListGroupResourcesItem); ok {
		return output, err
	}

	return nil, err
}

func waitResourceDeleted(ctx context.Context, conn *resourcegroups.Client, groupARN, resourceARN string, timeout time.Duration) (*types.ListGroupResourcesItem, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ResourceStatusValuePending),
		Target:  []string{},
		Refresh: statusResource(ctx, conn, groupARN, resourceARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ListGroupResourcesItem); ok {
		return output, err
	}

	return nil, err
}

func failedResourceError(apiObject types.FailedResource) error {
	return fmt.Errorf("%s: %s", aws.ToString(apiObject.ErrorCode), aws.ToString(apiObject.ErrorMessage))
}

func failedResourcesError(apiObjects []types.FailedResource) error {
	var errs []error

	for _, apiObject := range apiObjects {
		err := failedResourceError(apiObject)

		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws.ToString(apiObject.ResourceArn), err))
		}
	}

	return errors.Join(errs...)
}
