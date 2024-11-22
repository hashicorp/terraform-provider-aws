// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const BucketACLSeparator = ","

// @SDKResource("aws_s3_bucket_acl", name="Bucket ACL")
func resourceBucketACL() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketACLCreate,
		ReadWithoutTimeout:   resourceBucketACLRead,
		UpdateWithoutTimeout: resourceBucketACLUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_control_policy": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"access_control_policy", "acl"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grant": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"grantee": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"email_address": {
													Type:     schema.TypeString,
													Optional: true,
												},
												names.AttrDisplayName: {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrID: {
													Type:     schema.TypeString,
													Optional: true,
												},
												names.AttrType: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.Type](),
												},
												names.AttrURI: {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"permission": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.Permission](),
									},
								},
							},
						},
						names.AttrOwner: {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDisplayName: {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									names.AttrID: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"acl": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"access_control_policy", "acl"},
				ValidateFunc: validation.StringInSlice(bucketCannedACL_Values(), false),
			},
			names.AttrBucket: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			names.AttrExpectedBucketOwner: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
			if d.HasChange("acl") {
				_, n := d.GetChange("acl")
				if n.(string) != "" {
					return d.SetNewComputed("access_control_policy")
				}
			}

			return nil
		},
	}
}

func resourceBucketACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	expectedBucketOwner := d.Get(names.AttrExpectedBucketOwner).(string)
	acl := d.Get("acl").(string)
	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
	}
	if acl != "" {
		input.ACL = types.BucketCannedACL(acl)
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("access_control_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AccessControlPolicy = expandAccessControlPolicy(v.([]interface{}))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketAcl(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotImplemented) {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) ACL: %s", bucket, err)
	}

	d.SetId(BucketACLCreateResourceID(bucket, expectedBucketOwner, acl))

	_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findBucketACL(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket ACL (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketACLRead(ctx, d, meta)...)
}

func resourceBucketACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, acl, err := BucketACLParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	bucketACL, err := findBucketACL(ctx, conn, bucket, expectedBucketOwner)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket ACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket ACL (%s): %s", d.Id(), err)
	}

	if err := d.Set("access_control_policy", flattenBucketACL(bucketACL)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting access_control_policy: %s", err)
	}
	d.Set("acl", acl)
	d.Set(names.AttrBucket, bucket)
	d.Set(names.AttrExpectedBucketOwner, expectedBucketOwner)

	return diags
}

func resourceBucketACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, acl, err := BucketACLParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if d.HasChange("access_control_policy") {
		input.AccessControlPolicy = expandAccessControlPolicy(d.Get("access_control_policy").([]interface{}))
	}

	if d.HasChange("acl") {
		acl = d.Get("acl").(string)
		input.ACL = types.BucketCannedACL(acl)
	}

	_, err = conn.PutBucketAcl(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 bucket ACL (%s): %s", d.Id(), err)
	}

	if d.HasChange("acl") {
		// Set new ACL value back in resource ID
		d.SetId(BucketACLCreateResourceID(bucket, expectedBucketOwner, acl))
	}

	return append(diags, resourceBucketACLRead(ctx, d, meta)...)
}

func findBucketACL(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*s3.GetBucketAclOutput, error) {
	input := &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketAcl(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandAccessControlPolicy(l []interface{}) *types.AccessControlPolicy {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.AccessControlPolicy{}

	if v, ok := tfMap["grant"].(*schema.Set); ok && v.Len() > 0 {
		result.Grants = expandGrants(v.List())
	}

	if v, ok := tfMap[names.AttrOwner].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.Owner = expandOwner(v)
	}

	return result
}

func expandGrants(l []interface{}) []types.Grant {
	var grants []types.Grant

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		grant := types.Grant{}

		if v, ok := tfMap["grantee"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			grant.Grantee = expandACLGrantee(v)
		}

		if v, ok := tfMap["permission"].(string); ok && v != "" {
			grant.Permission = types.Permission(v)
		}

		grants = append(grants, grant)
	}

	return grants
}

func expandACLGrantee(l []interface{}) *types.Grantee {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.Grantee{}

	if v, ok := tfMap["email_address"].(string); ok && v != "" {
		result.EmailAddress = aws.String(v)
	}

	if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
		result.ID = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		result.Type = types.Type(v)
	}

	if v, ok := tfMap[names.AttrURI].(string); ok && v != "" {
		result.URI = aws.String(v)
	}

	return result
}

func expandOwner(l []interface{}) *types.Owner {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	owner := &types.Owner{}

	if v, ok := tfMap[names.AttrDisplayName].(string); ok && v != "" {
		owner.DisplayName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
		owner.ID = aws.String(v)
	}

	return owner
}

func flattenBucketACL(apiObject *s3.GetBucketAclOutput) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if len(apiObject.Grants) > 0 {
		m["grant"] = flattenGrants(apiObject.Grants)
	}

	if apiObject.Owner != nil {
		m[names.AttrOwner] = flattenOwner(apiObject.Owner)
	}

	return []interface{}{m}
}

func flattenGrants(grants []types.Grant) []interface{} {
	var results []interface{}

	for _, grant := range grants {
		m := map[string]interface{}{
			"permission": grant.Permission,
		}

		if grant.Grantee != nil {
			m["grantee"] = flattenACLGrantee(grant.Grantee)
		}

		results = append(results, m)
	}

	return results
}

