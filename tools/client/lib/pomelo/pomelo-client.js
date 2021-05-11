const WebSocket = require('ws');
const Protocol = require('pomelo-protocol');
const Package = Protocol.Package;
const Message = Protocol.Message;
const EventEmitter = require('events');
const { inherits } = require('util');
const debug = require('debug');
const pomeloConnDebug = debug('pomelo:conn');
const protoCoder = debug('-proto-');
const jsonCoder = debug('- json-');
const protobufjs = require('protobufjs');
let protobuf; // pomelo-protobuf

const JS_WS_CLIENT_TYPE = 'js-websocket';
const JS_WS_CLIENT_VERSION = '0.0.1';

function Pomelo() {
    EventEmitter.call(this);
}
inherits(Pomelo, EventEmitter);

const RES_OK = 200;
const RES_OLD_CLIENT = 501;

// const serverProto = require('../../../../game-server/config/serverProtos');
// const clientProto = require('../../../../game-server/config/clientProtos');
// const pushProto = require('../../../../game-server/config/pushProtos');

const decodeIO_decoder = protobufjs.Root.fromJSON(require("./target_res.json"));
const decodeIO_encoder = protobufjs.Root.fromJSON(require("./target_req.json"));
// const decodeIO_encoder = null;//= protobufjs.Root.fromJSON(clientProto);
const push_decoder = protobufjs.Root.fromJSON(require("./target_res.json"));

