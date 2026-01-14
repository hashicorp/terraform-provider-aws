/**
 * Copyright IBM Corp. 2014, 2026
 * SPDX-License-Identifier: MPL-2.0
 */

// Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-route-keys-connect-disconnect.html
// The $disconnect route is executed after the connection is closed.
// The connection can be closed by the server or by the client. As the connection is already closed when it is executed,
// $disconnect is a best-effort event.
// API Gateway will try its best to deliver the $disconnect event to your integration, but it cannot guarantee delivery.

const { DynamoDBClient } = require('@aws-sdk/client-dynamodb');
const { DynamoDBDocumentClient, DeleteCommand } = require('@aws-sdk/lib-dynamodb');

const client = new DynamoDBClient({});
const ddb = DynamoDBDocumentClient.from(client);

exports.handler = async event => {
	const deleteParams = {
		TableName: process.env.TABLE_NAME,
		Key: {
			connectionId: event.requestContext.connectionId
		}
	};

	try {
		await ddb.send(new DeleteCommand(deleteParams));
	} catch (err) {
		return { statusCode: 500, body: 'Failed to disconnect: ' + JSON.stringify(err) };
	}

	return { statusCode: 200, body: 'Disconnected.' };
};