func flattenACLGrantee(grantee *types.Grantee) []interface{} {
	if grantee == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrType: grantee.Type,
	}

	if grantee.DisplayName != nil {
		m[names.AttrDisplayName] = aws.ToString(grantee.DisplayName)
	}

	if grantee.EmailAddress != nil {
		m["email_address"] = aws.ToString(grantee.EmailAddress)
	}

	if grantee.ID != nil {
		m[names.AttrID] = aws.ToString(grantee.ID)
	}

	if grantee.URI != nil {
		m[names.AttrURI] = aws.ToString(grantee.URI)
	}

	return []interface{}{m}
}

func flattenOwner(owner *types.Owner) []interface{} {
	if owner == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if owner.DisplayName != nil {
		m[names.AttrDisplayName] = aws.ToString(owner.DisplayName)
	}

	if owner.ID != nil {
		m[names.AttrID] = aws.ToString(owner.ID)
	}

	return []interface{}{m}
}

// BucketACLCreateResourceID is a method for creating an ID string
// with the bucket name and optional accountID and/or ACL.
func BucketACLCreateResourceID(bucket, expectedBucketOwner, acl string) string {
	if expectedBucketOwner == "" {
		if acl == "" {
			return bucket
		}
		return strings.Join([]string{bucket, acl}, BucketACLSeparator)
	}

	if acl == "" {
		return strings.Join([]string{bucket, expectedBucketOwner}, BucketACLSeparator)
	}

	return strings.Join([]string{bucket, expectedBucketOwner, acl}, BucketACLSeparator)
}

// BucketACLParseResourceID is a method for parsing the ID string
// for the bucket name, accountID, and ACL if provided.
func BucketACLParseResourceID(id string) (string, string, string, error) {
	// For only bucket name in the ID  e.g. my-bucket or My_Bucket
	// ~> On or after 3/1/2018: Bucket names can consist of only lowercase letters, numbers, dots, and hyphens; Max 63 characters
	// ~> Before 3/1/2018: Bucket names could consist of uppercase letters and underscores if in us-east-1; Max 255 characters
	// Reference: https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html
	bucketRegex := regexache.MustCompile(`^([0-9a-z.-]{1,63}|[0-9A-Za-z_.-]{1,255})$`)
	// For bucket and accountID in the ID e.g. bucket,123456789101
	// ~> Account IDs must consist of 12 digits
	bucketAndOwnerRegex := regexache.MustCompile(`^([0-9a-z.-]{1,63}|[0-9A-Za-z_.-]{1,255}),\d{12}$`)
	// For bucket and ACL in the ID e.g. bucket,public-read
	// ~> (Canned) ACL values include: private, public-read, public-read-write, authenticated-read, aws-exec-read, and log-delivery-write
	bucketAndAclRegex := regexache.MustCompile(`^([0-9a-z.-]{1,63}|[0-9A-Za-z_.-]{1,255}),[a-z-]+$`)
	// For bucket, accountID, and ACL in the ID e.g. bucket,123456789101,public-read
	bucketOwnerAclRegex := regexache.MustCompile(`^([0-9a-z.-]{1,63}|[0-9A-Za-z_.-]{1,255}),\d{12},[a-z-]+$`)

	// Bucket name ONLY
	if bucketRegex.MatchString(id) {
		return id, "", "", nil
	}

	// Bucket and Account ID ONLY
	if bucketAndOwnerRegex.MatchString(id) {
		parts := strings.Split(id, BucketACLSeparator)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", "", fmt.Errorf("unexpected format for ID (%s), expected BUCKET%sEXPECTED_BUCKET_OWNER", id, BucketACLSeparator)
		}
		return parts[0], parts[1], "", nil
	}

	// Bucket and ACL ONLY
	if bucketAndAclRegex.MatchString(id) {
		parts := strings.Split(id, BucketACLSeparator)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", "", fmt.Errorf("unexpected format for ID (%s), expected BUCKET%sACL", id, BucketACLSeparator)
		}
		return parts[0], "", parts[1], nil
	}

	// Bucket, Account ID, and ACL
	if bucketOwnerAclRegex.MatchString(id) {
		parts := strings.Split(id, BucketACLSeparator)
		if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return "", "", "", fmt.Errorf("unexpected format for ID (%s), expected BUCKET%[2]sEXPECTED_BUCKET_OWNER%[2]sACL", id, BucketACLSeparator)
		}
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%s), expected BUCKET or BUCKET%[2]sEXPECTED_BUCKET_OWNER or BUCKET%[2]sACL "+
		"or BUCKET%[2]sEXPECTED_BUCKET_OWNER%[2]sACL", id, BucketACLSeparator)
}

// These should be defined in the AWS SDK for Go. There is an issue, https://github.com/aws/aws-sdk-go/issues/2683.
const (
	bucketCannedACLExecRead         = "aws-exec-read"
	bucketCannedACLLogDeliveryWrite = "log-delivery-write"
)

func bucketCannedACL_Values() []string {
	return tfslices.AppendUnique(enum.Values[types.BucketCannedACL](), bucketCannedACLExecRead, bucketCannedACLLogDeliveryWrite)
}
