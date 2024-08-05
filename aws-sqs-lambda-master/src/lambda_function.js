exports.handler = async (event, context) => {
    console.log('Received event:', prettyPrint(event))
    console.log('Received context:', prettyPrint(context))
};

function prettyPrint(obj) {
    return JSON.stringify(obj, null, 2);
}
