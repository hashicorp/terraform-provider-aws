package s3

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const BucketACLSeparator = ","

func ResourceBucketACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketACLCreate,
		ReadContext:   resourceBucketACLRead,
		UpdateContext: resourceBucketACLUpdate,
		DeleteContext: resourceBucketACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_control_policy": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"acl"},
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
												"display_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"id": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(s3.Type_Values(), false),
												},
												"uri": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"permission": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(s3.Permission_Values(), false),
									},
								},
							},
						},
						"owner": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"display_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"id": {
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
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"access_control_policy"},
				ValidateFunc:  validation.StringInSlice(BucketCannedACL_Values(), false),
			},
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			"expected_bucket_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},
	}
}

func resourceBucketACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)
	acl := d.Get("acl").(string)

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if acl != "" {
		input.ACL = aws.String(acl)
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("access_control_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AccessControlPolicy = expandBucketACLAccessControlPolicy(v.([]interface{}))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketAclWithContext(ctx, input)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating S3 bucket ACL for %s: %w", bucket, err))
	}

	d.SetId(BucketACLCreateResourceID(bucket, expectedBucketOwner, acl))

	return resourceBucketACLRead(ctx, d, meta)
}

func resourceBucketACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, acl, err := BucketACLParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketAclWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket ACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting S3 bucket ACL (%s): %w", d.Id(), err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error getting S3 bucket ACL (%s): empty output", d.Id()))
	}

	d.Set("acl", acl)
	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)
	if err := d.Set("access_control_policy", flattenBucketACLAccessControlPolicy(output)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting access_control_policy: %w", err))
	}

	return nil
}

func resourceBucketACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, acl, err := BucketACLParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if d.HasChange("access_control_policy") {
		input.AccessControlPolicy = expandBucketACLAccessControlPolicy(d.Get("access_control_policy").([]interface{}))
	}

	if d.HasChange("acl") {
		acl = d.Get("acl").(string)
		input.ACL = aws.String(acl)
	}

	_, err = conn.PutBucketAclWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating S3 bucket ACL (%s): %w", d.Id(), err))
	}

	if d.HasChange("acl") {
		// Set new ACL value back in resource ID
		d.SetId(BucketACLCreateResourceID(bucket, expectedBucketOwner, acl))
	}

	return resourceBucketACLRead(ctx, d, meta)
}

func resourceBucketACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[WARN] Cannot destroy S3 Bucket ACL. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

func expandBucketACLAccessControlPolicy(l []interface{}) *s3.AccessControlPolicy {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3.AccessControlPolicy{}

	if v, ok := tfMap["grant"].(*schema.Set); ok && v.Len() > 0 {
		result.Grants = expandBucketACLAccessControlPolicyGrants(v.List())
	}

	if v, ok := tfMap["owner"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.Owner = expandBucketACLAccessControlPolicyOwner(v)
	}

	return result
}

func expandBucketACLAccessControlPolicyGrants(l []interface{}) []*s3.Grant {
	var grants []*s3.Grant

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		grant := &s3.Grant{}

		if v, ok := tfMap["grantee"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			grant.Grantee = expandBucketACLAccessControlPolicyGrantsGrantee(v)
		}

		if v, ok := tfMap["permission"].(string); ok && v != "" {
			grant.Permission = aws.String(v)
		}

		grants = append(grants, grant)
	}

	return grants
}

func expandBucketACLAccessControlPolicyGrantsGrantee(l []interface{}) *s3.Grantee {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3.Grantee{}

	if v, ok := tfMap["email_address"].(string); ok && v != "" {
		result.EmailAddress = aws.String(v)
	}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		result.ID = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		result.Type = aws.String(v)
	}

	if v, ok := tfMap["uri"].(string); ok && v != "" {
		result.URI = aws.String(v)
	}

	return result
}

func expandBucketACLAccessControlPolicyOwner(l []interface{}) *s3.Owner {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	owner := &s3.Owner{}

	if v, ok := tfMap["display_name"].(string); ok && v != "" {
		owner.DisplayName = aws.String(v)
	}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		owner.ID = aws.String(v)
	}

	return owner
}

func flattenBucketACLAccessControlPolicy(output *s3.GetBucketAclOutput) []interface{} {
	if output == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if len(output.Grants) > 0 {
		m["grant"] = flattenBucketACLAccessControlPolicyGrants(output.Grants)
	}

	if output.Owner != nil {
		m["owner"] = flattenBucketACLAccessControlPolicyOwner(output.Owner)
	}

	return []interface{}{m}
}

func flattenBucketACLAccessControlPolicyGrants(grants []*s3.Grant) []interface{} {
	var results []interface{}

	for _, grant := range grants {
		if grant == nil {
			continue
		}

		m := make(map[string]interface{})

		if grant.Grantee != nil {
			m["grantee"] = flattenBucketACLAccessControlPolicyGrantsGrantee(grant.Grantee)
		}

		if grant.Permission != nil {
			m["permission"] = aws.StringValue(grant.Permission)
		}

		results = append(results, m)
	}

	return results
}

func flattenBucketACLAccessControlPolicyGrantsGrantee(grantee *s3.Grantee) []interface{} {
	if grantee == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if grantee.DisplayName != nil {
		m["display_name"] = aws.StringValue(grantee.DisplayName)
	}

	if grantee.EmailAddress != nil {
		m["email_address"] = aws.StringValue(grantee.EmailAddress)
	}

	if grantee.ID != nil {
		m["id"] = aws.StringValue(grantee.ID)
	}

	if grantee.Type != nil {
		m["type"] = aws.StringValue(grantee.Type)
	}

	if grantee.URI != nil {
		m["uri"] = aws.StringValue(grantee.URI)
	}

	return []interface{}{m}
}

func flattenBucketACLAccessControlPolicyOwner(owner *s3.Owner) []interface{} {
	if owner == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if owner.DisplayName != nil {
		m["display_name"] = aws.StringValue(owner.DisplayName)
	}

	if owner.ID != nil {
		m["id"] = aws.StringValue(owner.ID)
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
	bucketRegex := regexp.MustCompile(`^([a-z0-9.-]{1,63}|[a-zA-Z0-9.\-_]{1,255})$`)
	// For bucket and accountID in the ID e.g. bucket,123456789101
	// ~> Account IDs must consist of 12 digits
	bucketAndOwnerRegex := regexp.MustCompile(`^([a-z0-9.-]{1,63}|[a-zA-Z0-9.\-_]{1,255}),\d{12}$`)
	// For bucket and ACL in the ID e.g. bucket,public-read
	// ~> (Canned) ACL values include: private, public-read, public-read-write, authenticated-read, aws-exec-read, and log-delivery-write
	bucketAndAclRegex := regexp.MustCompile(`^([a-z0-9.-]{1,63}|[a-zA-Z0-9.\-_]{1,255}),[a-z-]+$`)
	// For bucket, accountID, and ACL in the ID e.g. bucket,123456789101,public-read
	bucketOwnerAclRegex := regexp.MustCompile(`^([a-z0-9.-]{1,63}|[a-zA-Z0-9.\-_]{1,255}),\d{12},[a-z-]+$`)

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
