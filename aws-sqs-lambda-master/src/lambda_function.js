exports.handler = async (event, context) => {
    if (event.Records[0].body === "error") {
        throw new Error('This is an error');
    }
    if (event.Records[0].body === "timeout") {
        console.time("timeout");
        await sleep(40 * 1000);
        console.timeEnd("timeout");
    }
    console.log('Received event:', prettyPrint(event));
    console.log('Received context:', prettyPrint(context));
};

function prettyPrint(obj) {
    return JSON.stringify(obj, null, 2);
}

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}