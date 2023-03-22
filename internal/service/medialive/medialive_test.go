package medialive_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMediaLive_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Multiplex": {
			"basic":      testAccMultiplex_basic,
			"disappears": testAccMultiplex_disappears,
			"update":     testAccMultiplex_update,
			"updateTags": testAccMultiplex_updateTags,
			"start":      testAccMultiplex_start,
		},
		"MultiplexProgram": {
			"basic":      testAccMultiplexProgram_basic,
			"update":     testAccMultiplexProgram_update,
			"disappears": testAccMultiplexProgram_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
