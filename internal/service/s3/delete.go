package s3

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	deleteBatchSize = 500
)

type objectVersionDeleter struct {
	batchDelete         *s3manager.BatchDelete
	batchDeleteIterator s3manager.BatchDeleteIterator
}

func NewObjectVersionDeleter(conn *s3.S3, bucket, key string) *objectVersionDeleter {
	return &objectVersionDeleter{
		batchDelete:         s3manager.NewBatchDeleteWithClient(conn, func(o *s3manager.BatchDelete) { o.BatchSize = deleteBatchSize }),
		batchDeleteIterator: NewDeleteVersionListIterator(conn, bucket, key),
	}
}

func (self *objectVersionDeleter) DeleteAll(ctx context.Context) error {
	return self.batchDelete.Delete(ctx, self.batchDeleteIterator)
}

// deleteVersionListIterator implements s3manager.BatchDeleteIterator.
// It is inspired by s3manager.DeleteListIterator.
type deleteVersionListIterator struct {
	Bucket    *string
	Paginator request.Pagination
	key       string
	objects   []*s3.ObjectVersion
}

func NewDeleteVersionListIterator(conn *s3.S3, bucket, key string) s3manager.BatchDeleteIterator {
	return &deleteVersionListIterator{
		Bucket: aws.String(bucket),
		Paginator: request.Pagination{
			NewRequest: func() (*request.Request, error) {
				input := &s3.ListObjectVersionsInput{
					Bucket: aws.String(bucket),
				}

				if key != "" {
					input.Prefix = aws.String(key)
				}

				request, _ := conn.ListObjectVersionsRequest(input)

				return request, nil
			},
		},
		key: key,
	}
}

func (iter *deleteVersionListIterator) Next() bool {
	if len(iter.objects) > 0 {
		iter.objects = iter.objects[1:]
	}

	if len(iter.objects) == 0 && iter.Paginator.Next() {
		if iter.key == "" {
			iter.objects = iter.Paginator.Page().(*s3.ListObjectVersionsOutput).Versions
		} else {
			// ListObjectVersions uses Prefix as an argument but we use Key.
			// Ignore any object versions that do not have the required Key.
			for _, v := range iter.Paginator.Page().(*s3.ListObjectVersionsOutput).Versions {
				if iter.key != aws.StringValue(v.Key) {
					continue
				}

				iter.objects = append(iter.objects, v)
			}
		}
	}

	return len(iter.objects) > 0
}

func (iter *deleteVersionListIterator) Err() error {
	return iter.Paginator.Err()
}

func (iter *deleteVersionListIterator) DeleteObject() s3manager.BatchDeleteObject {
	return s3manager.BatchDeleteObject{
		Object: &s3.DeleteObjectInput{
			Bucket:    iter.Bucket,
			Key:       iter.objects[0].Key,
			VersionId: iter.objects[0].VersionId,
		},
	}
}
