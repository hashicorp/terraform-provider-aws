// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.ObjectCannedACL](),
				ConflictsWith:    []string{"grant"},
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
			"checksum_algorithm": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.ChecksumAlgorithm](),
			},
			"checksum_crc32": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum_crc32c": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum_sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum_sha256": {
				Type:     schema.TypeString,
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
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(enum.Slice(
									//write permission not valid here
									types.PermissionFullControl,
									types.PermissionRead,
									types.PermissionReadAcp,
									types.PermissionWriteAcp,
								), false)),
							},
						},
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.Type](),
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.MetadataDirective](),
			},
			"object_lock_legal_hold_status": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.ObjectLockLegalHoldStatus](),
			},
			"object_lock_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.ObjectLockMode](),
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.RequestPayer](),
			},
			"server_side_encryption": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.ServerSideEncryption](),
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.ObjectStorageClass](),
			},
			"tagging_directive": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.TaggingDirective](),
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
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get("bucket").(string)
	key := sdkv1CompatibleCleanKey(d.Get("key").(string))
	output, err := findObjectByBucketAndKey(ctx, conn, bucket, key, "", d.Get("checksum_algorithm").(string))

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
	d.Set("checksum_crc32", output.ChecksumCRC32)
	d.Set("checksum_crc32c", output.ChecksumCRC32C)
	d.Set("checksum_sha1", output.ChecksumSHA1)
	d.Set("checksum_sha256", output.ChecksumSHA256)
	d.Set("content_disposition", output.ContentDisposition)
	d.Set("content_encoding", output.ContentEncoding)
	d.Set("content_language", output.ContentLanguage)
	d.Set("content_type", output.ContentType)
	d.Set("customer_algorithm", output.SSECustomerAlgorithm)
	d.Set("customer_key_md5", output.SSECustomerKeyMD5)
	// See https://forums.aws.amazon.com/thread.jspa?threadID=44003
	d.Set("etag", strings.Trim(aws.ToString(output.ETag), `"`))
	d.Set("expiration", output.Expiration)
	d.Set("kms_key_id", output.SSEKMSKeyId)
	d.Set("last_modified", flattenObjectDate(output.LastModified))
	d.Set("metadata", output.Metadata)
	d.Set("object_lock_legal_hold_status", output.ObjectLockLegalHoldStatus)
	d.Set("object_lock_mode", output.ObjectLockMode)
	d.Set("object_lock_retain_until_date", flattenObjectDate(output.ObjectLockRetainUntilDate))
	d.Set("server_side_encryption", output.ServerSideEncryption)
	// The "STANDARD" (which is also the default) storage
	// class when set would not be included in the results.
	d.Set("storage_class", types.ObjectStorageClassStandard)
	if output.StorageClass != "" {
		d.Set("storage_class", output.StorageClass)
	}
	d.Set("version_id", output.VersionId)
	d.Set("website_redirect", output.WebsiteRedirectLocation)

	if err := resourceObjectSetKMS(ctx, d, meta, output.SSEKMSKeyId); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if tags, err := ObjectListTags(ctx, conn, bucket, key); err == nil {
		setTagsOut(ctx, Tags(tags))
	} else if !tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotImplemented) { // Directory buckets return HTTP status code 501, NotImplemented.
		return sdkdiag.AppendErrorf(diags, "listing tags for S3 Bucket (%s) Object (%s): %s", bucket, key, err)
	}

	return diags
}

func resourceObjectCopyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// if any of these exist, let the API decide whether to copy
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
		"checksum_algorithm",
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
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get("bucket").(string)
	key := sdkv1CompatibleCleanKey(d.Get("key").(string))

	var err error
	if _, ok := d.GetOk("version_id"); ok {
		_, err = deleteAllObjectVersions(ctx, conn, bucket, key, d.Get("force_destroy").(bool), false)
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
	conn := meta.(*conns.AWSClient).S3Client(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	input := &s3.CopyObjectInput{
		Bucket:     aws.String(d.Get("bucket").(string)),
		CopySource: aws.String(url.QueryEscape(d.Get("source").(string))),
		Key:        aws.String(sdkv1CompatibleCleanKey(d.Get("key").(string))),
	}

	if v, ok := d.GetOk("acl"); ok {
		input.ACL = types.ObjectCannedACL(v.(string))
	}

	if v, ok := d.GetOk("bucket_key_enabled"); ok {
		input.BucketKeyEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("cache_control"); ok {
		input.CacheControl = aws.String(v.(string))
	}

	if v, ok := d.GetOk("checksum_algorithm"); ok {
		input.ChecksumAlgorithm = types.ChecksumAlgorithm(v.(string))
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
		input.ACL = ""
	}

	if v, ok := d.GetOk("kms_encryption_context"); ok {
		input.SSEKMSEncryptionContext = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.SSEKMSKeyId = aws.String(v.(string))
		input.ServerSideEncryption = types.ServerSideEncryptionAwsKms
	}

	if v, ok := d.GetOk("metadata"); ok {
		input.Metadata = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("metadata_directive"); ok {
		input.MetadataDirective = types.MetadataDirective(v.(string))
	}

	if v, ok := d.GetOk("object_lock_legal_hold_status"); ok {
		input.ObjectLockLegalHoldStatus = types.ObjectLockLegalHoldStatus(v.(string))
	}

	if v, ok := d.GetOk("object_lock_mode"); ok {
		input.ObjectLockMode = types.ObjectLockMode(v.(string))
	}

	if v, ok := d.GetOk("object_lock_retain_until_date"); ok {
		input.ObjectLockRetainUntilDate = expandObjectDate(v.(string))
	}

	if v, ok := d.GetOk("request_payer"); ok {
		input.RequestPayer = types.RequestPayer(v.(string))
	}

	if v, ok := d.GetOk("server_side_encryption"); ok {
		input.ServerSideEncryption = types.ServerSideEncryption(v.(string))
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
		input.StorageClass = types.StorageClass(v.(string))
	}

	if v, ok := d.GetOk("tagging_directive"); ok {
		input.TaggingDirective = types.TaggingDirective(v.(string))
	}

	if len(tags) > 0 {
		// The tag-set must be encoded as URL Query parameters.
		input.Tagging = aws.String(tags.IgnoreAWS().URLEncode())
	}

	if v, ok := d.GetOk("website_redirect"); ok {
		input.WebsiteRedirectLocation = aws.String(v.(string))
	}

	output, err := conn.CopyObject(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "copying %s to S3 Bucket (%s) Object (%s): %s", aws.ToString(input.CopySource), aws.ToString(input.Bucket), aws.ToString(input.Key), err)
	}

	if d.IsNewResource() {
		d.SetId(d.Get("key").(string))
	}

	// These attributes aren't returned from HeadObject.
	d.Set("kms_encryption_context", output.SSEKMSEncryptionContext)
	d.Set("request_charged", output.RequestCharged == types.RequestChargedRequester)
	d.Set("source_version_id", output.CopySourceVersionId)

	return append(diags, resourceObjectCopyRead(ctx, d, meta)...)
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

	apiObject := &types.Grantee{}

	if v, ok := tfMap["email"].(string); ok && v != "" {
		apiObject.EmailAddress = aws.String(v)
	}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.ID = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = types.Type(v)
	}

	if v, ok := tfMap["uri"].(string); ok && v != "" {
		apiObject.URI = aws.String(v)
	}

	// Examples:
	//"GrantFullControl": "emailaddress=user1@example.com,emailaddress=user2@example.com",
	//"GrantRead": "uri=http://acs.amazonaws.com/groups/global/AllUsers",
	//"GrantFullControl": "id=examplee7a2f25102679df27bb0ae12b3f85be6f290b936c4393484",
	//"GrantWrite": "uri=http://acs.amazonaws.com/groups/s3/LogDelivery"

	switch apiObject.Type {
	case types.TypeAmazonCustomerByEmail:
		return fmt.Sprintf("emailaddress=%s", aws.ToString(apiObject.EmailAddress))
	case types.TypeCanonicalUser:
		return fmt.Sprintf("id=%s", aws.ToString(apiObject.ID))
	}

	return fmt.Sprintf("uri=%s", aws.ToString(apiObject.URI))
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
				switch types.Permission(perm.(string)) {
				case types.PermissionFullControl:
					grantFullControl = append(grantFullControl, v)
				case types.PermissionRead:
					grantRead = append(grantRead, v)
				case types.PermissionReadAcp:
					grantReadACP = append(grantReadACP, v)
				case types.PermissionWriteAcp:
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
