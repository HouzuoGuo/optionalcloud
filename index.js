var child_process = require('child_process');

exports.handler = (event, context, callback) => {
    var proc = child_process.spawnSync('./optionalcloud', [], {'input': JSON.stringify(event)});
    if (proc.status == 0) {
        callback(null, JSON.parse(proc.stdout.toString()));
    } else {
        callback(proc.error, null);
    }
};
