package s3

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/mitchellh/go-homedir"
)

const objectCreationTimeout = 2 * time.Minute

func ResourceObject() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceObjectCreate,
		ReadWithoutTimeout:   resourceObjectRead,
		UpdateWithoutTimeout: resourceObjectUpdate,
		DeleteWithoutTimeout: resourceObjectDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceObjectImport,
		},

		CustomizeDiff: customdiff.Sequence(
			resourceObjectCustomizeDiff,
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"acl": {
				Type:         schema.TypeString,
				Default:      s3.ObjectCannedACLPrivate,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(s3.ObjectCannedACL_Values(), false),
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
			},
			"content": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"source", "content_base64"},
			},
			"content_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"source", "content"},
			},
			"content_disposition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"content_encoding": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"content_language": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"content_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"etag": {
				Type: schema.TypeString,
				// This will conflict with SSE-C and SSE-KMS encryption and multi-part upload
				// if/when it's actually implemented. The Etag then won't match raw-file MD5.
				// See http://docs.aws.amazon.com/AmazonS3/latest/API/RESTCommonResponseHeaders.html
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"kms_key_id"},
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// ignore diffs where the user hasn't specified a kms_key_id but the bucket has a default KMS key configured
					if new == "" && d.Get("server_side_encryption") == s3.ServerSideEncryptionAwsKms {
						return true
					}
					return false
				},
			},
			"metadata": {
				Type:         schema.TypeMap,
				ValidateFunc: validateMetadataIsLowerCase,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
			},
			"object_lock_legal_hold_status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(s3.ObjectLockLegalHoldStatus_Values(), false),
			},
			"object_lock_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(s3.ObjectLockMode_Values(), false),
			},
			"object_lock_retain_until_date": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"server_side_encryption": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(s3.ServerSideEncryption_Values(), false),
				Computed:     true,
			},
			"source": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"content", "content_base64"},
			},
			"source_hash": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage_class": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(s3.ObjectStorageClass_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website_redirect": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceObjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return append(diags, resourceObjectUpload(ctx, d, meta)...)
}

func resourceObjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	var resp *s3.HeadObjectOutput

	err := resource.RetryContext(ctx, objectCreationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.HeadObjectWithContext(ctx, input)

		if d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.HeadObjectWithContext(ctx, input)
	}

	if !d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
		log.Printf("[WARN] S3 Object (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Object (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Reading S3 Object meta: %s", resp)

	d.Set("bucket_key_enabled", resp.BucketKeyEnabled)
	d.Set("cache_control", resp.CacheControl)
	d.Set("content_disposition", resp.ContentDisposition)
	d.Set("content_encoding", resp.ContentEncoding)
	d.Set("content_language", resp.ContentLanguage)
	d.Set("content_type", resp.ContentType)
	metadata := flex.PointersMapToStringList(resp.Metadata)

	// AWS Go SDK capitalizes metadata, this is a workaround. https://github.com/aws/aws-sdk-go/issues/445
	for k, v := range metadata {
		delete(metadata, k)
		metadata[strings.ToLower(k)] = v
	}

	if err := d.Set("metadata", metadata); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting metadata: %s", err)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("server_side_encryption", resp.ServerSideEncryption)
	d.Set("website_redirect", resp.WebsiteRedirectLocation)
	d.Set("object_lock_legal_hold_status", resp.ObjectLockLegalHoldStatus)
	d.Set("object_lock_mode", resp.ObjectLockMode)
	d.Set("object_lock_retain_until_date", flattenObjectDate(resp.ObjectLockRetainUntilDate))

	if err := resourceObjectSetKMS(ctx, d, meta, resp.SSEKMSKeyId); err != nil {
		return sdkdiag.AppendErrorf(diags, "object KMS: %s", err)
	}

	// See https://forums.aws.amazon.com/thread.jspa?threadID=44003
	d.Set("etag", strings.Trim(aws.StringValue(resp.ETag), `"`))

	// The "STANDARD" (which is also the default) storage
	// class when set would not be included in the results.
	if resp.StorageClass == nil {
		d.Set("storage_class", s3.StorageClassStandard)
	} else {
		d.Set("storage_class", resp.StorageClass)
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

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceObjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if hasObjectContentChanges(d) {
		return append(diags, resourceObjectUpload(ctx, d, meta)...)
	}

	conn := meta.(*conns.AWSClient).S3Conn()

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	if d.HasChange("acl") {
		_, err := conn.PutObjectAclWithContext(ctx, &s3.PutObjectAclInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			ACL:    aws.String(d.Get("acl").(string)),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting S3 object ACL: %s", err)
		}
	}

	if d.HasChange("object_lock_legal_hold_status") {
		_, err := conn.PutObjectLegalHoldWithContext(ctx, &s3.PutObjectLegalHoldInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			LegalHold: &s3.ObjectLockLegalHold{
				Status: aws.String(d.Get("object_lock_legal_hold_status").(string)),
			},
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting S3 object lock legal hold: %s", err)
		}
	}

	if d.HasChanges("object_lock_mode", "object_lock_retain_until_date") {
		req := &s3.PutObjectRetentionInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Retention: &s3.ObjectLockRetention{
				Mode:            aws.String(d.Get("object_lock_mode").(string)),
				RetainUntilDate: expandObjectDate(d.Get("object_lock_retain_until_date").(string)),
			},
		}

		// Bypass required to lower or clear retain-until date.
		if d.HasChange("object_lock_retain_until_date") {
			oraw, nraw := d.GetChange("object_lock_retain_until_date")
			o := expandObjectDate(oraw.(string))
			n := expandObjectDate(nraw.(string))
			if n == nil || (o != nil && n.Before(*o)) {
				req.BypassGovernanceRetention = aws.Bool(true)
			}
		}

		_, err := conn.PutObjectRetentionWithContext(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting S3 object lock retention: %s", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := ObjectUpdateTags(ctx, conn, bucket, key, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceObjectRead(ctx, d, meta)...)
}

func resourceObjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn()

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

func resourceObjectImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	id := d.Id()
	id = strings.TrimPrefix(id, "s3://")
	parts := strings.Split(id, "/")

	if len(parts) < 2 {
		return []*schema.ResourceData{d}, fmt.Errorf("id %s should be in format <bucket>/<key> or s3://<bucket>/<key>", id)
	}

	bucket := parts[0]
	key := strings.Join(parts[1:], "/")

	d.SetId(key)
	d.Set("bucket", bucket)
	d.Set("key", key)

	return []*schema.ResourceData{d}, nil
}

