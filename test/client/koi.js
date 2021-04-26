const debug = require('debug');
const packageDebug = debug('koi:package');

class Package {
    constructor(p, params, socket) {
        this.path = p;
        if (!params) {
            // console.warn('talk content is undefined');
            params = {};
        }
        if (!p) {
            throw new Error('no router')
        }
        this.req = JSON.parse(JSON.stringify(params));
        this.socket = socket;
        packageDebug('router: %o. params: %o', this.path, this.req);
    }

    send(cb) {
        try {
            this.socket.request(this.path, this.req, cb || ((d) => {
                console.warn('%s, %o', this.path, d);
            }));
        } catch (e) {
            console.error(e);
            if (e.message === 'not opened') {
                this.emit('close', e);
            }
        }
    }

    async sendAsync() {
        let r;
        r = await this.socket.requestAsync(this.path, this.req);
        return r;
    }
}

exports.Package = Package;
