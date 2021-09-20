package applicationautoscaling

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// suppressEquivalentTime suppresses differences for time values that represent the same
// instant in different timezones.
func suppressEquivalentTime(k, old, new string, d *schema.ResourceData) bool {
	oldTime, err := time.Parse(time.RFC3339, old)
	if err != nil {
		return false
	}

	newTime, err := time.Parse(time.RFC3339, new)
	if err != nil {
		return false
	}

	return oldTime.Equal(newTime)
}
