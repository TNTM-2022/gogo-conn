const EventEmitter = require('events');
const Pomelo = require('./lib/pomelo');
const Koi = require('./koi');
const debug = require('debug');
const Axios = require('axios');

const logoutDebugger = debug("user:logout");
const connDebugger = debug("user:conn");

const serverConfig = {};



class User extends EventEmitter { // 如果整个过程出错, 记得直接 emit error
    /**
     * @param mobile
     * @param password
     * @returns {Promise<void>}
     */
    constructor(host, port) {
        super();

        serverConfig.host = host;
        serverConfig.port = port;

        this.pomelo = Pomelo();
    }

    async talk(p, msg, cb) {
        if (typeof msg === 'function') {
            cb = msg;
            msg = undefined;
        }
        if (!cb) {
            return await new Koi.Package(p, msg, this.pomelo).sendAsync();
        } else {
            return new Koi.Package(p, msg, this.pomelo).send(cb);
        }
    }

    async listen(router, cb) {
        if (cb) {
            this.pomelo.on(router, cb);
        } else {
            return new Promise(resolve => {
                this.pomelo.once(router, resolve);
            });
        }
    }

    async listenOnce(router, cb) {
        if (cb) {
            this.pomelo.once(router, cb);
        } else {
            return new Promise(resolve => {
                this.pomelo.once(router, resolve);
            });
        }
    }

    async login() {
        this.pomelo.on('close', reason => {
            this.emit('logout');
            try {
                this.pomelo.disconnect && this.pomelo.disconnect();
            } catch (e) {
                throw e;
            }
        });


        await this.pomelo.initAsync({
            host: serverConfig.host,
            port: serverConfig.port,
            reconnect: true
        })

        return this;
    }

}

module.exports = User;

// process.on('unhandledRejection', console.error);