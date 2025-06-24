package main

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func test1(t *testing.T) {
	ctx := acctest.Context(t)

	const resourceName = "aws_ec2_serial_console_access.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSerialConsoleAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSerialConsoleAccessConfig_basic(false),
				ConfigStateChecks: []statecheck.StateCheck{
					// ruleid: knownvalue_account_id
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringExact(acctest.AccountID(ctx))),
					// ok: knownvalue_account_id
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), tfknownvalue.AccountID()),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						// ruleid: knownvalue_account_id
						names.AttrAccountID: knownvalue.StringExact(acctest.AccountID(ctx)),
					}),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						// ok: knownvalue_account_id
						names.AttrAccountID: tfknownvalue.AccountID(),
					}),
				},
			},
		},
	})

}
