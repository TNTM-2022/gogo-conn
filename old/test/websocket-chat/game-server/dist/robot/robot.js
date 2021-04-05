"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const pinus_robot_plugin_1 = require("pinus-robot-plugin");
class Robot {
    constructor(actor) {
        this.actor = actor;
        this.openid = String(Math.round(Math.random() * 1000));
        this.pinusClient = new pinus_robot_plugin_1.PinusWSClient();
    }
    connectGate() {
        let host = '127.0.0.1';
        let port = '3014';
        this.pinusClient.on(pinus_robot_plugin_1.PinusWSClientEvent.EVENT_IO_ERROR, (event) => {
            // 错误处理
            console.error('error', event);
        });
        this.pinusClient.on(pinus_robot_plugin_1.PinusWSClientEvent.EVENT_CLOSE, function (event) {
            // 关闭处理
            console.error('close', event);
        });
        this.pinusClient.on(pinus_robot_plugin_1.PinusWSClientEvent.EVENT_HEART_BEAT_TIMEOUT, function (event) {
            // 心跳timeout
            console.error('heart beat timeout', event);
        });
        this.pinusClient.on(pinus_robot_plugin_1.PinusWSClientEvent.EVENT_KICK, function (event) {
            // 踢出
            console.error('kick', event);
        });
        // this.actor.emit("incr" , "gateConnReq");
        this.actor.emit('start', 'gateConn', this.actor.id);
        this.pinusClient.init({
            host: host,
            port: port
        }, () => {
            this.actor.emit('end', 'gateConn', this.actor.id);
            // 连接成功执行函数
            console.log('gate连接成功');
            this.gateQuery();
        });
    }
    gateQuery() {
        // this.actor.emit("incr" , "gateQueryReq");
        this.actor.emit('start', 'gateQuery', this.actor.id);
        this.pinusClient.request('gate.gateHandler.queryEntry', { uid: this.openid }, (result) => {
            // 消息回调
            // console.log("gate返回",JSON.stringify(result));
            this.actor.emit('end', 'gateQuery', this.actor.id);
            this.pinusClient.disconnect();
            this.connectToConnector(result);
        });
    }
    connectToConnector(result) {
        // this.actor.emit("incr" , "loginConnReq");
        this.actor.emit('start', 'loginConn', this.actor.id);
        this.pinusClient.init({
            host: result.host,
            port: result.port
        }, () => {
            this.actor.emit('end', 'loginConn', this.actor.id);
            // 连接成功执行函数
            console.log('connector连接成功');
            this.loginQuery({ rid: this.actor.id.toString(), username: this.actor.id.toString() });
        });
    }
    loginQuery(result) {
        // this.actor.emit("incr" , "loginQueryReq");
        this.actor.emit('start', 'loginQuery', this.actor.id);
        this.pinusClient.request('connector.entryHandler.enter', result, (ret) => {
            // 消息回调
            this.actor.emit('end', 'loginQuery', this.actor.id);
            console.log('connector返回', JSON.stringify(result));
            setTimeout(() => this.loginQuery(result), Math.random() * 5000 + 1000);
        });
    }
}
exports.Robot = Robot;
function default_1(actor) {
    let client = new Robot(actor);
    client.connectGate();
    return client;
}
exports.default = default_1;
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoicm9ib3QuanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyIuLi8uLi9yb2JvdC9yb2JvdC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOztBQUVBLDJEQUFzRTtBQUd0RSxNQUFhLEtBQUs7SUFDZCxZQUFvQixLQUFZO1FBQVosVUFBSyxHQUFMLEtBQUssQ0FBTztRQUloQyxXQUFNLEdBQUcsTUFBTSxDQUFDLElBQUksQ0FBQyxLQUFLLENBQUMsSUFBSSxDQUFDLE1BQU0sRUFBRSxHQUFHLElBQUksQ0FBQyxDQUFDLENBQUM7UUFFbEQsZ0JBQVcsR0FBRyxJQUFJLGtDQUFhLEVBQUUsQ0FBQztJQUpsQyxDQUFDO0lBTU0sV0FBVztRQUVkLElBQUksSUFBSSxHQUFHLFdBQVcsQ0FBQztRQUN2QixJQUFJLElBQUksR0FBRyxNQUFNLENBQUM7UUFDbEIsSUFBSSxDQUFDLFdBQVcsQ0FBQyxFQUFFLENBQUMsdUNBQWtCLENBQUMsY0FBYyxFQUFFLENBQUMsS0FBSyxFQUFFLEVBQUU7WUFDN0QsT0FBTztZQUNQLE9BQU8sQ0FBQyxLQUFLLENBQUMsT0FBTyxFQUFFLEtBQUssQ0FBQyxDQUFDO1FBQ2xDLENBQUMsQ0FBQyxDQUFDO1FBQ0gsSUFBSSxDQUFDLFdBQVcsQ0FBQyxFQUFFLENBQUMsdUNBQWtCLENBQUMsV0FBVyxFQUFFLFVBQVMsS0FBSztZQUM5RCxPQUFPO1lBQ1AsT0FBTyxDQUFDLEtBQUssQ0FBQyxPQUFPLEVBQUUsS0FBSyxDQUFDLENBQUM7UUFDbEMsQ0FBQyxDQUFDLENBQUM7UUFDSCxJQUFJLENBQUMsV0FBVyxDQUFDLEVBQUUsQ0FBQyx1Q0FBa0IsQ0FBQyx3QkFBd0IsRUFBRSxVQUFTLEtBQUs7WUFDM0UsWUFBWTtZQUNaLE9BQU8sQ0FBQyxLQUFLLENBQUMsb0JBQW9CLEVBQUUsS0FBSyxDQUFDLENBQUM7UUFDL0MsQ0FBQyxDQUFDLENBQUM7UUFDSCxJQUFJLENBQUMsV0FBVyxDQUFDLEVBQUUsQ0FBQyx1Q0FBa0IsQ0FBQyxVQUFV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