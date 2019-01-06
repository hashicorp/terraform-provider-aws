package aws

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsS3BucketDirectory() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketDirectoryCreate,
		Read:   resourceAwsS3BucketDirectoryRead,
		Update: resourceAwsS3BucketDirectoryUpdate,
		Delete: resourceAwsS3BucketDirectoryDelete,

		CustomizeDiff: updateComputedBucketDirectoryAttributes,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"source": {
				Type:     schema.TypeString,
				Required: true,
			},

			"target": {
				Type:     schema.TypeString,
				Required: true,
			},

			"exclude": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"files": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target": {
							Type:     schema.TypeString,
							Required: true,
						},
						"etag": {
							Type:     schema.TypeString,
							Required: true,
						},
						"content_type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"etag": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func updateComputedBucketDirectoryAttributes(d *schema.ResourceDiff, meta interface{}) error {
	source := d.Get("source").(string)
	target := d.Get("target").(string)

	exclude := map[string]bool{}
	if v, ok := d.GetOk("exclude"); ok {
		list := v.([]interface{})
		for _, elem := range list {
			exclude[elem.(string)] = true
		}
	}

	resourceETag := md5.New()

	var files = []map[string]string{}
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && exclude[info.Name()] {
			return filepath.SkipDir
		}

		if !info.IsDir() && !exclude[info.Name()] {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("ReadFile error, %q", err.Error())
			}

			target := strings.Replace(path, source, target, 1)

			objectETag := md5.New()
			objectETag.Write(content)
			etag := hex.EncodeToString(objectETag.Sum(nil))
			resourceETag.Write([]byte(etag))

			parts := strings.Split(path, ".")
			contentType := mime.TypeByExtension("." + parts[len(parts)-1])
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			files = append(files, map[string]string{
				"source":       path,
				"target":       target,
				"etag":         etag,
				"content_type": contentType,
			})
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Walk error, %q", err.Error())
	}

	if v, ok := d.GetOk("etag"); !ok || v.(string) != hex.EncodeToString(resourceETag.Sum(nil)) {
		d.SetNewComputed("etag")
	}

	d.SetNew("files", files)
	return nil
}

func resourceAwsS3BucketDirectoryCreate(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn
	uploader := s3manager.NewUploaderWithClient(s3conn)

	bucket := d.Get("bucket").(string)
	target := d.Get("target").(string)

	files := make([]file, len(d.Get("files").(*schema.Set).List()))
	for idx, elem := range d.Get("files").(*schema.Set).List() {
		m := elem.(map[string]interface{})
		files[idx].source = m["source"].(string)
		files[idx].target = m["target"].(string)
		files[idx].etag = m["etag"].(string)
		files[idx].contentType = m["content_type"].(string)
	}

	iterator := s3manager.BatchUploadIterator(&directoryIterator{bucket: bucket, files: files})
	if err := uploader.UploadWithIterator(aws.BackgroundContext(), iterator); err != nil {
		return err
	}

	d.SetId(target)
	return resourceAwsS3BucketDirectoryRead(d, meta)
}

func resourceAwsS3BucketDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	source := d.Get("source").(string)
	target := d.Get("target").(string)

	input := s3.ListObjectsInput{Bucket: &bucket, Prefix: &target}
	resp, err := s3conn.ListObjects(&input)
	if err != nil {
		return fmt.Errorf("ListObjects error, %q", err.Error())
	}

	resourceETag := md5.New()
	var files = []map[string]string{}
	for _, obj := range resp.Contents {
		input := s3.GetObjectInput{Bucket: &bucket, Key: obj.Key}
		resp, err := s3conn.GetObject(&input)
		if err != nil {
			return fmt.Errorf("ListObjects error, %q", err.Error())
		}

		resourceETag.Write([]byte(strings.Trim(*obj.ETag, `"`)))

		files = append(files, map[string]string{
			"source":       strings.Replace(*obj.Key, target, source, 1),
			"target":       *obj.Key,
			"etag":         strings.Trim(*obj.ETag, `"`),
			"content_type": *resp.ContentType,
		})
	}

	d.Set("etag", hex.EncodeToString(resourceETag.Sum(nil)))
	d.Set("files", files)
	return nil
}

