// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplestructfield

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplecrossservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplecrossservice/types"
)

// parentLookup has a conn field, like deploymentOrchestrator, but for a different service's client.
type parentLookup struct {
	conn *examplecrossservice.Client
}

func newParentLookup(conn *examplecrossservice.Client) *parentLookup {
	return &parentLookup{
		conn: conn,
	}
}

func (l *parentLookup) Find(ctx context.Context, id string) (*awstypes.Foo, error) {
	input := examplecrossservice.DescribeFooInput{
		FooId: aws.String(id),
	}

	output, err := l.conn.DescribeFoo(ctx, &input)

	if err != nil {
		return nil, err
	}

	return output.Foo, nil
}
