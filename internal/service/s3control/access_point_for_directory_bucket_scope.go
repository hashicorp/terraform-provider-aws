// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3control_directory_access_point_scope", name="Directory Access Point Scope")
func resourceAccessPointForDirectoryBucketScope() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessPointForDirectoryBucketScopeCreate,
		ReadWithoutTimeout:   resourceAccessPointForDirectoryBucketScopeRead,
		UpdateWithoutTimeout: resourceAccessPointForDirectoryBucketScopeUpdate,
		DeleteWithoutTimeout: resourceAccessPointForDirectoryBucketScopeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAccessPointScopeImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(AccessPointForDirectoryBucketNameRegex,
					"must be in the format [access_point_name]--[azid]--xa-s3. Use the aws_s3_access_point resource to manage general purpose access points"),
			},
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrScope: {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permissions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"prefixes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceAccessPointForDirectoryBucketScopeCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)
	name := d.Get(names.AttrName).(string)
	accountID := d.Get(names.AttrAccountID).(string)

	resourceID, err := AccessPointForDirectoryBucketCreateResourceID(name, accountID)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Access Point for Directory Bucket (%s:%s) Scope: %s", name, accountID, err)
	}
	d.SetId(resourceID)

	var scope *types.Scope
	if v, ok := d.GetOk(names.AttrScope); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		scope = expandScope(v.([]any)[0].(map[string]any))
	}

	input := &s3control.PutAccessPointScopeInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Scope:     scope,
	}

	_, err = conn.PutAccessPointScope(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Access Point for Directory Bucket (%s:%s) Scope: %s", name, accountID, err)
	}

	return append(diags, resourceAccessPointForDirectoryBucketScopeRead(ctx, d, meta)...)
}

func resourceAccessPointForDirectoryBucketScopeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	name, accountID, err := AccessPointForDirectoryBucketParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrName, name)
	d.Set(names.AttrAccountID, accountID)

	scope, err := FindAccessPointScopeByTwoPartKey(ctx, conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Access Point for Directory Bucket Scope (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Directory S3 Access Point Scope (%s): %s", d.Id(), err)
	}

	if scope != nil {
		flattened := flattenScope(scope)
		if len(flattened) > 0 {
			if err := d.Set(names.AttrScope, []any{flattened}); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else {
			if err := d.Set(names.AttrScope, []any{}); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	} else {
		if err := d.Set(names.AttrScope, []any{}); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return diags
}

func resourceAccessPointForDirectoryBucketScopeUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	name, accountID, err := AccessPointForDirectoryBucketParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	var scope *types.Scope
	if v, ok := d.GetOk(names.AttrScope); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		scope = expandScope(v.([]any)[0].(map[string]any))
	}

	input := &s3control.PutAccessPointScopeInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Scope:     scope,
	}

	_, err = conn.PutAccessPointScope(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Access Point for Directory Bucket Scope (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAccessPointForDirectoryBucketScopeRead(ctx, d, meta)...)
}

func resourceAccessPointForDirectoryBucketScopeDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	name, accountID, err := AccessPointForDirectoryBucketParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting S3 Access Point for Directory Bucket Policy: %s", d.Id())
	_, err = conn.DeleteAccessPointScope(ctx, &s3control.DeleteAccessPointScopeInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Access Point Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAccessPointScopeImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	name, accountID, err := AccessPointForDirectoryBucketParseResourceID(d.Id())

	if err != nil {
		return nil, err
	}

	resourceID, err := AccessPointForDirectoryBucketCreateResourceID(name, accountID)
	if err != nil {
		return nil, err
	}

	d.SetId(resourceID)
	return []*schema.ResourceData{d}, nil
}
