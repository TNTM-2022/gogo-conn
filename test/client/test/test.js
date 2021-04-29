const User = require('../user');

void async function () {
    const u = new User('127.0.0.1', 23456);
    // const u = new User('127.0.0.1', 3050);
    console.log(1)
    await u.login();
    console.log(11)
    // await u.talk('connector.entryHandler.enter', {rid: "1999", username: 'username'})
    // for (let i = 0; i < 1000000; i++) {
    //     let s = [];
        // for (let j = 0; j < 100; j++) {
        //     s.push(u.talk('chat.chatHandler.test', {name: 'test'}).catch(console.error))
        // }
        // await Promise.allSettled(s)
    // }
    console.log(await u.talk('chat.chatHandler.test', {name: 'test'}));
    console.log(111)

    await u.exit();
}().catch(console.error);