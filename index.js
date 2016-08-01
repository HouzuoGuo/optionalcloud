var child_process = require('child_process');

exports.handler = (event, context, callback) => {
    var proc = child_process.spawnSync('./optionalcloud', [], {'input': JSON.stringify(event)});
    console.log('stderr: ' + proc.stderr.toString());
    if (proc.status == 0) {
        // Process exited successfully
        // The standard output is a JSON object of status integer, header (key-value map), and body-json object.
        // Use the status integer to construct a success callback or error callback
        var resp_obj = JSON.parse(proc.stdout.toString());
        var http_status = resp_obj['status'];
        if (~~(http_status / 200) == 1) {
            // Assume there is only one normal response status (2xx/3xx) that is also the default mapping on API gateway
            callback(null, resp_obj);
        } else {
            // Error callbacks carry a string message prefixed by status integer and a colon symbol
            // By convention, these callbacks do not carry header map (hence cannot set response headers) and expect body-json to be a string
            callback('' + http_status + ': ' + resp_obj['body-json']);
        }
    } else {
        callback('500:' + proc.error, null);
    }
};
