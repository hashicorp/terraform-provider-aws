// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TestAccODBDBNodeAction_existingNode interrupts an existing node. The opt-in prevents broad ODB
// acceptance runs from stopping a node, and no AWS infrastructure enters state.
func TestAccODBDBNodeAction_existingNode(t *testing.T) {
	if os.Getenv("TF_ACC_ODB_DB_NODE_ACTIONS") != "1" {
		t.Skip("set TF_ACC_ODB_DB_NODE_ACTIONS=1 to allow stop, start, and reboot of an existing DB node")
	}

	// The entire sequence must reserve the selected node without parallel tests.
	acctest.RunSerialTests1Level(t, map[string]func(*testing.T){
		"actions": testAccODBDBNodeAction_existingNode,
	}, 0)
}

func testAccODBDBNodeAction_existingNode(t *testing.T) {
	ctx := acctest.Context(t)
	clusterID := os.Getenv("TF_ACC_ODB_CLOUD_VM_CLUSTER_ID")
	nodeID := os.Getenv("TF_ACC_ODB_DB_NODE_ID")
	const trigger = "terraform_data.test"

	stopConfig := testAccDBNodeActionConfig(clusterID, nodeID, "stop", 1)
	startConfig := testAccDBNodeActionConfig(clusterID, nodeID, "start", 2)
	rebootConfig := testAccDBNodeActionConfig(clusterID, nodeID, "reboot", 3)
	repeatRebootConfig := testAccDBNodeActionConfig(clusterID, nodeID, "reboot", 4)

	unchangedStep := func(config string, status odbtypes.DbNodeResourceStatus) resource.TestStep {
		return resource.TestStep{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
					testAccDBNodeActionPlanCheck{},
				},
			},
			Check: testAccCheckDBNodeActionStatus(ctx, t, clusterID, nodeID, status),
		}
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			if clusterID == "" || nodeID == "" {
				t.Fatal("TF_ACC_ODB_CLOUD_VM_CLUSTER_ID and TF_ACC_ODB_DB_NODE_ID must identify an existing, disposable Exadata DB node")
			}

			// The static partition service catalog omits ODB. Read the selected
			// resources to verify availability, including at a custom endpoint.
			conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
			cluster, err := conn.GetCloudVmCluster(ctx, &odb.GetCloudVmClusterInput{
				CloudVmClusterId: aws.String(clusterID),
			})
			if err != nil {
				t.Fatalf("reading existing cloud VM cluster: %s", err)
			}
			if cluster == nil || cluster.CloudVmCluster == nil {
				t.Fatal("reading existing cloud VM cluster: empty result")
			}
			if cluster.CloudVmCluster.Status != odbtypes.ResourceStatusAvailable || aws.ToString(cluster.CloudVmCluster.CloudExadataInfrastructureId) == "" {
				t.Fatal("the test requires an AVAILABLE Exadata cloud VM cluster")
			}
			if err := testAccCheckDBNodeActionStatus(ctx, t, clusterID, nodeID, odbtypes.DbNodeResourceStatusAvailable)(nil); err != nil {
				t.Fatalf("checking existing DB node before mutation: %s", err)
			}

			// t.Context is canceled before cleanup. Preserve logging but give recovery
			// its own deadline so a failed stop/start test can still restore the node.
			t.Cleanup(func() {
				cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Hour)
				defer cancel()
				if err := testAccRestoreDBNodeAction(cleanupCtx, conn, clusterID, nodeID); err != nil {
					t.Errorf("restoring existing DB node to AVAILABLE; manual recovery may be required: %s", err)
				}
			})
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckDBNodeActionStatus(ctx, t, clusterID, nodeID, odbtypes.DbNodeResourceStatusAvailable),
		Steps: []resource.TestStep{
			{
				Config: stopConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(trigger, plancheck.ResourceActionCreate),
						testAccDBNodeActionPlanCheck{
							operation: "stop",
							// Verify planning did not stop the node before Terraform applies the action.
							checkStatus: testAccCheckDBNodeActionStatus(ctx, t, clusterID, nodeID, odbtypes.DbNodeResourceStatusAvailable),
						},
					},
				},
				Check: testAccCheckDBNodeActionStatus(ctx, t, clusterID, nodeID, odbtypes.DbNodeResourceStatusStopped),
			},
			unchangedStep(stopConfig, odbtypes.DbNodeResourceStatusStopped),
			{
				Config: startConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(trigger, plancheck.ResourceActionUpdate),
						testAccDBNodeActionPlanCheck{operation: "start"},
					},
				},
				Check: testAccCheckDBNodeActionStatus(ctx, t, clusterID, nodeID, odbtypes.DbNodeResourceStatusAvailable),
			},
			unchangedStep(startConfig, odbtypes.DbNodeResourceStatusAvailable),
			{
				Config: rebootConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(trigger, plancheck.ResourceActionUpdate),
						testAccDBNodeActionPlanCheck{operation: "reboot"},
					},
				},
				Check: testAccCheckDBNodeActionStatus(ctx, t, clusterID, nodeID, odbtypes.DbNodeResourceStatusAvailable),
			},
			unchangedStep(rebootConfig, odbtypes.DbNodeResourceStatusAvailable),
			{
				Config: repeatRebootConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(trigger, plancheck.ResourceActionUpdate),
						testAccDBNodeActionPlanCheck{operation: "reboot"},
					},
				},
				Check: testAccCheckDBNodeActionStatus(ctx, t, clusterID, nodeID, odbtypes.DbNodeResourceStatusAvailable),
			},
			unchangedStep(repeatRebootConfig, odbtypes.DbNodeResourceStatusAvailable),
		},
	})
}