let _uuid = 1000;
function create() {
    const pomelo = new Pomelo();
    var socket = null;
    var reqId = 0;
    var callbacks = {};
    var handlers = {};
    var routeMap = {};

    var heartbeatInterval = 5000;
    var heartbeatTimeout = heartbeatInterval * 2;
    var nextHeartbeatTimeout = 0;
    var gapThreshold = 100; // heartbeat gap threshold
    var heartbeatId = null;
    var heartbeatTimeoutId = null;

    var handshakeCallback = null;

    var handshakeBuffer = {
        'sys': {
            type: JS_WS_CLIENT_TYPE,
            version: JS_WS_CLIENT_VERSION,
            protoVersion: -1
        },
        'user': {
        }
    };

    var initCallback = null;

    pomelo.init = function (params, cb, option) {
        pomelo.params = params;
        params.debug = true;
        initCallback = cb;
        var host = params.host;
        var port = params.port;

        var url = 'ws://' + host;
        if (port) {
            url += ':' + port;
        }
        if (!params.type) {
            pomeloConnDebug('init websocket');
            handshakeBuffer.user = params.user;
            handshakeCallback = params.handshakeCallback;
            this.initWebSocket(url, cb, option);
        }
    };

    pomelo.initWebSocket = function (url, cb, options) {
        pomeloConnDebug(url);
        var onopen = function (event) {
            // console.log(event);
            pomeloConnDebug('[pomeloclient.init] websocket connected!');
            var obj = Package.encode(Package.TYPE_HANDSHAKE, Protocol.strencode(JSON.stringify(handshakeBuffer)));
            send(obj);
        };
        var onmessage = function (event) {
            // console.log("event.data", event.data)
            const a = processPackage(Package.decode(event.data), cb);
            // console.log("event.data", Package.decode(event.data)?.body?.toString())
            // new package arrived, update the heartbeat timeout
            if (heartbeatTimeout) {
                nextHeartbeatTimeout = Date.now() + heartbeatTimeout;
            }
        };
        var onerror = function (event) {
            pomelo.emit('io-error', event);
            pomeloConnDebug('socket error %j ', event);
        };
        var onclose = function (event) {
            pomelo.emit('close', event);

            pomeloConnDebug('socket close %j ', event.reason, pomelo.listeners('close'));
        };
        socket = new WebSocket(`${url}/ws`, options);
        socket.binaryType = 'arraybuffer';
        socket.onopen = onopen;
        socket.onmessage = onmessage;
        socket.onerror = onerror;
        socket.onclose = onclose;
    };

    pomelo.disconnect = function () {
        if (socket) {
            // if(socket.disconnect) socket.disconnect();
            if (socket.close) socket.close();
            pomeloConnDebug('disconnect');
            socket = null;
        }

        if (heartbeatId) {
            clearTimeout(heartbeatId);
            heartbeatId = null;
        }
        if (heartbeatTimeoutId) {
            clearTimeout(heartbeatTimeoutId);
            heartbeatTimeoutId = null;
        }
    };

    pomelo.request = function (route, msg, cb) {
        msg = msg || {};
        route = route || msg.route;
        if (!route) {
            console.error('fail to send request without route.');
            return;
        }

        reqId++;
        sendMessage(reqId, route, msg);

        callbacks[reqId] = cb;
        routeMap[reqId] = route;
    };

    pomelo.notify = function (route, msg) {
        msg = msg || {};
        sendMessage(0, route, msg);
    };

    var sendMessage = function (reqId, route, msg) {
        var type = reqId ? Message.TYPE_REQUEST : Message.TYPE_NOTIFY;

        //compress message by protobuf
        // var protos = !!pomelo.data.protos?pomelo.data.protos.client:{};

        // if(protobuf && protos[route]) {
        //     msg = protobuf.encode(route, msg);
        // } else
        if (decodeIO_encoder && decodeIO_encoder.lookup(route)) {
            protoCoder('send', route);
            const um = decodeIO_encoder.lookupType(route);
            const buf = um.encode(um.create(msg)).finish();
            msg = buf;
            // msg = buf.encodeNB();
        } else {
            jsonCoder('send', route);
            msg = Protocol.strencode(JSON.stringify(msg));
        }

        var compressRoute = 0;
        if (pomelo.dict && pomelo.dict[route]) {
            route = pomelo.dict[route];
            compressRoute = 1;
        }

        msg = Message.encode(reqId, type, compressRoute, route, msg);
        var packet = Package.encode(Package.TYPE_DATA, msg);
        send(packet);
    };


    var _host = "";
    var _port = "";
    var _token = "";

    /*
    var send = function(packet){
      if (!!socket) {
        socket.send(packet.buffer || packet,{binary: true, mask: true});
      } else {
        setTimeout(function() {
          entry(_host, _port, _token, function() {console.log('Socket is null. ReEntry!')});
        }, 3000);
      }
    };
    */

    var send = function (packet) {
        if (!!socket) {
            //8192!? problem fix packet.buffer ||
            try {
                socket.send(packet, { binary: true, mask: true });
            } catch (e) {
                console.error(e)
            }
        }
    };


    var handler = {};

    var heartbeat = function (data) {
        var obj = Package.encode(Package.TYPE_HEARTBEAT);
        if (heartbeatTimeoutId) {
            clearTimeout(heartbeatTimeoutId);
            heartbeatTimeoutId = null;
        }

        if (heartbeatId) {
            // already in a heartbeat interval
            return;
        }

        heartbeatId = setTimeout(function () {
            heartbeatId = null;
            send(obj);

            nextHeartbeatTimeout = Date.now() + heartbeatTimeout;
            heartbeatTimeoutId = setTimeout(heartbeatTimeoutCb, heartbeatTimeout);
        }, heartbeatInterval);
    };

    var heartbeatTimeoutCb = function () {
        var gap = nextHeartbeatTimeout - Date.now();
        if (gap > gapThreshold) {
            heartbeatTimeoutId = setTimeout(heartbeatTimeoutCb, gap);
        } else {
            console.error('server heartbeat timeout');
            pomelo.emit('heartbeat timeout');
            pomelo.disconnect();
        }
    };

    var handshake = function (data) {
        data = JSON.parse(Protocol.strdecode(data));
        // console.log(246, data);
        if (data.code === RES_OLD_CLIENT) {
            pomelo.emit('error', 'client version not fullfill');
            return;
        }

        if (data.code !== RES_OK) {
            pomelo.emit('error', 'handshake fail');
            return;
        }
        handshakeInit(data);

        var obj = Package.encode(Package.TYPE_HANDSHAKE_ACK);
        console.log('hand shake ack', obj, obj.toString())
        try {
            send(obj);
        } catch (e) {
            console.error(e)
        }
        if (initCallback) {
            initCallback(socket);
            initCallback = null;
        }
    };

    var onData = function (data) {
        //probuff decode
        // console.log('on data', data.toString())
        var msg = Message.decode(data);

        if (msg.id > 0) {
            msg.route = routeMap[msg.id];
            delete routeMap[msg.id];
            if (!msg.route) {
                return;
            }
        }
        if (!(msg.id > 0) && msg.route.startsWith('_.')) {
            msg.route = msg.route.slice(2);
        }
        console.log(msg)
        msg.body = deCompose(msg);

        processMessage(pomelo, msg);
    };

    var onKick = function (data) {
        data = JSON.parse(Protocol.strdecode(data));
        pomelo.emit('kick', data);
        pomelo.emit('onKick', data);
    };

    handlers[Package.TYPE_HANDSHAKE] = handshake;
    handlers[Package.TYPE_HEARTBEAT] = heartbeat;
    handlers[Package.TYPE_DATA] = onData;
    handlers[Package.TYPE_KICK] = onKick;
    var processPackage = function (msg) {
        handlers[msg.type](msg.body);
    };

    var processMessage = function (pomelo, msg) {
        if (!msg.id) {
            // server push message
            if (msg.route === 'message') {
                if (msg.body.event) {
                    pomelo.emit(`event.${msg.body.event}`, { route: `event.${msg.body.event}`, code: 0, data: msg.body, msg });
                    pomelo.emit(msg.body.event, { route: `event.${msg.body.event}`, code: 0, data: msg.body, msg });
                } else {
                    pomelo.emit(msg.route, { route: `event.${msg.body.event}`, code: 0, data: msg.body, msg });
                }
            }
            else {
                const data = msg.body;
                data.data = data;
                data.msg = msg
                pomelo.emit(msg.route, data);
            }
            return;
        }

        //if have a id then find the callback function with the request
        var cb = callbacks[msg.id];

        delete callbacks[msg.id];
        if (typeof cb !== 'function') {
            //     return pomelo.emit('message', Object.assign({route: msg.route}, msg.body));
            pomelo.emit('message', Object.assign({ route: msg.route, msg }, msg.body));
        }

        var err = (msg.body && msg.body.code) ? {
            code: msg.body.code,
            message: msg.body.message,
            data: msg.body.data,
            msg: msg.body,
        } : null;
        return cb(err, msg.body);
    };


    var processMessageBatch = function (pomelo, msgs) {
        for (var i = 0, l = msgs.length; i < l; i++) {
            processMessage(pomelo, msgs[i]);
        }
    };

    var deCompose = function (msg) {
        var isPush = !(msg.id > 0);
        // var protos = !!pomelo.data.protos ? pomelo.data.protos.server : {};
        if (!pomelo.data || !pomelo.data.abbrs) {
            console.log(msg)
        }
        var abbrs = pomelo.data?.abbrs || {};
        var route = msg.route;

        try {
            //Decompose route from dict
            if (msg.compressRoute) {
                if (!abbrs[route]) {
                    console.error('illegal msg!');
                    return {};
                }

                route = msg.route = abbrs[route];
            }
            if (isPush && push_decoder && push_decoder.lookup(route)) {
                protoCoder(isPush ? 'push' : 'resp', route);
                const um = push_decoder.lookupType(route);
                const d = um.decode(msg.body);
                return um.toObject(d, {
                    longs: String,
                    enums: String,
                    bytes: String,
                    defaults: true,
                    arrays: true,
                    objects: true,
                    oneof: true
                });
            } else if (!isPush && decodeIO_decoder && decodeIO_decoder.lookup(route)) {
                protoCoder(isPush ? 'push' : 'resp', route);
                const um = decodeIO_decoder.lookupType(route);
                const d = um.decode(msg.body);
                const o = um.toObject(d, {
                    longs: String,
                    enums: String,
                    bytes: String,
                    defaults: true,
                    arrays: true,
                    objects: true,
                    oneof: true
                });
                // console.log(msg.body, (msg.body).toString(), o)
                return o;
            } else {
                jsonCoder(isPush ? 'push' : 'resp', route, msg.body);

                return JSON.parse(Protocol.strdecode(msg.body));
            }

        } catch (ex) {
            console.error('route, body = ' + route + ", " + msg.body, ex);
        }

        return msg;
    };

    var handshakeInit = function (data) {
        if (data.sys && data.sys.heartbeat) {
            heartbeatInterval = data.sys.heartbeat * 1000;   // heartbeat interval
            heartbeatTimeout = heartbeatInterval * 2;        // max heartbeat timeout
        } else {
            heartbeatInterval = 0;
            heartbeatTimeout = 0;
        }

        initData(data);

        if (typeof handshakeCallback === 'function') {
            handshakeCallback(data.user);
        }
        console.log('handshakeInited')
    };

    //Initilize data used in pomelo client
    var initData = function (data) {
        if (!data || !data.sys) {
            return;
        }
        pomelo.data = pomelo.data || {};
        var dict = data.sys.dict;
        var protos = data.sys.protos;
        console.log("intDate")
        //Init compress dict
        if (!!dict) {
            pomelo.data.dict = dict;
            pomelo.data.abbrs = {};

            for (var route in dict) {
                pomelo.data.abbrs[dict[route]] = route;
            }
        }
        // console.log('protos give =>', !!protos);
        //Init protobuf protos
        // if(!!protos){
        //     pomelo.data.protos = {
        //         server : protos.server || {},
        //         client : protos.client || {},
        //         push: protos.push || {},
        //     };
        //     if(!!protobuf){
        //         protobuf.init({encoderProtos: protos.client, decoderProtos: protos.server});
        //     }
        //     if(!!protobufjs) {
        //         decodeIO_encoder = protobufjs.Root.fromJSON(protos.client);
        //         decodeIO_decoder = protobufjs.Root.fromJSON(protos.server);
        //         push_decoder = protobufjs.Root.fromJSON(protos.push);
        //     }
        // }
    };
    return pomelo;
}
module.exports.create = create;
