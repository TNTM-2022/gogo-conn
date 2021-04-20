"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
function default_1(app) {
    return new ChatHandler(app);
}
exports.default = default_1;
class ChatHandler {
    constructor(app) {
        this.app = app;
    }
    async test(msg, session) {
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
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiY2hhdEhhbmRsZXIuanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyIuLi8uLi8uLi8uLi8uLi9hcHAvc2VydmVycy9jaGF0L2hhbmRsZXIvY2hhdEhhbmRsZXIudHMiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6Ijs7QUFJQSxtQkFBd0IsR0FBZ0I7SUFDcEMsT0FBTyxJQUFJLFdBQVcsQ0FBQyxHQUFHLENBQUMsQ0FBQztBQUNoQyxDQUFDO0FBRkQsNEJBRUM7QUFFRCxNQUFhLFdBQVc7SUFDcEIsWUFBb0IsR0FBZ0I7UUFBaEIsUUFBRyxHQUFILEdBQUcsQ0FBYTtJQUNwQyxDQUFDO0lBRUQsS0FBSyxDQUFDLElBQUksQ0FBQyxHQUFHLEVBQUUsT0FBTztRQUNuQixPQUFPO1lBQ0gsSUFBSSxFQUFFLEdBQUc7WUFDVCxJQUFJLEVBQUUsTUFBTTtZQUNaLEdBQUcsRUFBRSxFQUFFO1NBQ1YsQ0FBQztJQUNOLENBQUM7SUFDRDs7Ozs7O09BTUc7SUFDSCxLQUFLLENBQUMsSUFBSSxDQUFDLEdBQXVDLEVBQUUsT0FBdUI7UUFDdkUsSUFBSSxHQUFHLEdBQUcsT0FBTyxDQUFDLEdBQUcsQ0FBQyxLQUFLLENBQUMsQ0FBQztRQUM3QixJQUFJLFFBQVEsR0FBRyxPQUFPLENBQUMsR0FBRyxDQUFDLEtBQUssQ0FBQyxHQUFHLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQztRQUN6QyxJQUFJLGNBQWMsR0FBRyxJQUFJLENBQUMsR0FBRyxDQUFDLEdBQUcsQ0FBQyxnQkFBZ0IsQ0FBQyxDQUFDO1FBQ3BELElBQUksS0FBSyxHQUFHO1lBQ1IsR0FBRyxFQUFFLEdBQUcsQ0FBQyxPQUFPO1lBQ2hCLElBQUksRUFBRSxRQUFRO1lBQ2QsTUFBTSxFQUFFLEdBQUcsQ0FBQyxNQUFNO1NBQ3JCLENBQUM7UUFDRixJQUFJLE9BQU8sR0FBRyxjQUFjLENBQUMsVUFBVSxDQUFDLEdBQUcsRUFBRSxLQUFLLENBQUMsQ0FBQztRQUVwRCwwQkFBMEI7UUFDMUIsSUFBSSxHQUFHLENBQUMsTUFBTSxLQUFLLEdBQUcsRUFBRTtZQUNwQixPQUFPLENBQUMsV0FBVyxDQUFDLFFBQVEsRUFBRSxLQUFLLENBQUMsQ0FBQztTQUN4QztRQUNELDhCQUE4QjthQUN6QjtZQUNELElBQUksSUFBSSxHQUFHLEdBQUcsQ0FBQyxNQUFNLEdBQUcsR0FBRyxHQUFHLEdBQUcsQ0FBQztZQUNsQyxJQUFJLElBQUksR0FBRyxPQUFPLENBQUMsU0FBUyxDQUFDLElBQUksQ0FBQyxDQUFDLEtBQUssQ0FBQyxDQUFDO1lBQzFDLGNBQWMsQ0FBQyxpQkFBaUIsQ0FBQyxRQUFRLEVBQUUsS0FBSyxFQUFFLENBQUM7b0JBQy9DLEdBQUcsRUFBRSxJQUFJO29CQUNULEdBQUcsRUFBRSxJQUFJO2lCQUNaLENBQUMsQ0FBQyxDQUFDO1NBQ1A7SUFDTCxDQUFDO0NBQ0o7QUEzQ0Qsa0NBMkNDIn0=