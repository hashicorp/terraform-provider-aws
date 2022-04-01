package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testStepTaint(ctx context.Context, state *terraform.State, step TestStep) error {
	if len(step.Taint) == 0 {
		return nil
	}

	logging.HelperResourceTrace(ctx, fmt.Sprintf("Using TestStep Taint: %v", step.Taint))

	for _, p := range step.Taint {
		m := state.RootModule()
		if m == nil {
			return errors.New("no state")
		}
		rs, ok := m.Resources[p]
		if !ok {
			return fmt.Errorf("resource %q not found in state", p)
		}
		logging.HelperResourceWarn(ctx, fmt.Sprintf("Explicitly tainting resource %q", p))
		rs.Taint()
	}
	return nil
}
