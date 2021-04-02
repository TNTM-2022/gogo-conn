const pbjs = require("protobufjs/cli/pbjs");
const fs = require('fs');

pbjs.main(["--target", "json", "/Users/mac/Codes/golang/go-connector/test_proto/user.proto", "/Users/mac/Codes/golang/go-connector/test_proto/mail.proto"], (err, output) => {
    if (err)
        throw err;

    fs.writeFileSync("/Users/mac/Codes/golang/go-connector/test_proto/target.json", output );
    // do something with output
});