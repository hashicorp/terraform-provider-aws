package meta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/intf"
)

// TODO: This can all be generated.

var sd = &serviceData{}

func registerDataSourceFactory(factory func(context.Context) (datasource.DataSource, error)) {
	sd.dataSourceFactories = append(sd.dataSourceFactories, factory)
}

var ServiceData intf.ServiceData = sd

type serviceData struct {
	dataSourceFactories []func(context.Context) (datasource.DataSource, error)
}

func (d *serviceData) Configure(ctx context.Context, providerData intf.ProviderData) error {
	return nil
}

func (d *serviceData) DataSources(ctx context.Context) []func(context.Context) (datasource.DataSource, error) {
	return d.dataSourceFactories
}
