// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mediaconvert_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	terraformsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmediaconvert "github.com/hashicorp/terraform-provider-aws/internal/service/mediaconvert"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Reserved queues require a yearly commitment, so exercise their update requests locally.
func TestResourceQueueUpdate_concurrentJobs(t *testing.T) {
	t.Parallel()

	for _, pricingPlan := range []types.PricingPlan{types.PricingPlanOnDemand, types.PricingPlanReserved} {
		t.Run(string(pricingPlan), func(t *testing.T) {
			t.Parallel()

			var updateCalls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPut {
					updateCalls++
					var input map[string]any
					if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
						t.Errorf("decoding update request: %v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					value, present := input["concurrentJobs"]
					if want := pricingPlan == types.PricingPlanOnDemand; present != want {
						t.Errorf("concurrentJobs present = %t, want %t", present, want)
					}
					if present && value != float64(5) {
						t.Errorf("concurrentJobs = %v, want 5", value)
					}
					if got := input[names.AttrDescription]; got != "updated" {
						t.Errorf("description = %v, want updated", got)
					}
				}
				if _, err := fmt.Fprintf(w, `{"queue":{"name":"test","pricingPlan":%q,"concurrentJobs":5,"status":"ACTIVE","description":"original"}}`, pricingPlan); err != nil {
					t.Errorf("writing queue response: %v", err)
				}
			}))
			defer server.Close()

			ctx := t.Context()
			client := new(conns.AWSClient)
			client.SetServicePackages(ctx, map[string]conns.ServicePackage{names.MediaConvert: tfmediaconvert.ServicePackage(ctx)})
			config := conns.Config{
				AccessKey:               "test",
				SecretKey:               "test",
				Region:                  endpoints.UsEast1RegionID,
				Endpoints:               map[string]string{names.MediaConvert: server.URL},
				SkipCredsValidation:     true,
				SkipRequestingAccountId: true,
				SharedConfigFiles:       []string{},
				SharedCredentialsFiles:  []string{},
			}
			client, diags := config.ConfigureProvider(ctx, client)
			if diags.HasError() {
				t.Fatal(diags)
			}

			queue := tfmediaconvert.ResourceQueue()
			d := schema.TestResourceDataRaw(t, queue.SchemaMap(), map[string]any{
				names.AttrName: "test",
				"pricing_plan": string(pricingPlan),
			})
			d.SetId("test")
			state, diags := queue.RefreshWithoutUpgrade(ctx, d.State(), client)
			if diags.HasError() {
				t.Fatal(diags)
			}
			diff := &terraformsdk.InstanceDiff{
				Attributes: map[string]*terraformsdk.ResourceAttrDiff{
					names.AttrDescription: {Old: "original", New: "updated"},
				},
			}
			if _, diags := queue.Apply(ctx, state, diff, client); diags.HasError() {
				t.Fatal(diags)
			}
			if updateCalls != 1 {
				t.Errorf("UpdateQueue calls = %d, want 1", updateCalls)
			}
		})
	}
}

func TestAccMediaConvertQueue_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediaconvert", regexache.MustCompile(`queues/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "concurrent_jobs"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", string(types.PricingPlanOnDemand)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.QueueStatusActive)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMediaConvertQueue_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					acctest.CheckSDKResourceDisappears(ctx, t, tfmediaconvert.ResourceQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccMediaConvertQueue_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccQueueConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccMediaConvertQueue_reservationPlanSettings(t *testing.T) {
	acctest.Skip(t, "MediaConvert Reserved Queues are $400/month and cannot be deleted for 1 year.")

	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_reserved(rName, string(types.CommitmentOneYear), string(types.RenewalTypeAutoRenew), 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", string(types.PricingPlanReserved)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.commitment", string(types.CommitmentOneYear)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.renewal_type", string(types.RenewalTypeAutoRenew)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.reserved_slots", "1"),
				),
			},
			{
				Config: testAccQueueConfig_reserved(rName, string(types.CommitmentOneYear), string(types.RenewalTypeExpire), 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "pricing_plan", string(types.PricingPlanReserved)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.commitment", string(types.CommitmentOneYear)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.renewal_type", string(types.RenewalTypeExpire)),
					resource.TestCheckResourceAttr(resourceName, "reservation_plan_settings.0.reserved_slots", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMediaConvertQueue_withStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_status(rName, string(types.QueueStatusPaused)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.QueueStatusPaused)),
				),
			},
			{
				Config: testAccQueueConfig_status(rName, string(types.QueueStatusActive)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.QueueStatusActive)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMediaConvertQueue_withDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	description1 := acctest.RandomWithPrefix(t, "Description: ")
	description2 := acctest.RandomWithPrefix(t, "Description: ")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_description(rName, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description1),
				),
			},
			{
				Config: testAccQueueConfig_description(rName, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description2),
				),
			},
		},
	})
}

func TestAccMediaConvertQueue_withConcurrentJobs(t *testing.T) {
	ctx := acctest.Context(t)
	var queue types.Queue
	resourceName := "aws_media_convert_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConvertServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_concurrentJobs(rName, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "concurrent_jobs", "100"),
				),
			},
			{
				Config: testAccQueueConfig_concurrentJobs(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, t, resourceName, &queue),
					resource.TestCheckResourceAttr(resourceName, "concurrent_jobs", "5"),
				),
			},
		},
	})
}

func testAccCheckQueueDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_convert_queue" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).MediaConvertClient(ctx)

			_, err := tfmediaconvert.FindQueueByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Media Convert Queue %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckQueueExists(ctx context.Context, t *testing.T, n string, v *types.Queue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).MediaConvertClient(ctx)

		output, err := tfmediaconvert.FindQueueByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccQueueConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q
}
`, rName)
}

func testAccQueueConfig_status(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name   = %[1]q
  status = %[2]q
}
`, rName, status)
}

func testAccQueueConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccQueueConfig_concurrentJobs(rName string, concurrentJobs int) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name            = %[1]q
  concurrent_jobs = %[2]d
}
`, rName, concurrentJobs)
}

func testAccQueueConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q

  tags = {
    %[2]s = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccQueueConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name = %[1]q

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccQueueConfig_reserved(rName, commitment, renewalType string, reservedSlots int) string {
	return fmt.Sprintf(`
resource "aws_media_convert_queue" "test" {
  name         = %[1]q
  pricing_plan = %[2]q

  reservation_plan_settings {
    commitment     = %[3]q
    renewal_type   = %[4]q
    reserved_slots = %[5]d
  }
}
`, rName, string(types.PricingPlanReserved), commitment, renewalType, reservedSlots)
}
