const util = require('util');

exports.handler = async (event, context) => {
    if (event.Records[0].body === "error") {
        throw new Error('This is an error');
    }
    if (event.Records[0].body === "timeout") {
        console.log("sleep in 40s");
        sleep(40 * 1000);
    }
    console.log('Received event:', prettyPrint(event));
    console.log('Received context:', prettyPrint(context));
};

function prettyPrint(obj) {
    return JSON.stringify(obj, null, 2);
}

async function sleep(milSec) {
    util.promisify(setTimeout);
}