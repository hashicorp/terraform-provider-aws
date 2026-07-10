#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Authenticate internally using TeamCity's system-provided tokens
curl -H "Authorization: Bearer %system.teamcity.auth.userId%:%system.teamcity.auth.password%" \
     -H "Accept: application/json" \
     "%teamcity.serverUrl%/app/rest/testOccurrences?locator=build:(id:%teamcity.build.id%)"
