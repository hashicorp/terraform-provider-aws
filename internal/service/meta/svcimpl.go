package meta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/intf"
)

// TODO: This can all be generated.

var sd = &serviceData{}

func registerFWDataSourceFactory(factory func(context.Context) (datasource.DataSource, error)) {
	sd.fwDataSourceFactories = append(sd.fwDataSourceFactories, factory)
}

var ServiceData intf.ServiceData = sd

type serviceData struct {
	fwDataSourceFactories []func(context.Context) (datasource.DataSource, error)
}

func (d *serviceData) Configure(ctx context.Context, meta any) error {
	return nil
}

func (d *serviceData) FrameworkDataSources(ctx context.Context) []func(context.Context) (datasource.DataSource, error) {
	return d.fwDataSourceFactories
}