// ExpectEmptyPlan does not inspect action_invocations, so assert these separately.
type testAccDBNodeActionPlanCheck struct {
	operation   string
	checkStatus resource.TestCheckFunc
}

func (c testAccDBNodeActionPlanCheck) CheckPlan(ctx context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	wantCount := 0
	if c.operation != "" {
		wantCount = 1
	}
	if got := len(req.Plan.ActionInvocations); got != wantCount {
		resp.Error = fmt.Errorf("expected %d DB node action invocations, got %d", wantCount, got)
		return
	}
	if wantCount == 0 {
		return
	}

	wantAddress := "action.aws_odb_" + c.operation + "_db_node.test"
	invocation := req.Plan.ActionInvocations[0]
	if invocation == nil || invocation.Address != wantAddress {
		resp.Error = fmt.Errorf("expected DB node action invocation %s", wantAddress)
		return
	}
	if invocation.LifecycleActionTrigger == nil || invocation.LifecycleActionTrigger.TriggeringResourceAddress != "terraform_data.test" {
		resp.Error = errors.New("expected DB node action to be triggered by terraform_data.test")
		return
	}
	if c.checkStatus != nil {
		if err := c.checkStatus(nil); err != nil {
			resp.Error = fmt.Errorf("planning must not change the DB node status: %w", err)
		}
	}
}

func testAccCheckDBNodeActionStatus(ctx context.Context, t *testing.T, clusterID, nodeID string, want odbtypes.DbNodeResourceStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
		output, err := conn.GetDbNode(ctx, &odb.GetDbNodeInput{
			CloudVmClusterId: aws.String(clusterID),
			DbNodeId:         aws.String(nodeID),
		})
		if err != nil {
			return fmt.Errorf("reading existing DB node: %w", err)
		}
		if output == nil || output.DbNode == nil {
			return errors.New("reading existing DB node: empty result")
		}
		if aws.ToString(output.DbNode.DbNodeId) != nodeID {
			return errors.New("reading existing DB node: response identifies a different node")
		}
		if got := output.DbNode.Status; got != want {
			return fmt.Errorf("expected DB node status %s, got %s", want, got)
		}
		return nil
	}
}

func testAccRestoreDBNodeAction(ctx context.Context, conn *odb.Client, clusterID, nodeID string) error {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	started := false
	availableChecks := 0

	for {
		output, err := conn.GetDbNode(ctx, &odb.GetDbNodeInput{
			CloudVmClusterId: aws.String(clusterID),
			DbNodeId:         aws.String(nodeID),
		})
		if err != nil {
			return fmt.Errorf("reading DB node during cleanup: %w", err)
		}
		if output == nil || output.DbNode == nil {
			return errors.New("reading DB node during cleanup: empty result")
		}
		if aws.ToString(output.DbNode.DbNodeId) != nodeID {
			return errors.New("reading DB node during cleanup: response identifies a different node")
		}
		if output.DbNode.Status != odbtypes.DbNodeResourceStatusAvailable {
			availableChecks = 0
		}
		switch status := output.DbNode.Status; status {
		case odbtypes.DbNodeResourceStatusAvailable:
			// An interrupted apply can leave stale AVAILABLE reads while the
			// service begins the accepted operation.
			availableChecks++
			if availableChecks == 3 {
				return nil
			}
		case odbtypes.DbNodeResourceStatusStopped:
			if !started {
				_, err := conn.StartDbNode(ctx, &odb.StartDbNodeInput{
					CloudVmClusterId: aws.String(clusterID),
					DbNodeId:         aws.String(nodeID),
				}, func(options *odb.Options) { options.RetryMaxAttempts = 1 })
				if err != nil {
					return fmt.Errorf("starting DB node during cleanup: %w", err)
				}
				started = true
			}
		case odbtypes.DbNodeResourceStatusStopping, odbtypes.DbNodeResourceStatusStarting, odbtypes.DbNodeResourceStatusUpdating:
		default:
			return fmt.Errorf("cannot automatically restore DB node in status %s", status)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for DB node cleanup: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func testAccDBNodeActionConfig(clusterID, nodeID, operation string, revision int) string {
	return fmt.Sprintf(`
action "aws_odb_stop_db_node" "test" {
  config {
    cloud_vm_cluster_id = %[1]q
    db_node_id          = %[2]q
  }
}

action "aws_odb_start_db_node" "test" {
  config {
    cloud_vm_cluster_id = %[1]q
    db_node_id          = %[2]q
  }
}

action "aws_odb_reboot_db_node" "test" {
  config {
    cloud_vm_cluster_id = %[1]q
    db_node_id          = %[2]q
    timeout             = 3600
  }
}

resource "terraform_data" "test" {
  input = %[4]d

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.aws_odb_%[3]s_db_node.test]
    }
  }
}
`, clusterID, nodeID, operation, revision)
}
