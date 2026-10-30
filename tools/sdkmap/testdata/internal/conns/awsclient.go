// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockbuildconstraints"
	"github.com/aws/aws-sdk-go-v2/service/mockcrossservice"
	"github.com/aws/aws-sdk-go-v2/service/mockcrud"
	"github.com/aws/aws-sdk-go-v2/service/mockcustomizediff"
	"github.com/aws/aws-sdk-go-v2/service/mockfuncrefs"
	"github.com/aws/aws-sdk-go-v2/service/mockimport"
	"github.com/aws/aws-sdk-go-v2/service/mockinterface"
	"github.com/aws/aws-sdk-go-v2/service/mockstructfield"
	"github.com/aws/aws-sdk-go-v2/service/mocktags"
	"github.com/aws/aws-sdk-go-v2/service/mocktagsgen"
	"github.com/aws/aws-sdk-go-v2/service/mocktagsnohandlers"
	"github.com/aws/aws-sdk-go-v2/service/mocktagsupdateonly"
	"github.com/aws/aws-sdk-go-v2/service/mockunanalyzed"
)

type AWSClient struct {
	accountID string
	awsConfig *aws.Config
	clients   map[string]any
	partition string
}

func (c *AWSClient) AccountID(context.Context) string {
	return c.accountID
}

func (c *AWSClient) AwsConfig(context.Context) aws.Config {
	return c.awsConfig.Copy()
}

func (c *AWSClient) Partition(context.Context) string {
	return c.partition
}

func (c *AWSClient) MockBuildConstraintsClient(ctx context.Context) *mockbuildconstraints.Client {
	return client[*mockbuildconstraints.Client](ctx, c, "mockbuildconstraints")
}

func (c *AWSClient) MockCrossServiceClient(ctx context.Context) *mockcrossservice.Client {
	return client[*mockcrossservice.Client](ctx, c, "mockcrossservice")
}

func (c *AWSClient) MockCRUDClient(ctx context.Context) *mockcrud.Client {
	return client[*mockcrud.Client](ctx, c, "mockcrud")
}

func (c *AWSClient) MockCustomizeDiffClient(ctx context.Context) *mockcustomizediff.Client {
	return client[*mockcustomizediff.Client](ctx, c, "mockcustomizediff")
}

func (c *AWSClient) MockFuncRefsClient(ctx context.Context) *mockfuncrefs.Client {
	return client[*mockfuncrefs.Client](ctx, c, "mockfuncrefs")
}

func (c *AWSClient) MockImportClient(ctx context.Context) *mockimport.Client {
	return client[*mockimport.Client](ctx, c, "mockimport")
}

func (c *AWSClient) MockInterfaceClient(ctx context.Context) *mockinterface.Client {
	return client[*mockinterface.Client](ctx, c, "mockinterface")
}

func (c *AWSClient) MockStructFieldClient(ctx context.Context) *mockstructfield.Client {
	return client[*mockstructfield.Client](ctx, c, "mockstructfield")
}

func (c *AWSClient) MockTagsClient(ctx context.Context) *mocktags.Client {
	return client[*mocktags.Client](ctx, c, "mocktags")
}

func (c *AWSClient) MockTagsGenClient(ctx context.Context) *mocktagsgen.Client {
	return client[*mocktagsgen.Client](ctx, c, "mocktagsgen")
}

func (c *AWSClient) MockTagsNoHandlersClient(ctx context.Context) *mocktagsnohandlers.Client {
	return client[*mocktagsnohandlers.Client](ctx, c, "mocktagsnohandlers")
}

func (c *AWSClient) MockTagsUpdateOnlyClient(ctx context.Context) *mocktagsupdateonly.Client {
	return client[*mocktagsupdateonly.Client](ctx, c, "mocktagsupdateonly")
}

func (c *AWSClient) MockUnanalyzedClient(ctx context.Context) *mockunanalyzed.Client {
	return client[*mockunanalyzed.Client](ctx, c, "mockunanalyzed")
}

func client[T any](_ context.Context, c *AWSClient, servicePackageName string) T {
	return c.clients[servicePackageName].(T)
}
