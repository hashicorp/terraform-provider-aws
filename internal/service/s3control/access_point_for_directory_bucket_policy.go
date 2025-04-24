// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3control_directory_access_point_policy", name="Directory Access Point Policy")
func resourceAccessPointForDirectoryBucketPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessPointForDirectoryBucketPolicyCreate,
		ReadWithoutTimeout:   resourceAccessPointForDirectoryBucketPolicyRead,
		UpdateWithoutTimeout: resourceAccessPointForDirectoryBucketPolicyUpdate,
		DeleteWithoutTimeout: resourceAccessPointForDirectoryBucketPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAccessPointForDirectoryBucketPolicyImport,
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
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceAccessPointForDirectoryBucketPolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	name := d.Get(names.AttrName).(string)
	accountID := d.Get(names.AttrAccountID).(string)

	resourceID, err := AccessPointForDirectoryBucketCreateResourceID(name, accountID)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Access Point for Directory Bucket (%s:%s) Scope: %s", name, accountID, err)
	}
	d.SetId(resourceID)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3control.PutAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Policy:    aws.String(policy),
	}

	_, err = conn.PutAccessPointPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Access Point for Directory Bucket (%s) Policy: %s", resourceID, err)
	}

	return append(diags, resourceAccessPointForDirectoryBucketPolicyRead(ctx, d, meta)...)
}

func resourceAccessPointForDirectoryBucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	name, accountID, err := AccessPointForDirectoryBucketParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrName, name)
	d.Set(names.AttrAccountID, accountID)

	policy, err := FindAccessPointForDirectoryBucketPolicyByTwoPartKey(ctx, conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Access Point for Directory Bucket Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Access Point for Directory Bucket Policy (%s): %s", d.Id(), err)
	}

	if policy != "" {
		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), policy)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	} else {
		d.Set(names.AttrPolicy, "")
	}

	return diags
}

func resourceAccessPointForDirectoryBucketPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	name, accountID, err := AccessPointForDirectoryBucketParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3control.PutAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Policy:    aws.String(policy),
	}

	_, err = conn.PutAccessPointPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Access Point for Directory Bucket Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAccessPointForDirectoryBucketPolicyRead(ctx, d, meta)...)
}

func resourceAccessPointForDirectoryBucketPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	name, accountID, err := AccessPointForDirectoryBucketParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting S3 Access Point for Directory Bucket Policy: %s", d.Id())
	_, err = conn.DeleteAccessPointPolicy(ctx, &s3control.DeleteAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Access Point for Directory Bucket Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAccessPointForDirectoryBucketPolicyImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
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
