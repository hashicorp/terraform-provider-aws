/**
 * Copyright IBM Corp. 2014, 2026
 * SPDX-License-Identifier: MPL-2.0
 */

/** IVS Chat message review handler */
exports.handler = async function ({ Content }) {
    return {
        ReviewResult: "ALLOW",
        Content: `${Content} - edited by Lambda`
    };
}
