import {Application, BackendSession} from 'pinus';

export default function (app: Application) {
    return new ChatHandler(app);
}

export class ChatHandler {
    constructor(private app: Application) {
    }

    async test(msg, session: BackendSession) {
       if (false) {
            console.log('push message 1')
            let channelService = this.app.get('channelService');
            channelService.createChannel("test123")
            // channelService.broadcast("connector", "broadcast.test", {isPush: true})
            let channel = channelService.getChannel("test123", false)
            if (!channel.getMember(session.uid)) {
                channel.add(session.uid, session.frontendId)
            }
            await channel.apushMessage("chat.push", {event1: "push.push", is_broad: true}, {opts: true})
            console.log('push message 2')
            await channel.destroy()
        }

        {
            console.log('session set 1', session.uid)
            console.log(typeof session.get('a'), typeof session.get('b'), typeof session.get('c'), session.settings)
            session.set("a", "s1234")
            session.set("b", 1234)
            session.set("c", false)
            await session.apush("a")
            await session.apushAll()
            console.log('session set 2', session.uid)
        }

        console.log(JSON.stringify(msg), msg.name)

        return {
            code: 200,
            user: msg.name,
            msg: "msg" + msg + session.uid
        };
    }

    /**
     * Send messages to users
     *
     * @param {Object} msg message from client
     * @param {Object} session
     *
     */
    async send(msg: { content: string, target: string }, session: BackendSession) {
        let rid = session.get('rid');
        let username = session.uid.split('*')[0];
        let channelService = this.app.get('channelService');
        let param = {
            msg: msg.content,
            from: username,
            target: msg.target
        };
        let channel = channelService.getChannel(rid, false);

        // the target is all users
        if (msg.target === '*') {
            channel.pushMessage('onChat', param);
        }
        // the target is specific user
        else {
            let tuid = msg.target + '*' + rid;
            let tsid = channel.getMember(tuid)['sid'];
            channelService.pushMessageByUids('onChat', param, [{
                uid: tuid,
                sid: tsid
            }]);
        }
    }
}