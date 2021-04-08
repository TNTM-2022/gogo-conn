const User = require('../user');

void async function () {
    const u = new User('127.0.0.1', 12345);
    // const u = new User('127.0.0.1', 3050);
    console.log(1)
    await u.login();
    console.log(11)
    // await u.talk('connector.entryHandler.enter', {rid: "1999", username: 'username'})
    console.log(await u.talk('chat.chatHandler.test', {name: 'test'}));
    console.log(111)

    await u.exit();
}().catch(console.error);