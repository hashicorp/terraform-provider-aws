#!/usr/bin/env bash

set -Eeuxo pipefail

SETUP_CONFIG='./import'
TEST_CONFIG='./refresh'
CONFIG=''

PLUGINS_DIR='./plugins'
PLUGIN_PATH=$PLUGINS_DIR/registry.terraform.io/hashicorp/aws/99.99.99/$(go env GOOS)_$(go env GOARCH)
TF12_PLUGIN_PATH=$PLUGINS_DIR

ORGANIZATION_ADDR='aws_organizations_organization.test'
POLICY_ADDR='aws_organizations_policy.test'
POLICY_ID='p-FullAWSAccess'

function cleanup {
    if [ -n "$CONFIG" ]; then
        echo "Cleaning up: terraform destroy -auto-approve"
        if terraform state list | grep -q $POLICY_ADDR; then
            terraform state rm $POLICY_ADDR
        fi
        terraform destroy -auto-approve "$CONFIG"
    fi
}
trap cleanup EXIT

rm -rf $PLUGINS_DIR
mkdir -p "$PLUGIN_PATH"
go build -o "$PLUGIN_PATH"/terraform-provider-aws_v99.99.99_x5 ../..
cp "$PLUGIN_PATH"/terraform-provider-aws_v99.99.99_x5 $TF12_PLUGIN_PATH

CONFIG=$SETUP_CONFIG
rm -f .terraform/plugin_path
terraform init -input=false $CONFIG
terraform apply -input=false -auto-approve -target $ORGANIZATION_ADDR $CONFIG

terraform import -input=false -config=$CONFIG $POLICY_ADDR $POLICY_ID

CONFIG=$TEST_CONFIG
terraform init -input=false -verify-plugins=false -plugin-dir=$PLUGINS_DIR $CONFIG
terraform plan -input=false $CONFIG
