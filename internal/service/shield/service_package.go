package shield

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	shield_sdkv1 "github.com/aws/aws-sdk-go/service/shield"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context) (*shield_sdkv1.Shield, error) {
	sess := p.config["session"].(*session_sdkv1.Session)
	config := &aws_sdkv1.Config{Endpoint: aws_sdkv1.String(p.config["endpoint"].(string))}

	// Force "global" services to correct Regions.
	if p.config["partition"].(string) == endpoints_sdkv1.AwsPartitionID {
		config.Region = aws.String(endpoints_sdkv1.UsEast1RegionID)
	}

	return shield_sdkv1.New(sess.Copy(config)), nil
}
