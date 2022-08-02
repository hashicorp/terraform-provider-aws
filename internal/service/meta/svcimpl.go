package meta

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-provider-aws/internal/intf"
)

// TODO: This can all be generated.

var sd = &serviceData{}

func registerDataSourceTypeFactory(name string, factory func(context.Context) (tfsdk.DataSourceType, error)) {
	sd.dataSourceTypeFactories = append(sd.dataSourceTypeFactories, struct {
		name    string
		factory func(context.Context) (tfsdk.DataSourceType, error)
	}{
		name:    name,
		factory: factory,
	})
}

var ServiceData intf.ServiceData = sd

type serviceData struct {
	dataSourceTypeFactories []struct {
		name    string
		factory func(context.Context) (tfsdk.DataSourceType, error)
	}
}

func (d *serviceData) DataSourceTypes(ctx context.Context) (map[string]tfsdk.DataSourceType, error) {
	dataSourceTypes := make(map[string]tfsdk.DataSourceType)

	for _, dataSourceTypeFactory := range d.dataSourceTypeFactories {
		name := dataSourceTypeFactory.name

		if _, ok := dataSourceTypes[name]; ok {
			return nil, fmt.Errorf("duplicate data source (%s)", name)
		} else {
			dataSourceType, err := dataSourceTypeFactory.factory(ctx)

			if err != nil {
				return nil, err
			}

			dataSourceTypes[name] = dataSourceType
		}
	}

	return dataSourceTypes, nil
}
