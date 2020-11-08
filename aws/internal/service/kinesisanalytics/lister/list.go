package lister

import (
	//"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
)

func ListApplicationsPages(conn *kinesisanalytics.KinesisAnalytics, input *kinesisanalytics.ListApplicationsInput, fn func(*kinesisanalytics.ListApplicationsOutput, bool) bool) error {
	return nil
}
