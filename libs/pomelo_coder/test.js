const protocal = require('/Users/mac/Codes/golang/go-connector/test/pinus/game-server/node_modules/pinus-protocol/dist/lib/protocol.js');

void function () {
    const m = protocal.Message
    const b = Buffer.from("this is a test")
    const cc = m.encode(10, 2, 0,  "test.testHandler.test", b, false)
    console.log(cc.toString('base64'));

    const p = protocal.Package
    const ccc = p.encode(1, cc)
    console.log(ccc.toString('base64'))



    const dp = p.decode(ccc);
    const dm = m.decode(dp.body)
    console.log(dm)
    console.log(dm.body.toString('base64'))
}();