func resourceAwsS3BucketDirectoryUpdate(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn
	uploader := s3manager.NewUploaderWithClient(s3conn)
	deleter := s3manager.NewBatchDeleteWithClient(s3conn)

	bucket := d.Get("bucket").(string)

	filesToBe := map[string]file{}
	for _, elem := range d.Get("files").(*schema.Set).List() {
		m := elem.(map[string]interface{})
		f := file{
			source:      m["source"].(string),
			target:      m["target"].(string),
			etag:        m["etag"].(string),
			contentType: m["content_type"].(string),
		}
		filesToBe[f.target] = f
	}

	if err := resourceAwsS3BucketDirectoryRead(d, meta); err != nil {
		return err
	}

	filesToUpload := []file{}
	objectsToDelete := []s3manager.BatchDeleteObject{}
	for _, elem := range d.Get("files").(*schema.Set).List() {
		m := elem.(map[string]interface{})
		asIs := file{
			target:      m["target"].(string),
			etag:        m["etag"].(string),
			contentType: m["content_type"].(string),
		}

		if toBe, present := filesToBe[asIs.target]; present {
			if asIs.etag != toBe.etag || asIs.contentType != toBe.contentType {
				filesToUpload = append(filesToUpload, toBe)
			}
		} else {
			objectsToDelete = append(objectsToDelete,
				s3manager.BatchDeleteObject{
					Object: &s3.DeleteObjectInput{
						Key:    &asIs.target,
						Bucket: &bucket,
					},
				})
		}
	}

	iterator := s3manager.BatchUploadIterator(&directoryIterator{bucket: bucket, files: filesToUpload})
	if err := uploader.UploadWithIterator(aws.BackgroundContext(), iterator); err != nil {
		return err
	}

	if err := deleter.Delete(aws.BackgroundContext(), &s3manager.DeleteObjectsIterator{Objects: objectsToDelete}); err != nil {
		return err
	}

	return resourceAwsS3BucketDirectoryRead(d, meta)
}

func resourceAwsS3BucketDirectoryDelete(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn
	deleter := s3manager.NewBatchDeleteWithClient(s3conn)

	bucket := d.Get("bucket").(string)

	objectsToDelete := []s3manager.BatchDeleteObject{}
	for _, elem := range d.Get("files").(*schema.Set).List() {
		m := elem.(map[string]interface{})
		objectsToDelete = append(objectsToDelete,
			s3manager.BatchDeleteObject{
				Object: &s3.DeleteObjectInput{
					Key:    aws.String(m["target"].(string)),
					Bucket: &bucket,
				},
			})
	}

	if err := deleter.Delete(aws.BackgroundContext(), &s3manager.DeleteObjectsIterator{Objects: objectsToDelete}); err != nil {
		return err
	}
	return nil
}

type file struct {
	etag        string
	contentType string
	source      string
	target      string
	f           *os.File
}

type directoryIterator struct {
	bucket string
	files  []file
	next   file
	err    error
}

// Next opens the next file and stops iteration if it fails to open a file.
func (iter *directoryIterator) Next() bool {
	if len(iter.files) == 0 {
		iter.next.f = nil
		return false
	}

	f, err := os.Open(iter.files[0].source)
	iter.err = err

	iter.next = iter.files[0]
	iter.next.f = f

	iter.files = iter.files[1:]
	return true && iter.Err() == nil
}

// Err returns an error that was set during opening the file
func (iter *directoryIterator) Err() error {
	return iter.err
}

// UploadObject returns a BatchUploadObject and sets the After field to close the file.
func (iter *directoryIterator) UploadObject() s3manager.BatchUploadObject {
	f := iter.next.f
	return s3manager.BatchUploadObject{
		Object: &s3manager.UploadInput{
			Bucket:      &iter.bucket,
			Key:         &iter.next.target,
			ContentType: &iter.next.contentType,
			Body:        f,
		},

		After: func() error {
			return f.Close()
		},
	}
}
