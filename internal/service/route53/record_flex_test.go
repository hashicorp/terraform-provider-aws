// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"slices"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func TestFlattenResourceRecords(t *testing.T) {
	t.Parallel()

	original := []string{
		`127.0.0.1`,
		`"abc def"`,
		`"abc" "def"`,
		`"abc" ""`,
	}

	dequoted := []string{
		`127.0.0.1`,
		`abc def`,
		`abc" "def`,
		`abc" "`,
	}

	var wrapped []awstypes.ResourceRecord
	for _, original := range original {
		wrapped = append(wrapped, awstypes.ResourceRecord{Value: aws.String(original)})
	}

	sub := func(recordType awstypes.RRType, expected []string) {
		t.Run(string(recordType), func(t *testing.T) {
			checkFlattenResourceRecords(t, recordType, wrapped, expected)
		})
	}

	// These record types should be dequoted.
	sub(awstypes.RRTypeTxt, dequoted)
	sub(awstypes.RRTypeSpf, dequoted)

	// These record types should not be touched.
	sub(awstypes.RRTypeCname, original)
	sub(awstypes.RRTypeMx, original)
}

func checkFlattenResourceRecords(t *testing.T, recordType awstypes.RRType, expanded []awstypes.ResourceRecord, expected []string) {
	result := flattenResourceRecords(expanded, recordType)

	if result == nil {
		t.Fatal("expected result to have value, but got nil")
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}

	for i, e := range expected {
		if result[i] != e {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	}
}

func TestTxtCharacterStringLengths(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", 255)

	cases := []struct {
		Input    string
		Expected []int
	}{
		{``, []int{0}},
		{`abc`, []int{3}},
		{`abc def`, []int{7}},
		{`abc" "def`, []int{3, 3}},
		{`abc""defg`, []int{3, 4}},
		{`abc"  "de" "f`, []int{3, 2, 1}},
		{`abc" "`, []int{3, 0}},
		{`a\"b`, []int{3}},
		{`a\\" "b`, []int{2, 1}},
		{`a\065b`, []int{3}},
		{`a\06`, []int{3}},
		{long, []int{255}},
		{long + `a`, []int{256}},
		{long + `""` + long, []int{255, 255}},
		{long + `" "` + long + `a`, []int{255, 256}},
	}

	for _, tc := range cases {
		t.Run(tc.Input, func(t *testing.T) {
			t.Parallel()

			if got := txtCharacterStringLengths(tc.Input); !slices.Equal(got, tc.Expected) {
				t.Errorf("input: %s\nexpected: %v\ngot: %v", tc.Input, tc.Expected, got)
			}
		})
	}
}

func TestExpandRecordName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input, Output string
	}{
		{"www", "www.example.com"},
		{"www.", "www.example.com"},
		{"dev.www", "dev.www.example.com"},
		{"*", "\\052.example.com"},
		{"example.com", "example.com"},
		{"test.example.com", "test.example.com"},
		{"test.example.com.", "test.example.com"},
	}

	zone_name := "example.com"
	for _, tc := range cases {
		actual := expandRecordName(tc.Input, zone_name)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
		}
	}
}

func TestParseRecordID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input, Zone, Name, Type, Set string
	}{
		{"ABCDEF", "", "", "", ""},
		{"ABCDEF_test.example.com", "ABCDEF", "", "", ""},
		{"ABCDEF_test.example.com_A", "ABCDEF", "test.example.com", "A", ""},
		{"ABCDEF_test.example.com._A", "ABCDEF", "test.example.com", "A", ""},
		{"ABCDEF_test.example.com_A_set1", "ABCDEF", "test.example.com", "A", "set1"},
		{"ABCDEF__underscore.example.com_A", "ABCDEF", "_underscore.example.com", "A", ""},
		{"ABCDEF__underscore.example.com_A_set1", "ABCDEF", "_underscore.example.com", "A", "set1"},
		{"ABCDEF__underscore.example.com_A_set_with1", "ABCDEF", "_underscore.example.com", "A", "set_with1"},
		{"ABCDEF__underscore.example.com_A_set_with_1", "ABCDEF", "_underscore.example.com", "A", "set_with_1"},
		{"ABCDEF_prefix._underscore.example.com_A", "ABCDEF", "prefix._underscore.example.com", "A", ""},
		{"ABCDEF_prefix._underscore.example.com_A_set", "ABCDEF", "prefix._underscore.example.com", "A", "set"},
		{"ABCDEF_prefix._underscore.example.com_A_set_underscore", "ABCDEF", "prefix._underscore.example.com", "A", "set_underscore"},
		{"ABCDEF__A", "ABCDEF", "", "A", ""},
	}

	for _, tc := range cases {
		t.Run(tc.Input, func(t *testing.T) {
			t.Parallel()

			parts := recordParseResourceID(tc.Input)
			if parts[0] != tc.Zone {
				t.Fatalf("input: %s\nzone: %s\nexpected:%s", tc.Input, parts[0], tc.Zone)
			}
			if parts[1] != tc.Name {
				t.Fatalf("input: %s\nname: %s\nexpected:%s", tc.Input, parts[1], tc.Name)
			}
			if parts[2] != tc.Type {
				t.Fatalf("input: %s\ntype: %s\nexpected:%s", tc.Input, parts[2], tc.Type)
			}
			if parts[3] != tc.Set {
				t.Fatalf("input: %s\nset: %s\nexpected:%s", tc.Input, parts[3], tc.Set)
			}
		})
	}
}
