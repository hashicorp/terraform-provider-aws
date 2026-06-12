// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package testparser

import (
	"os"
	"slices"
	"testing"
)

func TestParseBasicTest_Check(t *testing.T) {
	src := `package example_test

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccExampleThing_basic(t *testing.T) {
	resourceName := "aws_example_thing.test"

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrARN, "arn:aws:example:us-east-1:123456789012:thing/test"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}
`
	tmpFile := t.TempDir() + "/test.go"
	if err := os.WriteFile(tmpFile, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	attrs, err := ParseBasicTest(tmpFile, "TestAccExampleThing_basic", "resourceName")
	if err != nil {
		t.Fatal(err)
	}

	paths := make([]string, len(attrs))
	for i, a := range attrs {
		paths[i] = a.Path
	}

	for _, want := range []string{"name", "arn", "id"} {
		if !slices.Contains(paths, want) {
			t.Errorf("missing expected checked attribute: %s (got %v)", want, paths)
		}
	}
}

func TestParseBasicTest_ConfigStateChecks(t *testing.T) {
	src := `package example_test

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

func TestAccExampleThing_basic(t *testing.T) {
	resourceName := "aws_example_thing.test"

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExists(resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("name"), knownvalue.StringExact("test")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("config"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"key":   knownvalue.StringExact("k"),
							"value": knownvalue.StringExact("v"),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("settings"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"enabled": knownvalue.Bool(true),
					})),
				},
			},
		},
	})
}
`
	tmpFile := t.TempDir() + "/test.go"
	if err := os.WriteFile(tmpFile, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	attrs, err := ParseBasicTest(tmpFile, "TestAccExampleThing_basic", "resourceName")
	if err != nil {
		t.Fatal(err)
	}

	paths := make([]string, len(attrs))
	for i, a := range attrs {
		paths[i] = a.Path
	}

	want := []string{"name", "config", "config.key", "config.value", "settings", "settings.enabled"}
	for _, w := range want {
		if !slices.Contains(paths, w) {
			t.Errorf("missing expected checked attribute: %s (got %v)", w, paths)
		}
	}
}

func TestParseBasicTest_MixedCheckAndConfigStateChecks(t *testing.T) {
	src := `package example_test

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

func TestAccExampleThing_basic(t *testing.T) {
	resourceName := "aws_example_thing.test"

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("description"), knownvalue.StringExact("")),
				},
			},
		},
	})
}
`
	tmpFile := t.TempDir() + "/test.go"
	if err := os.WriteFile(tmpFile, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	attrs, err := ParseBasicTest(tmpFile, "TestAccExampleThing_basic", "resourceName")
	if err != nil {
		t.Fatal(err)
	}

	var checkPaths, statePaths []string
	for _, a := range attrs {
		switch a.Source {
		case "Check":
			checkPaths = append(checkPaths, a.Path)
		case "ConfigStateChecks":
			statePaths = append(statePaths, a.Path)
		}
	}

	if !slices.Contains(checkPaths, "name") {
		t.Errorf("missing 'name' in Check paths: %v", checkPaths)
	}
	if !slices.Contains(statePaths, "description") {
		t.Errorf("missing 'description' in ConfigStateChecks paths: %v", statePaths)
	}
}
