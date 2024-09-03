// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	resourceSharePropagationTimeout = 1 * time.Minute
)

// @SDKResource("aws_ram_resource_share", name="Resource Share")
// @Tags(identifierAttribute="id")
func resourceResourceShare() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceShareCreate,
		ReadWithoutTimeout:   resourceResourceShareRead,
		UpdateWithoutTimeout: resourceResourceShareUpdate,
		DeleteWithoutTimeout: resourceResourceShareDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"allow_external_principals": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"permission_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceResourceShareCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ram.CreateResourceShareInput{
		AllowExternalPrincipals: aws.Bool(d.Get("allow_external_principals").(bool)),
		Name:                    aws.String(name),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("permission_arns"); ok && v.(*schema.Set).Len() > 0 {
		input.PermissionArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateResourceShare(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RAM Resource Share (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ResourceShare.ResourceShareArn))

	_, err = tfresource.RetryWhenNotFound(ctx, resourceSharePropagationTimeout, func() (interface{}, error) {
		return findResourceShareOwnerSelfByARN(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM Resource Share (%s) create: %s", d.Id(), err)
	}

	if _, err := waitResourceShareOwnedBySelfActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM Resource Share (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceResourceShareRead(ctx, d, meta)...)
}

func resourceResourceShareRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	resourceShare, err := findResourceShareOwnerSelfByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RAM Resource Share (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Share (%s): %s", d.Id(), err)
	}

	d.Set("allow_external_principals", resourceShare.AllowExternalPrincipals)
	d.Set(names.AttrARN, resourceShare.ResourceShareArn)
	d.Set(names.AttrName, resourceShare.Name)

	setTagsOut(ctx, resourceShare.Tags)

	input := &ram.ListResourceSharePermissionsInput{
		ResourceShareArn: aws.String(d.Id()),
	}
	var permissions []awstypes.ResourceSharePermissionSummary

	pages := ram.NewListResourceSharePermissionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading RAM Resource Share (%s) permissions: %s", d.Id(), err)
		}

		permissions = append(permissions, page.Permissions...)
	}

	permissionARNs := tfslices.ApplyToAll(permissions, func(r awstypes.ResourceSharePermissionSummary) string {
		return aws.ToString(r.Arn)
	})
	d.Set("permission_arns", permissionARNs)

	return diags
}

func resourceResourceShareUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	if d.HasChanges("allow_external_principals", names.AttrName) {
		input := &ram.UpdateResourceShareInput{
			AllowExternalPrincipals: aws.Bool(d.Get("allow_external_principals").(bool)),
			Name:                    aws.String(d.Get(names.AttrName).(string)),
			ResourceShareArn:        aws.String(d.Id()),
		}

		_, err := conn.UpdateResourceShare(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RAM Resource Share (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceResourceShareRead(ctx, d, meta)...)
}

func resourceResourceShareDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	log.Printf("[DEBUG] Deleting RAM Resource Share: %s", d.Id())
	_, err := conn.DeleteResourceShare(ctx, &ram.DeleteResourceShareInput{
		ResourceShareArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.UnknownResourceException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RAM Resource Share (%s): %s", d.Id(), err)
	}

	if _, err := waitResourceShareOwnedBySelfDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM Resource Share (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findResourceShareOwnerSelfByARN(ctx context.Context, conn *ram.Client, arn string) (*awstypes.ResourceShare, error) {
	input := &ram.GetResourceSharesInput{
		ResourceOwner:     awstypes.ResourceOwnerSelf,
		ResourceShareArns: []string{arn},
	}
	output, err := findResourceShare(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.ResourceShareStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findResourceShare(ctx context.Context, conn *ram.Client, input *ram.GetResourceSharesInput) (*awstypes.ResourceShare, error) {
	output, err := findResourceShares(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findResourceShares(ctx context.Context, conn *ram.Client, input *ram.GetResourceSharesInput) ([]awstypes.ResourceShare, error) {
	var output []awstypes.ResourceShare

	pages := ram.NewGetResourceSharesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceArnNotFoundException](err) || errs.IsA[*awstypes.UnknownResourceException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ResourceShares...)
	}

	return output, nil
}

func statusResourceShareOwnerSelf(ctx context.Context, conn *ram.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findResourceShareOwnerSelfByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitResourceShareOwnedBySelfActive(ctx context.Context, conn *ram.Client, arn string, timeout time.Duration) (*awstypes.ResourceShare, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceShareStatusPending),
		Target:  enum.Slice(awstypes.ResourceShareStatusActive),
		Refresh: statusResourceShareOwnerSelf(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceShare); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitResourceShareOwnedBySelfDeleted(ctx context.Context, conn *ram.Client, arn string, timeout time.Duration) (*awstypes.ResourceShare, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceShareStatusDeleting),
		Target:  []string{},
		Refresh: statusResourceShareOwnerSelf(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceShare); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
