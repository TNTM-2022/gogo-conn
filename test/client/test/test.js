const User = require('../user');

void async function () {
    const u = new User('127.0.0.1', 12345);
    await u.login();
    console.log(await u.talk('user.userHandler.getAvatarBorderList', {name: 'test'}));
    // await u.exit();
}().catch(console.error);