package aws

import (
	"testing"
)

func TestAccAWSAppStreamResource_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"ImageBuilder": {
			"basic":          testAccAwsAppStreamImageBuilder_basic,
			"name_generated": testAccAwsAppStreamImageBuilder_Name_Generated,
			"name_prefix":    testAccAwsAppStreamImageBuilder_NamePrefix,
			"complete":       testAccAwsAppStreamImageBuilder_Complete,
			"tags":           testAccAwsAppStreamImageBuilder_withTags,
			"disappears":     testAccAwsAppStreamImageBuilder_disappears,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
