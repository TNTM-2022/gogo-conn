const Pomelo = require('./pomelo-client');
const debug = require('debug');

const reqRouterDebugger = debug('request:router');

module.exports = function () {
    const pomelo = Pomelo.create();

    pomelo.initAsync = function (params, options) {
        return new Promise((resolve, reject) => {
            pomelo.removeAllListeners('error');
            pomelo.removeAllListeners('io-error');
            pomelo.init(params, () => {
                resolve();
            }, options);
            pomelo.once('error', e => {
                reject(e);
                pomelo && pomelo.disconnect();
            });
            pomelo.once('io-error', e => {
                reject(e);
                pomelo && pomelo.disconnect();
            });
        });
    };

    pomelo.requestAsync = function (route, msg) {
        return new Promise((resolve, reject) => {
            try {
                reqRouterDebugger(route);
                pomelo.request(route, msg, (err, res) => {
                    if (err) return reject(err);
                    const r = res.data;
                    if (r) {
                        r._code_ = res.code;
                    }

                    resolve(r);
                });
            } catch(e) {
                reject(e);
            }
        });
    };

    // 实现路由分发工作.

    return pomelo;
};
