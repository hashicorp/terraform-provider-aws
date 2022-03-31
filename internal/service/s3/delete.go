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
	objects   []*s3.ObjectVersion
}

func NewDeleteVersionListIterator(conn *s3.S3, bucket, key string) s3manager.BatchDeleteIterator {
	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
	}

	if key != "" {
		input.Prefix = aws.String(key)
	}

	return &deleteVersionListIterator{
		Bucket: input.Bucket,
		Paginator: request.Pagination{
			NewRequest: func() (*request.Request, error) {
				inputCopy := *input
				request, _ := conn.ListObjectVersionsRequest(&inputCopy)

				return request, nil
			},
		},
	}
}

func (iter *deleteVersionListIterator) Next() bool {
	if len(iter.objects) > 0 {
		iter.objects = iter.objects[1:]
	}

	if len(iter.objects) == 0 && iter.Paginator.Next() {
		iter.objects = iter.Paginator.Page().(*s3.ListObjectVersionsOutput).Versions
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
