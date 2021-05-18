const User = require('../user');

async function run(n) {
    console.time("run" + n)
    const u = new User('127.0.0.1', 23456);
    // const u = new User('127.0.0.1', 3050);
    u.listen("chat.push",d =>console.log(d.data))
    await u.login();

    // await u.talk('chat.chatHandler.test', {name: 'testname', age:20}).catch(console.error)
    // u.notify("chat.chatHandler.test", {name: 'notify', age:20})
    await u.talk('chat.chatHandler.test', {name: 'test'}).catch(console.error)
    // await u.talk('connector.entryHandler.enter', {rid: "1999", username: 'username'})
    // for (let i = 0; i < 100 / 100; i++) {
    //     console.time("loop" + n)
    //     let s = [];
    //     for (let j = 0; j < 1; j++) {
    //         s.push(u.talk('chat.chatHandler.test', {name: 'test'}).catch())
    //     }
    //     await Promise.allSettled(s)
    //     console.timeEnd("loop" + n)
    // }
    // console.log(await u.talk('chat.chatHandler.test', {name: 'test'}));
    // console.timeEnd("run" + n)

    console.log(u.pomelo.disconnect())
} // ().catch(console.error);

void async function () {
    for (let i = 0; i < 1; i++) {
        run(i)
    }
}()