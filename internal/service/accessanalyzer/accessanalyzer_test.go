package accessanalyzer_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// AccessAnalyzer is limited to one per region, so run serially locally and in TeamCity.
func TestAccAccessAnalyzer_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Analyzer": {
			"basic":             testAccAnalyzer_basic,
			"disappears":        testAccAnalyzer_disappears,
			"Tags":              testAccAnalyzer_Tags,
			"Type_Organization": testAccAnalyzer_Type_Organization,
		},
		"ArchiveRule": {
			"basic":          testAccAnalyzerArchiveRule_basic,
			"disappears":     testAccAnalyzerArchiveRule_disappears,
			"update_filters": testAccAnalyzerArchiveRule_updateFilters,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerConn()

	input := &accessanalyzer.ListAnalyzersInput{}

	_, err := conn.ListAnalyzersWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
