// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockstructfield

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockcrossservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockcrossservice/types"
)

// parentLookup has a conn field, like deploymentOrchestrator, but for a different service's client.
type parentLookup struct {
	conn *mockcrossservice.Client
}

func newParentLookup(conn *mockcrossservice.Client) *parentLookup {
	return &parentLookup{
		conn: conn,
	}
}

func (l *parentLookup) Find(ctx context.Context, id string) (*awstypes.Foo, error) {
	input := mockcrossservice.DescribeFooInput{
		FooId: aws.String(id),
	}

	output, err := l.conn.DescribeFoo(ctx, &input)

	if err != nil {
		return nil, err
	}

	return output.Foo, nil
}
