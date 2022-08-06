package meta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func registerDataSourceTypeFactory(name string, factory func(context.Context) (tfsdk.DataSourceType, error)) {
}
