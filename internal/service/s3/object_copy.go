// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_object_copy", name="Object")
// @Tags
func ResourceObjectCopy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceObjectCopyCreate,
		ReadWithoutTimeout:   resourceObjectCopyRead,
		UpdateWithoutTimeout: resourceObjectCopyUpdate,
		DeleteWithoutTimeout: resourceObjectCopyDelete,

		Schema: map[string]*schema.Schema{
			"acl": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringInSlice(s3.ObjectCannedACL_Values(), false),
				ConflictsWith: []string{"grant"},
			},
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"bucket_key_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"cache_control": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"content_disposition": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"content_encoding": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"content_language": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"content_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"copy_if_match": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"copy_if_modified_since": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"copy_if_none_match": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"copy_if_unmodified_since": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"customer_algorithm": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"customer_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"customer_key_md5": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expected_bucket_owner": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"expected_source_bucket_owner": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"grant": {
				Type:          schema.TypeSet,
				Optional:      true,
				Set:           grantHash,
				ConflictsWith: []string{"acl"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"permissions": {
							Type:     schema.TypeSet,
							Required: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									//write permission not valid here
									s3.PermissionFullControl,
									s3.PermissionRead,
									s3.PermissionReadAcp,
									s3.PermissionWriteAcp,
								}, false),
							},
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
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"kms_encryption_context": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
				Sensitive:    true,
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
				Sensitive:    true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:         schema.TypeMap,
				ValidateFunc: validateMetadataIsLowerCase,
				Optional:     true,
				Computed:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
			},
			"metadata_directive": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(s3.MetadataDirective_Values(), false),
			},
			"object_lock_legal_hold_status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(s3.ObjectLockLegalHoldStatus_Values(), false),
			},
			"object_lock_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(s3.ObjectLockMode_Values(), false),
			},
			"object_lock_retain_until_date": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"request_charged": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"request_payer": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(s3.RequestPayer_Values(), false),
			},
			"server_side_encryption": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(s3.ServerSideEncryption_Values(), false),
			},
			"source": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"source_customer_algorithm": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_customer_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"source_customer_key_md5": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_class": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(s3.ObjectStorageClass_Values(), false),
			},
			"tagging_directive": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(s3.TaggingDirective_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website_redirect": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceObjectCopyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return append(diags, resourceObjectCopyDoCopy(ctx, d, meta)...)
}

func resourceObjectCopyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	output, err := FindObjectByThreePartKey(ctx, conn, bucket, key, "")

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Object (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Object (%s): %s", d.Id(), err)
	}

	d.Set("bucket_key_enabled", output.BucketKeyEnabled)
	d.Set("cache_control", output.CacheControl)
	d.Set("content_disposition", output.ContentDisposition)
	d.Set("content_encoding", output.ContentEncoding)
	d.Set("content_language", output.ContentLanguage)
	d.Set("content_type", output.ContentType)
	metadata := flex.PointersMapToStringList(output.Metadata)

	// AWS Go SDK capitalizes metadata, this is a workaround. https://github.com/aws/aws-sdk-go/issues/445
	for k, v := range metadata {
		delete(metadata, k)
		metadata[strings.ToLower(k)] = v
	}

	if err := d.Set("metadata", metadata); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting metadata: %s", err)
	}
	d.Set("version_id", output.VersionId)
	d.Set("server_side_encryption", output.ServerSideEncryption)
	d.Set("website_redirect", output.WebsiteRedirectLocation)
	d.Set("object_lock_legal_hold_status", output.ObjectLockLegalHoldStatus)
	d.Set("object_lock_mode", output.ObjectLockMode)
	d.Set("object_lock_retain_until_date", flattenObjectDate(output.ObjectLockRetainUntilDate))

	if err := resourceObjectSetKMS(ctx, d, meta, output.SSEKMSKeyId); err != nil {
		return sdkdiag.AppendErrorf(diags, "object KMS: %s", err)
	}

	// See https://forums.aws.amazon.com/thread.jspa?threadID=44003
	d.Set("etag", strings.Trim(aws.StringValue(output.ETag), `"`))

	// The "STANDARD" (which is also the default) storage
	// class when set would not be included in the results.
	d.Set("storage_class", s3.ObjectStorageClassStandard)
	if output.StorageClass != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("storage_class", output.StorageClass)
	}

	// Retry due to S3 eventual consistency
	tagsRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return ObjectListTags(ctx, conn, bucket, key)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for S3 Bucket (%s) Object (%s): %s", bucket, key, err)
	}

	tags, ok := tagsRaw.(tftags.KeyValueTags)

	if !ok {
		return sdkdiag.AppendErrorf(diags, "listing tags for S3 Bucket (%s) Object (%s): unable to convert tags", bucket, key)
	}

	setTagsOut(ctx, Tags(tags))

	return diags
}

func resourceObjectCopyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var
	// if any of these exist, let the API decide whether to copy
	diags diag.Diagnostics

	for _, key := range []string{
		"copy_if_match",
		"copy_if_modified_since",
		"copy_if_none_match",
		"copy_if_unmodified_since",
	} {
		if _, ok := d.GetOk(key); ok {
			return append(diags, resourceObjectCopyDoCopy(ctx, d, meta)...)
		}
	}

	args := []string{
		"acl",
		"bucket",
		"bucket_key_enabled",
		"cache_control",
		"content_disposition",
		"content_encoding",
		"content_language",
		"content_type",
		"customer_algorithm",
		"customer_key",
		"customer_key_md5",
		"expected_bucket_owner",
		"expected_source_bucket_owner",
		"expires",
		"grant",
		"key",
		"kms_encryption_context",
		"kms_key_id",
		"metadata",
		"metadata_directive",
		"object_lock_legal_hold_status",
		"object_lock_mode",
		"object_lock_retain_until_date",
		"request_payer",
		"server_side_encryption",
		"source",
		"source_customer_algorithm",
		"source_customer_key",
		"source_customer_key_md5",
		"storage_class",
		"tagging_directive",
		"tags",
		"tags_all",
		"website_redirect",
	}
	if d.HasChanges(args...) {
		return append(diags, resourceObjectCopyDoCopy(ctx, d, meta)...)
	}

	return diags
}

func resourceObjectCopyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	// We are effectively ignoring all leading '/'s in the key name and
	// treating multiple '/'s as a single '/' as aws.Config.DisableRestProtocolURICleaning is false
	key = strings.TrimLeft(key, "/")
	key = regexp.MustCompile(`/+`).ReplaceAllString(key, "/")

	var err error
	if _, ok := d.GetOk("version_id"); ok {
		_, err = DeleteAllObjectVersions(ctx, conn, bucket, key, d.Get("force_destroy").(bool), false)
	} else {
		err = deleteObjectVersion(ctx, conn, bucket, key, "", false)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket (%s) Object (%s): %s", bucket, key, err)
	}
	return diags
}

