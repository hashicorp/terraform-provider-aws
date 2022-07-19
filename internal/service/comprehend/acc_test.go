package comprehend_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/comprehend"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ComprehendConn
	ctx := context.Background()

	input := &comprehend.ListEntityRecognizersInput{}

	_, err := conn.ListEntityRecognizers(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
