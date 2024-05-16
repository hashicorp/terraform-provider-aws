// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ram.CreateResourceShareInput{
		AllowExternalPrincipals: aws.Bool(d.Get("allow_external_principals").(bool)),
		Name:                    aws.String(name),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("permission_arns"); ok && v.(*schema.Set).Len() > 0 {
		input.PermissionArns = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateResourceShareWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RAM Resource Share (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ResourceShare.ResourceShareArn))

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
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

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
	var permissions []*ram.ResourceSharePermissionSummary

	err = conn.ListResourceSharePermissionsPagesWithContext(ctx, input, func(page *ram.ListResourceSharePermissionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Permissions {
			if v != nil {
				permissions = append(permissions, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Share (%s) permissions: %s", d.Id(), err)
	}

	permissionARNs := tfslices.ApplyToAll(permissions, func(r *ram.ResourceSharePermissionSummary) string {
		return aws.StringValue(r.Arn)
	})
	d.Set("permission_arns", permissionARNs)

	return diags
}

func resourceResourceShareUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	if d.HasChanges("allow_external_principals", names.AttrName) {
		input := &ram.UpdateResourceShareInput{
			AllowExternalPrincipals: aws.Bool(d.Get("allow_external_principals").(bool)),
			Name:                    aws.String(d.Get(names.AttrName).(string)),
			ResourceShareArn:        aws.String(d.Id()),
		}

		_, err := conn.UpdateResourceShareWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RAM Resource Share (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceResourceShareRead(ctx, d, meta)...)
}

func resourceResourceShareDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	log.Printf("[DEBUG] Deleting RAM Resource Share: %s", d.Id())
	_, err := conn.DeleteResourceShareWithContext(ctx, &ram.DeleteResourceShareInput{
		ResourceShareArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
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

func findResourceShareOwnerSelfByARN(ctx context.Context, conn *ram.RAM, arn string) (*ram.ResourceShare, error) {
	input := &ram.GetResourceSharesInput{
		ResourceOwner:     aws.String(ram.ResourceOwnerSelf),
		ResourceShareArns: aws.StringSlice([]string{arn}),
	}
	output, err := findResourceShare(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status); status == ram.ResourceShareStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findResourceShare(ctx context.Context, conn *ram.RAM, input *ram.GetResourceSharesInput) (*ram.ResourceShare, error) {
	output, err := findResourceShares(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findResourceShares(ctx context.Context, conn *ram.RAM, input *ram.GetResourceSharesInput) ([]*ram.ResourceShare, error) {
	var output []*ram.ResourceShare

	err := conn.GetResourceSharesPagesWithContext(ctx, input, func(page *ram.GetResourceSharesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResourceShares {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeResourceArnNotFoundException, ram.ErrCodeUnknownResourceException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusResourceShareOwnerSelf(ctx context.Context, conn *ram.RAM, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findResourceShareOwnerSelfByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitResourceShareOwnedBySelfActive(ctx context.Context, conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShare, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ram.ResourceShareStatusPending},
		Target:  []string{ram.ResourceShareStatusActive},
		Refresh: statusResourceShareOwnerSelf(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ram.ResourceShare); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitResourceShareOwnedBySelfDeleted(ctx context.Context, conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShare, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ram.ResourceShareStatusDeleting},
		Target:  []string{},
		Refresh: statusResourceShareOwnerSelf(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ram.ResourceShare); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
