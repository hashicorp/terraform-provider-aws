package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

//
// Test harness for the EmptyBucket function.
// To run: AWS_REGION=us-west-2 go run main.go -b test-bucket
//

func main() {
	var bucket string
	var force bool

	flag.StringVar(&bucket, "b", "", "bucket")
	flag.BoolVar(&force, "f", false, "force")
	flag.Parse()

	if bucket == "" {
		fmt.Fprintf(os.Stderr, "bucket not specified\n")
		return
	}

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)
	ctx := context.Background()

	fmt.Printf("emptying %s...\n", bucket)

	err := tfs3.EmptyBucket(ctx, svc, bucket, force)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	} else {
		fmt.Printf("done!\n")
	}
}
