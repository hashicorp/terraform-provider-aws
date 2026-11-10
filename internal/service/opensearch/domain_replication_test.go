// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchDomainReplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replication opensearch.DescribeDomainReplicationOutput
	rName := testAccRandomDomainReplicationName(t)
	resourceName := "aws_opensearch_domain_replication.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainReplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainReplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainReplicationExists(ctx, t, resourceName, &replication),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, "aws_opensearch_domain.primary", names.AttrDomainName),
					resource.TestCheckResourceAttrPair(resourceName, "peer_domain_arn", "aws_opensearch_domain.secondary", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "domain_arn", "aws_opensearch_domain.primary", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, "PRIMARY"),
					resource.TestCheckResourceAttr(resourceName, "peer_role", "SECONDARY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrDomainName, "peer_domain_arn"),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTimeouts},
			},
		},
	})
}

func TestAccOpenSearchDomainReplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replication opensearch.DescribeDomainReplicationOutput
	rName := testAccRandomDomainReplicationName(t)
	resourceName := "aws_opensearch_domain_replication.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainReplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainReplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainReplicationExists(ctx, t, resourceName, &replication),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfopensearch.ResourceDomainReplication, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccOpenSearchDomainReplication_switchover(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replication opensearch.DescribeDomainReplicationOutput
	rName := testAccRandomDomainReplicationName(t)
	resourceName := "aws_opensearch_domain_replication.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainReplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainReplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainReplicationExists(ctx, t, resourceName, &replication),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, "PRIMARY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
				),
			},
			{
				PreConfig: func() {
					testAccDomainReplicationPromote(ctx, t, rName, false)
				},
				Config: testAccDomainReplicationConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainReplicationLogState(t, resourceName),
					testAccCheckDomainReplicationExists(ctx, t, resourceName, &replication),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, "SECONDARY"),
					resource.TestCheckResourceAttr(resourceName, "peer_role", "PRIMARY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
				),
			},
		},
	})
}

func TestAccOpenSearchDomainReplication_failover(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replication opensearch.DescribeDomainReplicationOutput
	rName := testAccRandomDomainReplicationName(t)
	resourceName := "aws_opensearch_domain_replication.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainReplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainReplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainReplicationExists(ctx, t, resourceName, &replication),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, "PRIMARY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
				),
			},
			{
				PreConfig: func() {
					testAccDomainReplicationPromote(ctx, t, rName, true)
				},
				Config: testAccDomainReplicationConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainReplicationLogState(t, resourceName),
					testAccCheckDomainReplicationExists(ctx, t, resourceName, &replication),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, "SECONDARY"),
					resource.TestCheckResourceAttr(resourceName, "peer_role", "PRIMARY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "EJECTED"),
				),
			},
		},
	})
}

// testAccRandomDomainReplicationName leaves room for the -p/-s suffix (28 max).
func testAccRandomDomainReplicationName(t *testing.T) string {
	return testAccRandomDomainName(t)[:26]
}

// testAccDomainReplicationPromote promotes the secondary outside Terraform, by switchover or
// by failover, and waits until the secondary reports PRIMARY and the transition has settled.
func testAccDomainReplicationPromote(ctx context.Context, t *testing.T, rName string, failover bool) {
	t.Helper()

	primaryName, secondaryName := rName+"-p", rName+"-s"
	conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)

	primary, err := tfopensearch.FindDomainByName(ctx, conn, primaryName)
	if err != nil {
		t.Fatalf("describing domain %s: %s", primaryName, err)
	}
	secondary, err := tfopensearch.FindDomainByName(ctx, conn, secondaryName)
	if err != nil {
		t.Fatalf("describing domain %s: %s", secondaryName, err)
	}
	primaryARN, secondaryARN := aws.ToString(primary.ARN), aws.ToString(secondary.ARN)

	if failover {
		out, err := conn.FailoverReplication(ctx, &opensearch.FailoverReplicationInput{
			AllowDataLoss: aws.Bool(true),
			DomainName:    aws.String(secondaryName),
			PeerDomainArn: aws.String(primaryARN),
		})
		if err != nil {
			t.Fatalf("FailoverReplication on %s: %s", secondaryName, err)
		}
		t.Logf("FailoverReplication on %s returned status %s", secondaryName, out.Status)
	} else {
		out, err := conn.SwitchoverReplication(ctx, &opensearch.SwitchoverReplicationInput{
			DomainName:    aws.String(secondaryName),
			PeerDomainArn: aws.String(primaryARN),
		})
		if err != nil {
			t.Fatalf("SwitchoverReplication on %s: %s", secondaryName, err)
		}
		t.Logf("SwitchoverReplication on %s returned status %s", secondaryName, out.Status)
	}

	settled := func(s awstypes.ReplicationStatus) bool {
		return s != awstypes.ReplicationStatusSwitchingOver && s != awstypes.ReplicationStatusFailingOver
	}

	deadline := time.Now().Add(60 * time.Minute)
	for {
		sec, secErr := tfopensearch.FindDomainReplicationByTwoPartKey(ctx, conn, secondaryName, primaryARN)
		pri, priErr := tfopensearch.FindDomainReplicationByTwoPartKey(ctx, conn, primaryName, secondaryARN)
		t.Logf("promote poll: %s=%s; %s=%s", secondaryName, describeReplicationForLog(sec, primaryARN, secErr), primaryName, describeReplicationForLog(pri, secondaryARN, priErr))

		if secErr == nil && sec.Role == awstypes.ReplicationRolePrimary && settled(replicationPeerStatus(sec, primaryARN)) &&
			(priErr != nil || settled(replicationPeerStatus(pri, secondaryARN))) {
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s to become PRIMARY", secondaryName)
		}

		time.Sleep(30 * time.Second)
	}
}

