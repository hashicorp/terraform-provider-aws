package s3

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	s3_sdkv1 "github.com/aws/aws-sdk-go/service/s3"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context) (*s3_sdkv1.S3, error) {
	sess := p.config["session"].(*session_sdkv1.Session)
	config := &aws_sdkv1.Config{
		Endpoint:         aws_sdkv1.String(p.config["endpoint"].(string)),
		S3ForcePathStyle: aws_sdkv1.Bool(p.config["s3_use_path_style"].(bool)),
	}

	return s3_sdkv1.New(sess.Copy(config)), nil
}
