// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_s3_object")
func DataSourceObject() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceObjectRead,

		Schema: map[string]*schema.Schema{
			"body": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bucket_key_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cache_control": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.ChecksumMode](),
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
				Computed: true,
			},
			"content_encoding": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_language": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"content_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"object_lock_legal_hold_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"object_lock_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"object_lock_retain_until_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"range": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"server_side_encryption": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sse_kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"version_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"website_redirect_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceObjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	bucket := d.Get("bucket").(string)
	key := sdkv1CompatibleCleanKey(d.Get("key").(string))
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if v, ok := d.GetOk("checksum_mode"); ok {
		input.ChecksumMode = types.ChecksumMode(v.(string))
	}
	if v, ok := d.GetOk("range"); ok {
		input.Range = aws.String(v.(string))
	}
	if v, ok := d.GetOk("version_id"); ok {
		input.VersionId = aws.String(v.(string))
	}

	output, err := findObject(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) Object (%s): %s", bucket, key, err)
	}

	if aws.ToBool(output.DeleteMarker) {
		return sdkdiag.AppendErrorf(diags, "S3 Bucket (%s) Object (%s) has been deleted", bucket, key)
	}

	id := bucket + "/" + d.Get("key").(string)
	if v, ok := d.GetOk("version_id"); ok {
		id += "@" + v.(string)
	}
	d.SetId(id)

	d.Set("bucket_key_enabled", output.BucketKeyEnabled)
	d.Set("cache_control", output.CacheControl)
	d.Set("checksum_crc32", output.ChecksumCRC32)
	d.Set("checksum_crc32c", output.ChecksumCRC32C)
	d.Set("checksum_sha1", output.ChecksumSHA1)
	d.Set("checksum_sha256", output.ChecksumSHA256)
	d.Set("content_disposition", output.ContentDisposition)
	d.Set("content_encoding", output.ContentEncoding)
	d.Set("content_language", output.ContentLanguage)
	d.Set("content_length", output.ContentLength)
	d.Set("content_type", output.ContentType)
	// See https://forums.aws.amazon.com/thread.jspa?threadID=44003
	d.Set("etag", strings.Trim(aws.ToString(output.ETag), `"`))
	d.Set("expiration", output.Expiration)
	if output.Expires != nil {
		d.Set("expires", output.Expires.Format(time.RFC1123))
	} else {
		d.Set("expires", nil)
	}
	if output.LastModified != nil {
		d.Set("last_modified", output.LastModified.Format(time.RFC1123))
	} else {
		d.Set("last_modified", nil)
	}
	d.Set("metadata", output.Metadata)
	d.Set("object_lock_legal_hold_status", output.ObjectLockLegalHoldStatus)
	d.Set("object_lock_mode", output.ObjectLockMode)
	d.Set("object_lock_retain_until_date", flattenObjectDate(output.ObjectLockRetainUntilDate))
	d.Set("server_side_encryption", output.ServerSideEncryption)
	d.Set("sse_kms_key_id", output.SSEKMSKeyId)
	// The "STANDARD" (which is also the default) storage
	// class when set would not be included in the results.
	d.Set("storage_class", types.ObjectStorageClassStandard)
	if output.StorageClass != "" {
		d.Set("storage_class", output.StorageClass)
	}
	d.Set("version_id", output.VersionId)
	d.Set("website_redirect_location", output.WebsiteRedirectLocation)

	if isContentTypeAllowed(output.ContentType) {
		downloader := manager.NewDownloader(conn)
		buf := manager.NewWriteAtBuffer(make([]byte, 0))
		input := &s3.GetObjectInput{
			Bucket:    aws.String(bucket),
			Key:       aws.String(key),
			VersionId: output.VersionId,
		}
		if v, ok := d.GetOk("range"); ok {
			input.Range = aws.String(v.(string))
		}

		_, err := downloader.Download(ctx, buf, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "downloading S3 Bucket (%s) Object (%s): %s", bucket, key, err)
		}

		d.Set("body", string(buf.Bytes()))
	}

	if tags, err := ObjectListTags(ctx, conn, bucket, key); err == nil {
		if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}
	} else if !tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotImplemented) { // Directory buckets return HTTP status code 501, NotImplemented.
		return sdkdiag.AppendErrorf(diags, "listing tags for S3 Bucket (%s) Object (%s): %s", bucket, key, err)
	}

	return diags
}

// This is to prevent potential issues w/ binary files and generally unprintable characters.
// See https://github.com/hashicorp/terraform/pull/3858#issuecomment-156856738.
func isContentTypeAllowed(contentType *string) bool {
	if contentType == nil {
		return false
	}

	allowedContentTypes := []*regexp.Regexp{
		regexache.MustCompile(`^application/atom\+xml$`),
		regexache.MustCompile(`^application/json$`),
		regexache.MustCompile(`^application/ld\+json$`),
		regexache.MustCompile(`^application/x-csh$`),
		regexache.MustCompile(`^application/x-httpd-php$`),
		regexache.MustCompile(`^application/x-sh$`),
		regexache.MustCompile(`^application/xhtml\+xml$`),
		regexache.MustCompile(`^application/xml$`),
		regexache.MustCompile(`^text/.+`),
	}
	for _, r := range allowedContentTypes {
		if r.MatchString(aws.ToString(contentType)) {
			return true
		}
	}

	return false
}
