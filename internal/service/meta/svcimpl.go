package meta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func registerDataSourceTypeFactory(name string, factory func(context.Context) (provider.DataSourceType, error)) {
}