func resourceObjectCopyDoCopy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	input := &s3.CopyObjectInput{
		Bucket:     aws.String(d.Get("bucket").(string)),
		Key:        aws.String(d.Get("key").(string)),
		CopySource: aws.String(url.QueryEscape(d.Get("source").(string))),
	}

	if v, ok := d.GetOk("acl"); ok {
		input.ACL = aws.String(v.(string))
	}

	if v, ok := d.GetOk("bucket_key_enabled"); ok {
		input.BucketKeyEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("cache_control"); ok {
		input.CacheControl = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_disposition"); ok {
		input.ContentDisposition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_encoding"); ok {
		input.ContentEncoding = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_language"); ok {
		input.ContentLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_type"); ok {
		input.ContentType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("copy_if_match"); ok {
		input.CopySourceIfMatch = aws.String(v.(string))
	}

	if v, ok := d.GetOk("copy_if_modified_since"); ok {
		input.CopySourceIfModifiedSince = expandObjectDate(v.(string))
	}

	if v, ok := d.GetOk("copy_if_none_match"); ok {
		input.CopySourceIfNoneMatch = aws.String(v.(string))
	}

	if v, ok := d.GetOk("copy_if_unmodified_since"); ok {
		input.CopySourceIfUnmodifiedSince = expandObjectDate(v.(string))
	}

	if v, ok := d.GetOk("customer_algorithm"); ok {
		input.SSECustomerAlgorithm = aws.String(v.(string))
	}

	if v, ok := d.GetOk("customer_key"); ok {
		input.SSECustomerKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("customer_key_md5"); ok {
		input.SSECustomerKeyMD5 = aws.String(v.(string))
	}

	if v, ok := d.GetOk("expected_bucket_owner"); ok {
		input.ExpectedBucketOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("expected_source_bucket_owner"); ok {
		input.ExpectedSourceBucketOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("expires"); ok {
		input.Expires = expandObjectDate(v.(string))
	}

	if v, ok := d.GetOk("grant"); ok && v.(*schema.Set).Len() > 0 {
		grants := expandObjectCopyGrants(v.(*schema.Set).List())
		input.GrantFullControl = grants.FullControl
		input.GrantRead = grants.Read
		input.GrantReadACP = grants.ReadACP
		input.GrantWriteACP = grants.WriteACP
		input.ACL = nil
	}

	if v, ok := d.GetOk("kms_encryption_context"); ok {
		input.SSEKMSEncryptionContext = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.SSEKMSKeyId = aws.String(v.(string))
		input.ServerSideEncryption = aws.String(s3.ServerSideEncryptionAwsKms)
	}

	if v, ok := d.GetOk("metadata"); ok {
		input.Metadata = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("metadata_directive"); ok {
		input.MetadataDirective = aws.String(v.(string))
	}

	if v, ok := d.GetOk("object_lock_legal_hold_status"); ok {
		input.ObjectLockLegalHoldStatus = aws.String(v.(string))
	}

	if v, ok := d.GetOk("object_lock_mode"); ok {
		input.ObjectLockMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("object_lock_retain_until_date"); ok {
		input.ObjectLockRetainUntilDate = expandObjectDate(v.(string))
	}

	if v, ok := d.GetOk("request_payer"); ok {
		input.RequestPayer = aws.String(v.(string))
	}

	if v, ok := d.GetOk("server_side_encryption"); ok {
		input.ServerSideEncryption = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_customer_algorithm"); ok {
		input.CopySourceSSECustomerAlgorithm = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_customer_key"); ok {
		input.CopySourceSSECustomerKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_customer_key_md5"); ok {
		input.CopySourceSSECustomerKeyMD5 = aws.String(v.(string))
	}

	if v, ok := d.GetOk("storage_class"); ok {
		input.StorageClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tagging_directive"); ok {
		input.TaggingDirective = aws.String(v.(string))
	}

	if len(tags) > 0 {
		// The tag-set must be encoded as URL Query parameters.
		input.Tagging = aws.String(tags.IgnoreAWS().URLEncode())
	}

	if v, ok := d.GetOk("website_redirect"); ok {
		input.WebsiteRedirectLocation = aws.String(v.(string))
	}

	output, err := conn.CopyObjectWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "copying S3 object (bucket: %s; key: %s; source: %s): %s", aws.StringValue(input.Bucket), aws.StringValue(input.Key), aws.StringValue(input.CopySource), err)
	}

	d.Set("customer_algorithm", output.SSECustomerAlgorithm)
	d.Set("customer_key_md5", output.SSECustomerKeyMD5)

	if output.CopyObjectResult != nil {
		d.Set("etag", strings.Trim(aws.StringValue(output.CopyObjectResult.ETag), `"`))
		d.Set("last_modified", flattenObjectDate(output.CopyObjectResult.LastModified))
	}

	d.Set("expiration", output.Expiration)
	d.Set("kms_encryption_context", output.SSEKMSEncryptionContext)
	d.Set("kms_key_id", output.SSEKMSKeyId)
	d.Set("request_charged", output.RequestCharged)
	d.Set("server_side_encryption", output.ServerSideEncryption)
	d.Set("source_version_id", output.CopySourceVersionId)
	d.Set("version_id", output.VersionId)

	d.SetId(d.Get("key").(string))

	return append(diags, resourceObjectRead(ctx, d, meta)...)
}

type s3Grants struct {
	FullControl *string
	Read        *string
	ReadACP     *string
	WriteACP    *string
}

func expandObjectCopyGrant(tfMap map[string]interface{}) string {
	if tfMap == nil {
		return ""
	}

	apiObject := &s3.Grantee{}

	if v, ok := tfMap["email"].(string); ok && v != "" {
		apiObject.SetEmailAddress(v)
	}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.SetID(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.SetType(v)
	}

	if v, ok := tfMap["uri"].(string); ok && v != "" {
		apiObject.SetURI(v)
	}

	// Examples:
	//"GrantFullControl": "emailaddress=user1@example.com,emailaddress=user2@example.com",
	//"GrantRead": "uri=http://acs.amazonaws.com/groups/global/AllUsers",
	//"GrantFullControl": "id=examplee7a2f25102679df27bb0ae12b3f85be6f290b936c4393484",
	//"GrantWrite": "uri=http://acs.amazonaws.com/groups/s3/LogDelivery"

	switch *apiObject.Type {
	case s3.TypeAmazonCustomerByEmail:
		return fmt.Sprintf("emailaddress=%s", *apiObject.EmailAddress)
	case s3.TypeCanonicalUser:
		return fmt.Sprintf("id=%s", *apiObject.ID)
	}

	return fmt.Sprintf("uri=%s", *apiObject.URI)
}

func expandObjectCopyGrants(tfList []interface{}) *s3Grants {
	if len(tfList) == 0 {
		return nil
	}

	grantFullControl := make([]string, 0)
	grantRead := make([]string, 0)
	grantReadACP := make([]string, 0)
	grantWriteACP := make([]string, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		for _, perm := range tfMap["permissions"].(*schema.Set).List() {
			if v := expandObjectCopyGrant(tfMap); v != "" {
				switch perm.(string) {
				case s3.PermissionFullControl:
					grantFullControl = append(grantFullControl, v)
				case s3.PermissionRead:
					grantRead = append(grantRead, v)
				case s3.PermissionReadAcp:
					grantReadACP = append(grantReadACP, v)
				case s3.PermissionWriteAcp:
					grantWriteACP = append(grantWriteACP, v)
				}
			}
		}
	}

	apiObjects := &s3Grants{}

	if len(grantFullControl) > 0 {
		apiObjects.FullControl = aws.String(strings.Join(grantFullControl, ","))
	}

	if len(grantRead) > 0 {
		apiObjects.Read = aws.String(strings.Join(grantRead, ","))
	}

	if len(grantReadACP) > 0 {
		apiObjects.ReadACP = aws.String(strings.Join(grantReadACP, ","))
	}

	if len(grantWriteACP) > 0 {
		apiObjects.WriteACP = aws.String(strings.Join(grantWriteACP, ","))
	}

	return apiObjects
}

func grantHash(v interface{}) int {
	var buf bytes.Buffer
	m, ok := v.(map[string]interface{})

	if !ok {
		return 0
	}

	if v, ok := m["id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["uri"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if p, ok := m["permissions"]; ok {
		buf.WriteString(fmt.Sprintf("%v-", p.(*schema.Set).List()))
	}
	return create.StringHashcode(buf.String())
}