func resourceObjectUpload(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn()
	uploader := s3manager.NewUploaderWithClient(conn)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var body io.ReadSeeker

	if v, ok := d.GetOk("source"); ok {
		source := v.(string)
		path, err := homedir.Expand(source)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "expanding homedir in source (%s): %s", source, err)
		}
		file, err := os.Open(path)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "opening S3 object source (%s): %s", path, err)
		}

		body = file
		defer func() {
			err := file.Close()
			if err != nil {
				log.Printf("[WARN] Error closing S3 object source (%s): %s", path, err)
			}
		}()
	} else if v, ok := d.GetOk("content"); ok {
		content := v.(string)
		body = bytes.NewReader([]byte(content))
	} else if v, ok := d.GetOk("content_base64"); ok {
		content := v.(string)
		// We can't do streaming decoding here (with base64.NewDecoder) because
		// the AWS SDK requires an io.ReadSeeker but a base64 decoder can't seek.
		contentRaw, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "decoding content_base64: %s", err)
		}
		body = bytes.NewReader(contentRaw)
	} else {
		body = bytes.NewReader([]byte{})
	}

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	input := &s3manager.UploadInput{
		ACL:    aws.String(d.Get("acl").(string)),
		Body:   body,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if v, ok := d.GetOk("storage_class"); ok {
		input.StorageClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cache_control"); ok {
		input.CacheControl = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_type"); ok {
		input.ContentType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("metadata"); ok {
		input.Metadata = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("content_encoding"); ok {
		input.ContentEncoding = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_language"); ok {
		input.ContentLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_disposition"); ok {
		input.ContentDisposition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("bucket_key_enabled"); ok {
		input.BucketKeyEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("server_side_encryption"); ok {
		input.ServerSideEncryption = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.SSEKMSKeyId = aws.String(v.(string))
		input.ServerSideEncryption = aws.String(s3.ServerSideEncryptionAwsKms)
	}

	if len(tags) > 0 {
		// The tag-set must be encoded as URL Query parameters.
		input.Tagging = aws.String(tags.IgnoreAWS().URLEncode())
	}

	if v, ok := d.GetOk("website_redirect"); ok {
		input.WebsiteRedirectLocation = aws.String(v.(string))
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

	if _, err := uploader.Upload(input); err != nil {
		return sdkdiag.AppendErrorf(diags, "uploading object to S3 bucket (%s): %s", bucket, err)
	}

	d.SetId(key)

	return append(diags, resourceObjectRead(ctx, d, meta)...)
}

func resourceObjectSetKMS(ctx context.Context, d *schema.ResourceData, meta interface{}, sseKMSKeyId *string) error {
	// Only set non-default KMS key ID (one that doesn't match default)
	if sseKMSKeyId != nil {
		// retrieve S3 KMS Default Master Key
		conn := meta.(*conns.AWSClient).KMSConn()
		keyMetadata, err := kms.FindKeyByID(ctx, conn, DefaultKMSKeyAlias)
		if err != nil {
			return fmt.Errorf("Failed to describe default S3 KMS key (%s): %s", DefaultKMSKeyAlias, err)
		}

		if aws.StringValue(sseKMSKeyId) != aws.StringValue(keyMetadata.Arn) {
			log.Printf("[DEBUG] S3 object is encrypted using a non-default KMS Key ID: %s", aws.StringValue(sseKMSKeyId))
			d.Set("kms_key_id", sseKMSKeyId)
		}
	}

	return nil
}

func validateMetadataIsLowerCase(v interface{}, k string) (ws []string, errors []error) {
	value := v.(map[string]interface{})

	for k := range value {
		if k != strings.ToLower(k) {
			errors = append(errors, fmt.Errorf(
				"Metadata must be lowercase only. Offending key: %q", k))
		}
	}
	return
}

func resourceObjectCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	if hasObjectContentChanges(d) {
		return d.SetNewComputed("version_id")
	}

	if d.HasChange("source_hash") {
		d.SetNewComputed("version_id")
		d.SetNewComputed("etag")
	}

	return nil
}

func hasObjectContentChanges(d verify.ResourceDiffer) bool {
	for _, key := range []string{
		"bucket_key_enabled",
		"cache_control",
		"content_base64",
		"content_disposition",
		"content_encoding",
		"content_language",
		"content_type",
		"content",
		"etag",
		"kms_key_id",
		"metadata",
		"server_side_encryption",
		"source",
		"source_hash",
		"storage_class",
		"website_redirect",
	} {
		if d.HasChange(key) {
			return true
		}
	}
	return false
}

// DeleteAllObjectVersions deletes all versions of a specified key from an S3 bucket.
// If key is empty then all versions of all objects are deleted.
// Set force to true to override any S3 object lock protections on object lock enabled buckets.
// Returns the number of objects deleted.
func DeleteAllObjectVersions(ctx context.Context, conn *s3.S3, bucketName, key string, force, ignoreObjectErrors bool) (int64, error) {
	var nObjects int64

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	}
	if key != "" {
		input.Prefix = aws.String(key)
	}

	var lastErr error
	err := conn.ListObjectVersionsPagesWithContext(ctx, input, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, objectVersion := range page.Versions {
			objectKey := aws.StringValue(objectVersion.Key)
			objectVersionID := aws.StringValue(objectVersion.VersionId)

			if key != "" && key != objectKey {
				continue
			}

			err := deleteObjectVersion(ctx, conn, bucketName, objectKey, objectVersionID, force)

			if err == nil {
				nObjects++
			}

			if tfawserr.ErrCodeEquals(err, "AccessDenied") && force {
				// Remove any legal hold.
				resp, err := conn.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
					Bucket:    aws.String(bucketName),
					Key:       objectVersion.Key,
					VersionId: objectVersion.VersionId,
				})

				if err != nil {
					log.Printf("[ERROR] Error getting S3 Bucket (%s) Object (%s) Version (%s) metadata: %s", bucketName, objectKey, objectVersionID, err)
					lastErr = err
					continue
				}

				if aws.StringValue(resp.ObjectLockLegalHoldStatus) == s3.ObjectLockLegalHoldStatusOn {
					_, err := conn.PutObjectLegalHoldWithContext(ctx, &s3.PutObjectLegalHoldInput{
						Bucket:    aws.String(bucketName),
						Key:       objectVersion.Key,
						VersionId: objectVersion.VersionId,
						LegalHold: &s3.ObjectLockLegalHold{
							Status: aws.String(s3.ObjectLockLegalHoldStatusOff),
						},
					})

					if err != nil {
						log.Printf("[ERROR] Error putting S3 Bucket (%s) Object (%s) Version(%s) legal hold: %s", bucketName, objectKey, objectVersionID, err)
						lastErr = err
						continue
					}

					// Attempt to delete again.
					err = deleteObjectVersion(ctx, conn, bucketName, objectKey, objectVersionID, force)

					if err != nil {
						lastErr = err
					} else {
						nObjects++
					}

					continue
				}

				// AccessDenied for another reason.
				lastErr = fmt.Errorf("AccessDenied deleting S3 Bucket (%s) Object (%s) Version: %s", bucketName, objectKey, objectVersionID)
				continue
			}

			if err != nil {
				lastErr = err
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		err = nil
	}

	if err != nil {
		return nObjects, err
	}

	if lastErr != nil {
		if !ignoreObjectErrors {
			return nObjects, fmt.Errorf("deleting at least one object version, last error: %s", lastErr)
		}

		lastErr = nil
	}

	err = conn.ListObjectVersionsPagesWithContext(ctx, input, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deleteMarker := range page.DeleteMarkers {
			deleteMarkerKey := aws.StringValue(deleteMarker.Key)
			deleteMarkerVersionID := aws.StringValue(deleteMarker.VersionId)

			if key != "" && key != deleteMarkerKey {
				continue
			}

			// Delete markers have no object lock protections.
			err := deleteObjectVersion(ctx, conn, bucketName, deleteMarkerKey, deleteMarkerVersionID, false)

			if err != nil {
				lastErr = err
			} else {
				nObjects++
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		err = nil
	}

	if err != nil {
		return nObjects, err
	}

	if lastErr != nil {
		if !ignoreObjectErrors {
			return nObjects, fmt.Errorf("deleting at least one object delete marker, last error: %s", lastErr)
		}

		lastErr = nil
	}

	return nObjects, nil
}

// deleteObjectVersion deletes a specific object version.
// Set force to true to override any S3 object lock protections.
func deleteObjectVersion(ctx context.Context, conn *s3.S3, b, k, v string, force bool) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(k),
	}

	if v != "" {
		input.VersionId = aws.String(v)
	}

	if force {
		input.BypassGovernanceRetention = aws.Bool(true)
	}

	log.Printf("[INFO] Deleting S3 Bucket (%s) Object (%s) Version: %s", b, k, v)
	_, err := conn.DeleteObjectWithContext(ctx, input)

	if err != nil {
		log.Printf("[WARN] Error deleting S3 Bucket (%s) Object (%s) Version (%s): %s", b, k, v, err)
	}

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) || tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchKey) {
		return nil
	}

	return err
}

func expandObjectDate(v string) *time.Time {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}

	return aws.Time(t)
}

func flattenObjectDate(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.Format(time.RFC3339)
}
