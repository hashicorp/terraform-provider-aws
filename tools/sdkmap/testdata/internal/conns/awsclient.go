// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplebuildconstraints"
	"github.com/aws/aws-sdk-go-v2/service/examplecrossservice"
	"github.com/aws/aws-sdk-go-v2/service/examplecrud"
	"github.com/aws/aws-sdk-go-v2/service/examplecustomizediff"
	"github.com/aws/aws-sdk-go-v2/service/examplefuncrefs"
	"github.com/aws/aws-sdk-go-v2/service/exampleimport"
	"github.com/aws/aws-sdk-go-v2/service/exampleinterface"
	"github.com/aws/aws-sdk-go-v2/service/examplestructfield"
	"github.com/aws/aws-sdk-go-v2/service/exampletags"
	"github.com/aws/aws-sdk-go-v2/service/exampletagsgen"
	"github.com/aws/aws-sdk-go-v2/service/exampletagsupdateonly"
	"github.com/aws/aws-sdk-go-v2/service/exampleunanalyzed"
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

func (c *AWSClient) ExampleBuildConstraintsClient(ctx context.Context) *examplebuildconstraints.Client {
	return client[*examplebuildconstraints.Client](ctx, c, "examplebuildconstraints")
}

func (c *AWSClient) ExampleCrossServiceClient(ctx context.Context) *examplecrossservice.Client {
	return client[*examplecrossservice.Client](ctx, c, "examplecrossservice")
}

func (c *AWSClient) ExampleCRUDClient(ctx context.Context) *examplecrud.Client {
	return client[*examplecrud.Client](ctx, c, "examplecrud")
}

func (c *AWSClient) ExampleCustomizeDiffClient(ctx context.Context) *examplecustomizediff.Client {
	return client[*examplecustomizediff.Client](ctx, c, "examplecustomizediff")
}

func (c *AWSClient) ExampleFuncRefsClient(ctx context.Context) *examplefuncrefs.Client {
	return client[*examplefuncrefs.Client](ctx, c, "examplefuncrefs")
}

func (c *AWSClient) ExampleImportClient(ctx context.Context) *exampleimport.Client {
	return client[*exampleimport.Client](ctx, c, "exampleimport")
}

func (c *AWSClient) ExampleInterfaceClient(ctx context.Context) *exampleinterface.Client {
	return client[*exampleinterface.Client](ctx, c, "exampleinterface")
}

func (c *AWSClient) ExampleStructFieldClient(ctx context.Context) *examplestructfield.Client {
	return client[*examplestructfield.Client](ctx, c, "examplestructfield")
}

func (c *AWSClient) ExampleTagsClient(ctx context.Context) *exampletags.Client {
	return client[*exampletags.Client](ctx, c, "exampletags")
}

func (c *AWSClient) ExampleTagsGenClient(ctx context.Context) *exampletagsgen.Client {
	return client[*exampletagsgen.Client](ctx, c, "exampletagsgen")
}

func (c *AWSClient) ExampleTagsUpdateOnlyClient(ctx context.Context) *exampletagsupdateonly.Client {
	return client[*exampletagsupdateonly.Client](ctx, c, "exampletagsupdateonly")
}

func (c *AWSClient) ExampleUnanalyzedClient(ctx context.Context) *exampleunanalyzed.Client {
	return client[*exampleunanalyzed.Client](ctx, c, "exampleunanalyzed")
}

func client[T any](_ context.Context, c *AWSClient, servicePackageName string) T {
	return c.clients[servicePackageName].(T)
}
