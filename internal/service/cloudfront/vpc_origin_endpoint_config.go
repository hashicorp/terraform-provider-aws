package cloudfront

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"time"
)

// @FrameworkResource("aws_cloudfront_vpc_origin_endpoint_config", name="VPC Origin Endpoint Config")
func newCloudfrontVPCOriginEndpointConfigResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &cloudfrontVPCOriginEndpointConfigResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type cloudfrontVPCOriginEndpointConfigResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cloudfront_vpc_origin_endpoint_config"
}
