"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ChatHandler = void 0;
function default_1(app) {
    return new ChatHandler(app);
}
exports.default = default_1;
class ChatHandler {
    constructor(app) {
        this.app = app;
    }
    async test(msg, session) {
        {
            console.log('push message 1');
            let channelService = this.app.get('channelService');
            channelService.createChannel("test123");
            // channelService.broadcast("connector", "broadcast.test", {isPush: true})
            let channel = channelService.getChannel("test123", false);
            if (!channel.getMember(session.uid)) {
                channel.add(session.uid, session.frontendId);
            }
            await channel.apushMessage("push.push", { event1: "push.push", is_broad: true }, { opts: true });
            console.log('push message 2');
            await channel.destroy();
        }
        {
            console.log('session set 1');
            console.log(typeof session.get('a'), typeof session.get('b'), typeof session.get('c'), session.settings);
            session.set("a", "s1234");
            session.set("b", 1234);
            session.set("c", false);
            await session.apush("a");
            await session.apushAll();
            console.log('session set 2');
        }
        return {
            code: 200,
            name: 'test',
            age: 10
        };
    }
    /**
     * Send messages to users
     *
     * @param {Object} msg message from client
     * @param {Object} session
     *
     */
    async send(msg, session) {
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
exports.ChatHandler = ChatHandler;
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiY2hhdEhhbmRsZXIuanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyIuLi8uLi8uLi8uLi8uLi9hcHAvc2VydmVycy9jaGF0L2hhbmRsZXIvY2hhdEhhbmRsZXIudHMiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6Ijs7O0FBRUEsbUJBQXlCLEdBQWdCO0lBQ3JDLE9BQU8sSUFBSSxXQUFXLENBQUMsR0FBRyxDQUFDLENBQUM7QUFDaEMsQ0FBQztBQUZELDRCQUVDO0FBRUQsTUFBYSxXQUFXO0lBQ3BCLFlBQW9CLEdBQWdCO1FBQWhCLFFBQUcsR0FBSCxHQUFHLENBQWE7SUFDcEMsQ0FBQztJQUVELEtBQUssQ0FBQyxJQUFJLENBQUMsR0FBRyxFQUFFLE9BQXVCO1FBQ25DO1lBQ0ksT0FBTyxDQUFDLEdBQUcsQ0FBQyxnQkFBZ0IsQ0FBQyxDQUFBO1lBQzdCLElBQUksY0FBYyxHQUFHLElBQUksQ0FBQyxHQUFHLENBQUMsR0FBRyxDQUFDLGdCQUFnQixDQUFDLENBQUM7WUFDcEQsY0FBYyxDQUFDLGFBQWEsQ0FBQyxTQUFTLENBQUMsQ0FBQTtZQUN2QywwRUFBMEU7WUFDMUUsSUFBSSxPQUFPLEdBQUcsY0FBYyxDQUFDLFVBQVUsQ0FBQyxTQUFTLEVBQUUsS0FBSyxDQUFDLENBQUE7WUFDekQsSUFBSSxDQUFDLE9BQU8sQ0FBQyxTQUFTLENBQUMsT0FBTyxDQUFDLEdBQUcsQ0FBQyxFQUFFO2dCQUNqQyxPQUFPLENBQUMsR0FBRyxDQUFDLE9BQU8sQ0FBQyxHQUFHLEVBQUUsT0FBTyxDQUFDLFVBQVUsQ0FBQyxDQUFBO2FBQy9DO1lBQ0QsTUFBTSxPQUFPLENBQUMsWUFBWSxDQUFDLFdBQVcsRUFBRSxFQUFDLE1BQU0sRUFBRSxXQUFXLEVBQUUsUUFBUSxFQUFFLElBQUksRUFBQyxFQUFFLEVBQUMsSUFBSSxFQUFFLElBQUksRUFBQyxDQUFDLENBQUE7WUFDNUYsT0FBTyxDQUFDLEdBQUcsQ0FBQyxnQkFBZ0IsQ0FBQyxDQUFBO1lBQzdCLE1BQU0sT0FBTyxDQUFDLE9BQU8sRUFBRSxDQUFBO1NBQzFCO1FBRUQ7WUFDSSxPQUFPLENBQUMsR0FBRyxDQUFDLGVBQWUsQ0FBQyxDQUFBO1lBQzVCLE9BQU8sQ0FBQyxHQUFHLENBQUMsT0FBTyxPQUFPLENBQUMsR0FBRyxDQUFDLEdBQUcsQ0FBQyxFQUFFLE9BQU8sT0FBTyxDQUFDLEdBQUcsQ0FBQyxHQUFHLENBQUMsRUFBRSxPQUFPLE9BQU8sQ0FBQyxHQUFHLENBQUMsR0FBRyxDQUFDLEVBQUUsT0FBTyxDQUFDLFFBQVEsQ0FBQyxDQUFBO1lBQ3hHLE9BQU8sQ0FBQyxHQUFHLENBQUMsR0FBRyxFQUFFLE9BQU8sQ0FBQyxDQUFBO1lBQ3pCLE9BQU8sQ0FBQyxHQUFHLENBQUMsR0FBRyxFQUFFLElBQUksQ0FBQyxDQUFBO1lBQ3RCLE9BQU8sQ0FBQyxHQUFHLENBQUMsR0FBRyxFQUFFLEtBQUssQ0FBQyxDQUFBO1lBQ3ZCLE1BQU0sT0FBTyxDQUFDLEtBQUssQ0FBQyxHQUFHLENBQUMsQ0FBQTtZQUN4QixNQUFNLE9BQU8sQ0FBQyxRQUFRLEVBQUUsQ0FBQTtZQUN4QixPQUFPLENBQUMsR0FBRyxDQUFDLGVBQWUsQ0FBQyxDQUFBO1NBQy9CO1FBRUQsT0FBTztZQUNILElBQUksRUFBRSxHQUFHO1lBQ1QsSUFBSSxFQUFFLE1BQU07WUFDWixHQUFHLEVBQUUsRUFBRTtTQUNWLENBQUM7SUFDTixDQUFDO0lBRUQ7Ozs7OztPQU1HO0lBQ0gsS0FBSyxDQUFDLElBQUksQ0FBQyxHQUF3QyxFQUFFLE9BQXVCO1FBQ3hFLElBQUksR0FBRyxHQUFHLE9BQU8sQ0FBQyxHQUFHLENBQUMsS0FBSyxDQUFDLENBQUM7UUFDN0IsSUFBSSxRQUFRLEdBQUcsT0FBTyxDQUFDLEdBQUcsQ0FBQyxLQUFLLENBQUMsR0FBRyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUM7UUFDekMsSUFBSSxjQUFjLEdBQUcsSUFBSSxDQUFDLEdBQUcsQ0FBQyxHQUFHLENBQUMsZ0JBQWdCLENBQUMsQ0FBQztRQUNwRCxJQUFJLEtBQUssR0FBRztZQUNSLEdBQUcsRUFBRSxHQUFHLENBQUMsT0FBTztZQUNoQixJQUFJLEVBQUUsUUFBUTtZQUNkLE1BQU0sRUFBRSxHQUFHLENBQUMsTUFBTTtTQUNyQixDQUFDO1FBQ0YsSUFBSSxPQUFPLEdBQUcsY0FBYyxDQUFDLFVBQVUsQ0FBQyxHQUFHLEVBQUUsS0FBSyxDQUFDLENBQUM7UUFFcEQsMEJBQTBCO1FBQzFCLElBQUksR0FBRyxDQUFDLE1BQU0sS0FBSyxHQUFHLEVBQUU7WUFDcEIsT0FBTyxDQUFDLFdBQVcsQ0FBQyxRQUFRLEVBQUUsS0FBSyxDQUFDLENBQUM7U0FDeEM7UUFDRCw4QkFBOEI7YUFDekI7WUFDRCxJQUFJLElBQUksR0FBRyxHQUFHLENBQUMsTUFBTSxHQUFHLEdBQUcsR0FBRyxHQUFHLENBQUM7WUFDbEMsSUFBSSxJQUFJLEdBQUcsT0FBTyxDQUFDLFNBQVMsQ0FBQyxJQUFJLENBQUMsQ0FBQyxLQUFLLENBQUMsQ0FBQztZQUMxQyxjQUFjLENBQUMsaUJBQWlCLENBQUMsUUFBUSxFQUFFLEtBQUssRUFBRSxDQUFDO29CQUMvQyxHQUFHLEVBQUUsSUFBSTtvQkFDVCxHQUFHLEVBQUUsSUFBSTtpQkFDWixDQUFDLENBQUMsQ0FBQztTQUNQO0lBQ0wsQ0FBQztDQUNKO0FBckVELGtDQXFFQyJ9