import { util } from '@aws-appsync/utils';

/**
 * Request a single item from the attached DynamoDB table update
 * @param ctx the request context
 */
export function request(ctx) {
  return {
    operation: 'GetItem',
    key: util.dynamodb.toMapValues({ id: ctx.args.id }),
  };
}

/**
 * Returns the DynamoDB result directly
 * @param ctx the request context
 */
export function response(ctx) {
  return ctx.result;
}