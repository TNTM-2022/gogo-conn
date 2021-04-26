"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const pinus_1 = require("pinus");
const routeUtil = require("./app/util/routeUtil");
const preload_1 = require("./preload");
const pinus_robot_plugin_1 = require("pinus-robot-plugin");
/**
 *  替换全局Promise
 *  自动解析sourcemap
 *  捕获全局错误
 */
preload_1.preload();
/**
 * Init app for client.
 */
let app = pinus_1.pinus.createApp();
app.set('name', 'chatofpomelo-websocket');
// app configuration
app.configure('production|development', 'connector', function () {
    app.set('connectorConfig', {
        connector: pinus_1.pinus.connectors.hybridconnector,
        heartbeat: 3,
        useDict: true,
        useProtobuf: true
    });
    app.set('serverConfig', {
        reloadHandlers: true,
    });
});
app.configure('production|development', 'gate', function () {
    app.set('connectorConfig', {
        connector: pinus_1.pinus.connectors.hybridconnector,
        useProtobuf: true
    });
});
// app configure
app.configure('production|development', function () {
    // route configures
    app.route('chat', routeUtil.chat);
    // filter configures
    app.filter(new pinus_1.pinus.filters.timeout());
    // 热更新 handler配置
    // app.set('serverConfig',{
    //     reloadHandlers:true,
    // });
    // 热更新 remote 配置
    // app.set('remoteConfig', {
    //     reloadRemotes: true
    // });
});
app.configure('development', function () {
    // enable the system monitor modules
    app.enable('systemMonitor');
});
if (app.isMaster()) {
    app.use(pinus_robot_plugin_1.createRobotPlugin({ scriptFile: __dirname + '/robot/robot.js' }));
}
// start app
app.start();
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiYXBwLmpzIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsiLi4vYXBwLnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiI7O0FBQUEsaUNBQThCO0FBQzlCLGtEQUFtRDtBQUNuRCx1Q0FBb0M7QUFDcEMsMkRBQXVEO0FBRXZEOzs7O0dBSUc7QUFDSCxpQkFBTyxFQUFFLENBQUM7QUFFVjs7R0FFRztBQUNILElBQUksR0FBRyxHQUFHLGFBQUssQ0FBQyxTQUFTLEVBQUUsQ0FBQztBQUM1QixHQUFHLENBQUMsR0FBRyxDQUFDLE1BQU0sRUFBRSx3QkFBd0IsQ0FBQyxDQUFDO0FBRTFDLG9CQUFvQjtBQUNwQixHQUFHLENBQUMsU0FBUyxDQUFDLHdCQUF3QixFQUFFLFdBQVcsRUFBRTtJQUNqRCxHQUFHLENBQUMsR0FBRyxDQUFDLGlCQUFpQixFQUNyQjtRQUNJLFNBQVMsRUFBRSxhQUFLLENBQUMsVUFBVSxDQUFDLGVBQWU7UUFDM0MsU0FBUyxFQUFFLENBQUM7UUFDWixPQUFPLEVBQUUsSUFBSTtRQUNiLFdBQVcsRUFBRSxJQUFJO0tBQ3BCLENBQUMsQ0FBQztJQUVQLEdBQUcsQ0FBQyxHQUFHLENBQUMsY0FBYyxFQUFFO1FBQ3BCLGNBQWMsRUFBRSxJQUFJO0tBQ3ZCLENBQUMsQ0FBQTtBQUNOLENBQUMsQ0FBQyxDQUFDO0FBRUgsR0FBRyxDQUFDLFNBQVMsQ0FBQyx3QkFBd0IsRUFBRSxNQUFNLEVBQUU7SUFDNUMsR0FBRyxDQUFDLEdBQUcsQ0FBQyxpQkFBaUIsRUFDckI7UUFDSSxTQUFTLEVBQUUsYUFBSyxDQUFDLFVBQVUsQ0FBQyxlQUFlO1FBQzNDLFdBQVcsRUFBRSxJQUFJO0tBQ3BCLENBQUMsQ0FBQztBQUNYLENBQUMsQ0FBQyxDQUFDO0FBRUgsZ0JBQWdCO0FBQ2hCLEdBQUcsQ0FBQyxTQUFTLENBQUMsd0JBQXdCLEVBQUU7SUFDcEMsbUJBQW1CO0lBQ25CLEdBQUcsQ0FBQyxLQUFLLENBQUMsTUFBTSxFQUFFLFNBQVMsQ0FBQyxJQUFJLENBQUMsQ0FBQztJQUVsQyxvQkFBb0I7SUFDcEIsR0FBRyxDQUFDLE1BQU0sQ0FBQyxJQUFJLGFBQUssQ0FBQyxPQUFPLENBQUMsT0FBTyxFQUFFLENBQUMsQ0FBQztJQUV4QyxnQkFBZ0I7SUFDaEIsMkJBQTJCO0lBQzNCLDJCQUEyQjtJQUMzQixNQUFNO0lBQ04sZ0JBQWdCO0lBQ2hCLDRCQUE0QjtJQUM1QiwwQkFBMEI7SUFDMUIsTUFBTTtBQUNWLENBQUMsQ0FBQyxDQUFDO0FBRUgsR0FBRyxDQUFDLFNBQVMsQ0FBQyxhQUFhLEVBQUU7SUFDekIsb0NBQW9DO0lBQ3BDLEdBQUcsQ0FBQyxNQUFNLENBQUMsZUFBZSxDQUFDLENBQUM7QUFDaEMsQ0FBQyxDQUFDLENBQUM7QUFFSCxJQUFJLEdBQUcsQ0FBQyxRQUFRLEVBQUUsRUFBRTtJQUNoQixHQUFHLENBQUMsR0FBRyxDQUFDLHNDQUFpQixDQUFDLEVBQUUsVUFBVSxFQUFFLFNBQVMsR0FBRyxpQkFBaUIsRUFBRSxDQUFDLENBQUMsQ0FBQztDQUM3RTtBQUVELFlBQVk7QUFDWixHQUFHLENBQUMsS0FBSyxFQUFFLENBQUMifQ==