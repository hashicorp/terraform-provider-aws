package fwresource

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// @FrameworkResource("aws_test_widget", name="Test Widget")
func newTestWidgetResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &testWidgetResource{}
	return r, nil
}

type testWidgetResource struct{}

func (r *testWidgetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var conn *dynamodb.Client
	conn.CreateTable(ctx, &dynamodb.CreateTableInput{})
	conn.TagResource(ctx, &dynamodb.TagResourceInput{})
}

func (r *testWidgetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var conn *dynamodb.Client
	findTestWidget(ctx, conn, "id")
}

func (r *testWidgetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var conn *dynamodb.Client
	conn.UpdateTable(ctx, &dynamodb.UpdateTableInput{})
}

func (r *testWidgetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var conn *dynamodb.Client
	conn.DeleteTable(ctx, &dynamodb.DeleteTableInput{})
}

func findTestWidget(ctx context.Context, conn *dynamodb.Client, id string) (*dynamodb.DescribeTableOutput, error) {
	return conn.DescribeTable(ctx, &dynamodb.DescribeTableInput{})
}
