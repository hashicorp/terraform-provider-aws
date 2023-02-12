package lambda

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// CustomizeDiffValidateInput validates that `input` is JSON dict when `lifecycle_scope is not "CREATE_ONLY"
func CustomizeDiffValidateInput(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Get("lifecycle_scope") == lambdaLifecycleScopeCreateOnly {
		return nil
	}
	// input is validated to be valid JSON in the schema already.
	inputNoSpaces := strings.TrimSpace(diff.Get("input").(string))
	if strings.HasPrefix(inputNoSpaces, "{") && strings.HasSuffix(inputNoSpaces, "}") {
		return nil
	}

	return errors.New(`lifecycle_scope other than "CREATE" require input to be a JSON object`)
}
