/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import { SSMClient, PutParameterCommand } from "@aws-sdk/client-ssm";

const ssmClient = new SSMClient();

export const handler = async (event) => {
    let tf_key = "tf";
    if (tf_key in event) {
        if (event[tf_key].action == "delete" && process.env.TEST_DATA != "") {
            try {
                await ssmClient.send(new PutParameterCommand({
                    Name: process.env.TEST_DATA,
                    Value: JSON.stringify(event),
                    Type: "String"
                }));
            } catch (error) {
                console.error("Error putting parameter:", error);
                throw error;
            }
        }
    }
    return event;
}
