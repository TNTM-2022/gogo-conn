const pbjs = require("protobufjs/cli/pbjs");
const fs = require('fs');
const  [input, output=input] = process.argv.slice(2)
pbjs.main(["--target", "json", input], (err, output) => {
    if (err)
        throw err;

    fs.writeFileSync(`${output}.json`, output );
    // do something with output
});