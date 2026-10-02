---
subcategory: "Marketplace Agreement"
layout: "aws"
page_title: "AWS: aws_marketplaceagreement_agreement"
description: |-
  Manages an AWS Marketplace agreement, subscribing the account to an AWS Marketplace offer.
---

# Resource: aws_marketplaceagreement_agreement

Manages an AWS Marketplace agreement. Creating the resource accepts an offer's terms on behalf of the account, the same as subscribing to the offer in the AWS Marketplace console. Destroying it cancels the agreement.

Look up the `agreement_proposal_id` with the AWS Marketplace Discovery API's [`GetOffer`](https://docs.aws.amazon.com/marketplace/latest/APIReference/API_marketplace-discovery_GetOffer.html) operation, and the term IDs with [`GetOfferTerms`](https://docs.aws.amazon.com/marketplace/latest/APIReference/API_marketplace-discovery_GetOfferTerms.html). See [Work with the Agreement API as a buyer](https://docs.aws.amazon.com/marketplace/latest/APIReference/work-with-agreement-api-buyer.html) for which terms each pricing model requires.

~> **Note:** Changing any argument cancels the agreement and accepts a new one. AWS Marketplace only lets the acceptor cancel some agreements (for example, free and usage-based ones) through the API. Destroying or replacing an agreement it won't cancel fails after the delete timeout, and the agreement stays active.

~> **Note:** An agreement that is no longer `ACTIVE`, including one that renewed into a new agreement, is removed from state.

## Example Usage

### Free Offer

```terraform
# Datadog Operator EKS add-on.
resource "aws_marketplaceagreement_agreement" "example" {
  agreement_proposal_id = "at-6nuclnsh88uc2s8hs1mjzr5eu"

  requested_term {
    id = "term-331241c92e99c39ff990052b4ae78d0165d892136e6af1aee8d55b9c28b540f9" # LegalTerm
  }

  requested_term {
    id = "term-17e763eebddabad9921f51d9360900fb6fd07cd91a0ab8699859b40d2cd7109a" # SupportTerm
  }
}

resource "aws_eks_addon" "example" {
  cluster_name = aws_eks_cluster.example.name
  addon_name   = "datadog_operator"

  depends_on = [aws_marketplaceagreement_agreement.example]
}
```

### Contract Offer

```terraform
resource "aws_marketplaceagreement_agreement" "example" {
  agreement_proposal_id = "at-0123456789abcdefghijklmno"

  requested_term {
    id = "term-legal-0123456789abcdef"
  }

  requested_term {
    id = "term-configurable-pricing-0123456789abcdef"

    configuration {
      configurable_upfront_pricing_term {
        selector_value = "P12M"

        dimension {
          dimension_key   = "Users"
          dimension_value = 50
        }
      }
    }
  }

  requested_term {
    id = "term-renewal-0123456789abcdef"

    configuration {
      renewal_term {
        enable_auto_renew = false
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `agreement_proposal_id` - (Required) ID of the agreement proposal for the offer, from the `agreementProposalId` field of the Discovery API's `GetOffer` response.
* `requested_term` - (Required) Terms of the offer to accept. Include every term the offer's pricing model requires. See [`requested_term` Block](#requested_term-block) below.

The following arguments are optional:

* `timeouts` - (Optional) See [Timeouts](#timeouts) below.

### `requested_term` Block

The `requested_term` block supports:

* `configuration` - (Optional) Buyer-supplied configuration, for the term types that take one. See [`requested_term.configuration` Block](#requested_termconfiguration-block) below.
* `id` - (Required) ID of the term, from the Discovery API's `GetOfferTerms` response.

### `requested_term.configuration` Block

The `requested_term.configuration` block requires exactly one of the following:

* `configurable_upfront_pricing_term` - (Optional) Configuration for a `ConfigurableUpfrontPricingTerm`. See [`requested_term.configuration.configurable_upfront_pricing_term` Block](#requested_termconfigurationconfigurable_upfront_pricing_term-block) below.
* `renewal_term` - (Optional) Configuration for a `RenewalTerm`. See [`requested_term.configuration.renewal_term` Block](#requested_termconfigurationrenewal_term-block) below.
* `variable_payment_term` - (Optional) Configuration for a `VariablePaymentTerm`. See [`requested_term.configuration.variable_payment_term` Block](#requested_termconfigurationvariable_payment_term-block) below.

### `requested_term.configuration.configurable_upfront_pricing_term` Block

The `requested_term.configuration.configurable_upfront_pricing_term` block supports:

* `dimension` - (Required) Dimensions to purchase. See [`requested_term.configuration.configurable_upfront_pricing_term.dimension` Block](#requested_termconfigurationconfigurable_upfront_pricing_termdimension-block) below.
* `selector_value` - (Required) Duration of the purchase, from the rate card's `selector.value` (for example, `P12M`).

### `requested_term.configuration.configurable_upfront_pricing_term.dimension` Block

The `requested_term.configuration.configurable_upfront_pricing_term.dimension` block supports:

* `dimension_key` - (Required) Key of the dimension, from the rate card's `dimensionKey`.
* `dimension_value` - (Required) Number of units of the dimension to purchase.

### `requested_term.configuration.renewal_term` Block

The `requested_term.configuration.renewal_term` block supports:

* `enable_auto_renew` - (Required) Whether the agreement renews automatically at its end date.

### `requested_term.configuration.variable_payment_term` Block

The `requested_term.configuration.variable_payment_term` block supports:

* `expiration_duration` - (Optional) ISO 8601 duration after which a payment request is approved automatically (for example, `P10D`). Required when `payment_request_approval_strategy` is `AUTO_APPROVE_ON_EXPIRATION`.
* `payment_request_approval_strategy` - (Required) How the seller's payment requests are approved. Valid values: `AUTO_APPROVE_ON_EXPIRATION`, `WAIT_FOR_APPROVAL`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `acceptance_time` - Time the agreement was accepted, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.8).
* `agreement_id` - ID of the agreement.
* `agreement_type` - Type of the agreement, for example `PurchaseAgreement`.
* `end_time` - Time the agreement ends, in RFC3339 format. Not set for agreements without an end date, such as pay-as-you-go agreements.
* `offer_id` - ID of the offer the agreement was accepted from.
* `proposer_account_id` - AWS account ID of the seller that proposed the agreement.
* `start_time` - Time the agreement starts, in RFC3339 format.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `delete` - (Default `10m`) AWS Marketplace rejects cancelling an agreement for about a minute and a half after it's accepted, and the provider retries until then.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_marketplaceagreement_agreement.example
  identity = {
    agreement_id = "agmt-0123456789abcdefghijklmno"
  }
}

resource "aws_marketplaceagreement_agreement" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `agreement_id` (String) ID of the agreement.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Marketplace agreements using the `agreement_id`. For example:

```terraform
import {
  to = aws_marketplaceagreement_agreement.example
  id = "agmt-0123456789abcdefghijklmno"
}
```

Using `terraform import`, import Marketplace agreements using the `agreement_id`. For example:

```console
% terraform import aws_marketplaceagreement_agreement.example agmt-0123456789abcdefghijklmno
```

The API doesn't return the proposal an agreement was accepted from, so `agreement_proposal_id` is set from configuration on the first apply after import. That apply doesn't change the agreement.
