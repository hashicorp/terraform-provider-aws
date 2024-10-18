// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
    "context"
    "reflect"
    "strings"
    "testing"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
    "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestFindTransitGatewayAttachment(t *testing.T) {
    t.Parallel()

    testCases := []struct {
        name                string
        input               []awstypes.TransitGatewayAttachment
        expectedOutput      *awstypes.TransitGatewayAttachment
        expectedErrContains string
    }{
        {
            name: "Single available attachment",
            input: []awstypes.TransitGatewayAttachment{
                {
                    TransitGatewayAttachmentId: aws.String("tgw-attach-12345"),
                    State:                      awstypes.TransitGatewayAttachmentStateAvailable,
                },
            },
            expectedOutput: &awstypes.TransitGatewayAttachment{
                TransitGatewayAttachmentId: aws.String("tgw-attach-12345"),
                State:                      awstypes.TransitGatewayAttachmentStateAvailable,
            },
        },
        {
            name: "Multiple attachments, one available",
            input: []awstypes.TransitGatewayAttachment{
                {
                    TransitGatewayAttachmentId: aws.String("tgw-attach-12345"),
                    State:                      awstypes.TransitGatewayAttachmentStateAvailable,
                },
                {
                    TransitGatewayAttachmentId: aws.String("tgw-attach-67890"),
                    State:                      awstypes.TransitGatewayAttachmentStateDeleted,
                },
            },
            expectedOutput: &awstypes.TransitGatewayAttachment{
                TransitGatewayAttachmentId: aws.String("tgw-attach-12345"),
                State:                      awstypes.TransitGatewayAttachmentStateAvailable,
            },
        },
        {
            name: "No available attachments",
            input: []awstypes.TransitGatewayAttachment{
                {
                    TransitGatewayAttachmentId: aws.String("tgw-attach-12345"),
                    State:                      awstypes.TransitGatewayAttachmentStateDeleted,
                },
            },
            expectedErrContains: "no results found",
        },
        {
            name:                "Empty input",
            input:               []awstypes.TransitGatewayAttachment{},
            expectedErrContains: "no results found",
        },
        {
            name: "Multiple available attachments",
            input: []awstypes.TransitGatewayAttachment{
                {
                    TransitGatewayAttachmentId: aws.String("tgw-attach-12345"),
                    State:                      awstypes.TransitGatewayAttachmentStateAvailable,
                },
                {
                    TransitGatewayAttachmentId: aws.String("tgw-attach-67890"),
                    State:                      awstypes.TransitGatewayAttachmentStateAvailable,
                },
            },
            expectedErrContains: "multiple results found",
        },
    }

    for _, tc := range testCases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            ctx := context.Background()
            conn := &mockEC2Client{}
            input := &ec2.DescribeTransitGatewayAttachmentsInput{}

            // Mock the findTransitGatewayAttachments function
            findTransitGatewayAttachments = func(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayAttachmentsInput) ([]awstypes.TransitGatewayAttachment, error) {
                return tc.input, nil
            }

            got, err := ec2.findTransitGatewayAttachment(ctx, conn, input)

            if tc.expectedErrContains != "" {
                if err == nil {
                    t.Fatalf("expected error containing %q, got no error", tc.expectedErrContains)
                }
                if !strings.Contains(err.Error(), tc.expectedErrContains) {
                    t.Fatalf("expected error containing %q, got %v", tc.expectedErrContains, err)
                }
            } else {
                if err != nil {
                    t.Fatalf("unexpected error: %v", err)
                }
                if !reflect.DeepEqual(got, tc.expectedOutput) {
                    t.Fatalf("got %v, want %v", got, tc.expectedOutput)
                }
            }
        })
    }
}
