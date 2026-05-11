/**
 * Copyright IBM Corp. 2014, 2026
 * SPDX-License-Identifier: MPL-2.0
 */

export const handler = async (event) => {
    let tf_key = "tf";
    if (tf_key in event) {
        if (event[tf_key].action === "update") {
            throw new Error("Update operation failed");
        }
    }
    return event;
}
