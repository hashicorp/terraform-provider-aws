/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

/** IVS Chat message review handler */
exports.handler = async function ({ Content }) {
    return {
        ReviewResult: "ALLOW",
        Content: `${Content} - edited by Lambda`
    };
}