// replicationPeerPair returns the entry for peerDomainARN in out. out comes from
// FindDomainReplicationByTwoPartKey, which guarantees the entry is present.
func replicationPeerPair(out *opensearch.DescribeDomainReplicationOutput, peerDomainARN string) awstypes.ReplicationPeer {
	for _, p := range out.Peers {
		if aws.ToString(p.PeerDomainArn) == peerDomainARN {
			return p
		}
	}

	return awstypes.ReplicationPeer{}
}

func replicationPeerStatus(out *opensearch.DescribeDomainReplicationOutput, peerDomainARN string) awstypes.ReplicationStatus {
	return replicationPeerPair(out, peerDomainARN).Status
}

func describeReplicationForLog(out *opensearch.DescribeDomainReplicationOutput, peerDomainARN string, err error) string {
	if err != nil {
		return fmt.Sprintf("error(%s)", err)
	}

	p := replicationPeerPair(out, peerDomainARN)
	return fmt.Sprintf("role=%s peer_role=%s status=%s msg=%q", out.Role, p.PeerRole, p.Status, aws.ToString(p.StatusMessage))
}

func testAccCheckDomainReplicationLogState(t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		a := rs.Primary.Attributes
		t.Logf("%s state: id=%s role=%s peer_role=%s status=%s status_message=%q", n, rs.Primary.ID, a[names.AttrRole], a["peer_role"], a[names.AttrStatus], a[names.AttrStatusMessage])

		return nil
	}
}

func testAccCheckDomainReplicationExists(ctx context.Context, t *testing.T, n string, v *opensearch.DescribeDomainReplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)

		domainName, peerDomainARN, err := tfopensearch.DomainReplicationParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfopensearch.FindDomainReplicationByTwoPartKey(ctx, conn, domainName, peerDomainARN)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDomainReplicationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_domain_replication" {
				continue
			}

			domainName, peerDomainARN, err := tfopensearch.DomainReplicationParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfopensearch.FindDomainReplicationByTwoPartKey(ctx, conn, domainName, peerDomainARN)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch Domain Replication %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// Pairing needs fine-grained access control, node-to-node encryption, encryption at rest, HTTPS,
// a matching engine version, and an empty secondary. A brand-new secondary is empty.
func testAccDomainReplicationConfig_basic(rName string) string {
	return fmt.Sprintf(`
locals {
  domain_names = ["%[1]s-p", "%[1]s-s"]
}

resource "aws_opensearch_domain" "primary" {
  domain_name    = local.domain_names[0]
  engine_version = "OpenSearch_3.7"

  cluster_config {
    instance_type  = "r6g.large.search"
    instance_count = 1
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 20
  }

  encrypt_at_rest {
    enabled = true
  }

  node_to_node_encryption {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = true
    master_user_options {
      master_user_name     = "tfacctest"
      master_user_password = "Barbarbarbar1!"
    }
  }
}

resource "aws_opensearch_domain" "secondary" {
  domain_name    = local.domain_names[1]
  engine_version = aws_opensearch_domain.primary.engine_version

  cluster_config {
    instance_type  = "r6g.large.search"
    instance_count = 1
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 20
  }

  encrypt_at_rest {
    enabled = true
  }

  node_to_node_encryption {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = true
    master_user_options {
      master_user_name     = "tfacctest"
      master_user_password = "Barbarbarbar1!"
    }
  }
}

resource "aws_opensearch_domain_replication" "test" {
  domain_name     = aws_opensearch_domain.primary.domain_name
  peer_domain_arn = aws_opensearch_domain.secondary.arn
}
`, rName)
}
