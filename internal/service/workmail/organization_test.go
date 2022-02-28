package workmail_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkMailConn

	input := &workmail.ListOrganizationsInput{}

	_, err := conn.ListOrganizations(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
