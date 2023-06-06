package logging

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

const (
	// SubsystemHelperResource is the tfsdklog subsystem name for helper/resource.
	SubsystemHelperResource = "helper_resource"
)

// HelperResourceTrace emits a helper/resource subsystem log at TRACE level.
func HelperResourceTrace(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	tfsdklog.SubsystemTrace(ctx, SubsystemHelperResource, msg, additionalFields...)
}

// HelperResourceDebug emits a helper/resource subsystem log at DEBUG level.
func HelperResourceDebug(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	tfsdklog.SubsystemDebug(ctx, SubsystemHelperResource, msg, additionalFields...)
}

// HelperResourceWarn emits a helper/resource subsystem log at WARN level.
func HelperResourceWarn(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	tfsdklog.SubsystemWarn(ctx, SubsystemHelperResource, msg, additionalFields...)
}

// HelperResourceError emits a helper/resource subsystem log at ERROR level.
func HelperResourceError(ctx context.Context, msg string, additionalFields ...map[string]interface{}) {
	tfsdklog.SubsystemError(ctx, SubsystemHelperResource, msg, additionalFields...)
}
