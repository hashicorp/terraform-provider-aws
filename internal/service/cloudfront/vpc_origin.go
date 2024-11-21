package cloudfront

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"time"
)

// @FrameworkResource("aws_cloudfront_vpc_origin", name="VPC Origin")
func newCloudfrontVPCOriginResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &cloudfrontVPCOriginResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type cloudfrontVPCOriginResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *cloudfrontVPCOriginResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_something_example"
}
