// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	// e.g. example--usw2-az2--xa-s3
	AccessPointForDirectoryBucketNameRegex = regexache.MustCompile(`^(?:[0-9a-z.-]+)--(?:[0-9a-za-z]+(?:-[0-9a-za-z]+)+)--xa-s3$`)

	// e.g. example--usw2-az2--x-s3
	DirectoryBucketNameRegex = regexache.MustCompile(`^(?:[0-9a-z.-]+)--(?:[0-9a-za-z]+(?:-[0-9a-za-z]+)+)--x-s3$`)

	// e.g. arn:aws:s3express:us-west-2:1234567890:accesspoint/ap-name--usw2-az2--xa-s3
	AccessPointArnRegex = regexache.MustCompile(`^arn:aws(-[a-z]+)?:s3express:[a-z0-9-]+:\d{12}:accesspoint/[a-zA-Z0-9\-]+--[a-z0-9-]+--xa-s3$`)
)

// @SdkResource("aws_s3_directory_access_point", name="Directory Access Point")
func resourceAccessPointForDirectoryBucket() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessPointForDirectoryBucketCreate,
		ReadWithoutTimeout:   resourceAccessPointForDirectoryBucketRead,
		UpdateWithoutTimeout: resourceAccessPointForDirectoryBucketUpdate,
		DeleteWithoutTimeout: resourceAccessPointForDirectoryBucketDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			names.AttrAlias: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(DirectoryBucketNameRegex,
					"must be in the format [bucket_name]--[azid]--x-s3. Use the aws_s3_bucket resource to manage general purpose buckets"),
			},
			"bucket_account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpoints: {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"has_public_access_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"network_origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"public_access_block_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         true,
				MinItems:         0,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"block_public_acls": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"block_public_policy": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"ignore_public_acls": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"restrict_public_buckets": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrVPCConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrScope: {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				ForceNew: true,
				Computed: true,
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

func resourceAccessPointForDirectoryBucketCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
	}
	name := d.Get(names.AttrName).(string)
	input := &s3control.CreateAccessPointInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(d.Get(names.AttrBucket).(string)),
		Name:      aws.String(name),
	}

	if v, ok := d.GetOk("bucket_account_id"); ok {
		input.BucketAccountId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("public_access_block_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.PublicAccessBlockConfiguration = expandPublicAccessBlockConfiguration(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrVPCConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.VpcConfiguration = expandVPCConfiguration(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrScope); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Scope = expandScope(v.([]any)[0].(map[string]any))
	}

	_, err := conn.CreateAccessPoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Access Point for Directory Bucket (%s): %s", name, err)
	}

	resourceID, err := AccessPointForDirectoryBucketCreateResourceID(name, accountID)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.SetId(resourceID)

	if v, ok := d.GetOk(names.AttrPolicy); ok && v.(string) != "" && v.(string) != "{}" {
		policy, err := structure.NormalizeJsonString(v.(string))
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
			return sdkdiag.AppendErrorf(diags, "creating S3 Access Point for Directory Bucket (%s) policy: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk(names.AttrScope); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		scope := expandScope(v.([]any)[0].(map[string]any))

		input := &s3control.PutAccessPointScopeInput{
			AccountId: aws.String(accountID),
			Name:      aws.String(name),
			Scope:     scope,
		}

		_, err = conn.PutAccessPointScope(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating S3 Access Point for Directory Bucket (%s) policy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceAccessPointForDirectoryBucketRead(ctx, d, meta)...)
}

func resourceAccessPointForDirectoryBucketRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	name, accountID, err := AccessPointForDirectoryBucketParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findAccessPointByTwoPartKey(ctx, conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Access Point for Directory Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Access Point for Directory Bucket (%s): %s", d.Id(), err)
	}

	accessPointARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "s3express",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: accountID,
		Resource:  fmt.Sprintf("accesspoint/%s", aws.ToString(output.Name)),
	}

	d.Set(names.AttrARN, accessPointARN.String())
	d.Set(names.AttrBucket, output.Bucket)
	d.Set(names.AttrAccountID, accountID)
	d.Set(names.AttrAlias, output.Alias)
	d.Set("bucket_account_id", output.BucketAccountId)
	d.Set(names.AttrDomainName, meta.(*conns.AWSClient).RegionalHostname(ctx, fmt.Sprintf("%s-%s.s3-accesspoint", aws.ToString(output.Name), accountID)))
	d.Set(names.AttrEndpoints, output.Endpoints)
	d.Set(names.AttrName, output.Name)
	d.Set("network_origin", output.NetworkOrigin)
	if output.PublicAccessBlockConfiguration != nil {
		if err := d.Set("public_access_block_configuration", []any{flattenPublicAccessBlockConfiguration(output.PublicAccessBlockConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting public_access_block_configuration: %s", err)
		}
	} else {
		d.Set("public_access_block_configuration", nil)
	}
	if output.VpcConfiguration != nil {
		if err := d.Set(names.AttrVPCConfiguration, []any{flattenVPCConfiguration(output.VpcConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_configuration: %s", err)
		}
	} else {
		d.Set(names.AttrVPCConfiguration, nil)
	}

	policy, err := FindAccessPointForDirectoryBucketPolicyByTwoPartKey(ctx, conn, accountID, name)

	if err == nil && policy != "" {
		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), policy)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		d.Set(names.AttrPolicy, policyToSet)
	} else if policy == "" || tfresource.NotFound(err) {
		d.Set(names.AttrPolicy, nil)
	} else {
		return sdkdiag.AppendErrorf(diags, "reading S3 Access Point for Directory Bucket (%s) policy: %s", d.Id(), err)
	}

	scope, err := FindAccessPointScopeByTwoPartKey(ctx, conn, accountID, name)
	if err == nil && scope != nil {
		flattened := flattenScope(scope)
		if err := d.Set(names.AttrScope, []any{flattened}); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	} else if tfresource.NotFound(err) || scope == nil {
		if err := d.Set(names.AttrScope, []any{}); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	} else {
		return sdkdiag.AppendErrorf(diags, "reading S3 Access Point for Directory Bucket (%s) scope: %s", d.Id(), err)
	}

	return diags
}

func resourceAccessPointForDirectoryBucketUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	name, accountID, err := AccessPointForDirectoryBucketParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChange(names.AttrPolicy) {
		if v, ok := d.GetOk(names.AttrPolicy); ok && v.(string) != "" && v.(string) != "{}" {
			policy, err := structure.NormalizeJsonString(v.(string))
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
				return sdkdiag.AppendErrorf(diags, "updating S3 Access Point for Directory Bucket (%s) policy: %s", d.Id(), err)
			}
		} else {
			input := &s3control.DeleteAccessPointPolicyInput{
				AccountId: aws.String(accountID),
				Name:      aws.String(name),
			}

			_, err := conn.DeleteAccessPointPolicy(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Directory S3 Access Point (%s) policy: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange(names.AttrScope) {
		if v, ok := d.GetOk(names.AttrScope); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			scope := expandScope(v.([]any)[0].(map[string]any))

			input := &s3control.PutAccessPointScopeInput{
				AccountId: aws.String(accountID),
				Name:      aws.String(name),
				Scope:     scope,
			}

			_, err = conn.PutAccessPointScope(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating S3 Access Point for Directory Bucket (%s) scope: %s", d.Id(), err)
			}
		} else {
			input := &s3control.DeleteAccessPointScopeInput{
				AccountId: aws.String(accountID),
				Name:      aws.String(name),
			}

			_, err := conn.DeleteAccessPointScope(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Directory S3 Access Point (%s) scope: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceAccessPointForDirectoryBucketRead(ctx, d, meta)...)
}

func resourceAccessPointForDirectoryBucketDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	name, accountID, err := AccessPointForDirectoryBucketParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting S3 Access Point for Directory Bucket: %s", d.Id())
	_, err = conn.DeleteAccessPoint(ctx, &s3control.DeleteAccessPointInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Access Point for Directory Bucket (%s): %s", d.Id(), err)
	}

	return diags
}

func AccessPointForDirectoryBucketCreateResourceID(accessPointName string, accountID string) (string, error) {

	if accessPointName == "" || accountID == "" {
		return "", fmt.Errorf("unexpected access point name for directory bucket: %s or accountID: %s", accessPointName, accountID)
	}
	id := accessPointName + accessPointResourceIDSeparator + accountID
	return id, nil
}

func AccessPointForDirectoryBucketParseResourceID(id string) (string, string, error) {
	log.Printf("[DEBUG] Parsing Resource ID: %s", id)

	parts := strings.Split(id, accessPointResourceIDSeparator)
	log.Printf("[DEBUG] Split parts: %#v", parts)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		log.Printf("[DEBUG] Parsed name: %s, accountID: %s", parts[0], parts[1])
		return parts[0], parts[1], nil
	}

	log.Printf("[ERROR] Invalid resource ID format: %s", id)
	return "", "", fmt.Errorf(
		"unexpected format for ID (%[1]s), expected name%[2]saccount-id",
		id, accessPointResourceIDSeparator,
	)
}

func expandScope(tfMap map[string]any) *types.Scope {
	if tfMap == nil {
		return nil
	}

	scope := &types.Scope{}

	if v, ok := tfMap["permissions"].([]any); ok {
		for _, perm := range v {
			if permStr, ok := perm.(string); ok {
				scope.Permissions = append(scope.Permissions, types.ScopePermission(permStr))
			}
		}
	}

	if v, ok := tfMap["prefixes"].([]any); ok {
		for _, prefix := range v {
			if prefixStr, ok := prefix.(string); ok {
				scope.Prefixes = append(scope.Prefixes, prefixStr)
			}
		}
	}

	if len(scope.Permissions) == 0 && len(scope.Prefixes) == 0 {
		return &types.Scope{
			Permissions: []types.ScopePermission{},
			Prefixes:    []string{},
		}
	}

	return scope
}

func flattenScope(scope *types.Scope) map[string]any {
	permissions := []string{}
	prefixes := []string{}

	if scope != nil {
		for _, p := range scope.Permissions {
			permissions = append(permissions, string(p))
		}
		for _, p := range scope.Prefixes {
			prefixes = append(prefixes, p)
		}
	}

	return map[string]any{
		"permissions": permissions,
		"prefixes":    prefixes,
	}
}

func FindAccessPointScopeByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, name string) (*types.Scope, error) {
	inputGAPS := &s3control.GetAccessPointScopeInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	outputGAPS, err := conn.GetAccessPointScope(ctx, inputGAPS)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: inputGAPS,
		}
	}

	if err != nil {
		return nil, err
	}

	if outputGAPS == nil {
		return nil, tfresource.NewEmptyResultError(inputGAPS)
	}

	return outputGAPS.Scope, nil
}

func FindAccessPointForDirectoryBucketPolicyByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, name string) (string, error) {
	inputGAPP := &s3control.GetAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	outputGAPP, err := conn.GetAccessPointPolicy(ctx, inputGAPP)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return "", &retry.NotFoundError{
			LastError:   err,
			LastRequest: inputGAPP,
		}
	}

	if err != nil {
		return "", err
	}

	if outputGAPP == nil {
		return "", tfresource.NewEmptyResultError(inputGAPP)
	}

	policy := aws.ToString(outputGAPP.Policy)

	if policy == "" {
		return "", tfresource.NewEmptyResultError(inputGAPP)
	}

	return policy, nil
}
