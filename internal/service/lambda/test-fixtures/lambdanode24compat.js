var http = require('http')

exports.handler = async function(event, context) {
    return {
        body: "ok",
        statusCode: 200
    }
}
