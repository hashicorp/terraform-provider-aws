# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

import json

def lambda_handler(event, context):
    response = {"message": "Hello"}

    response_body = {"application/json": {"body": json.dumps(response)}}

    action_response = {
        "actionGroup": event["actionGroup"],
        "apiPath": event["apiPath"],
        "httpMethod": event["httpMethod"],
        "httpStatusCode": 200,
        "response": response_body,
    }

    session_attributes = event["sessionAttributes"]
    prompt_session_attributes = event["promptSessionAttributes"]

    return {
        "messageVersion": "1.0",
        "response" : action_response,
        "sessionAttributes": session_attributes,
        "promptSessionAttributes": prompt_session_attributes,
    }