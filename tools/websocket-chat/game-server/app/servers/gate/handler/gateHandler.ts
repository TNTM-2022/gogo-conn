import { dispatch } from '../../../util/dispatcher';
import { Application , BackendSession} from 'pinus';

export default function (app: Application) {
    return new GateHandler(app);
}

export class GateHandler {
    constructor(private app: Application) {
        setInterval(() => {
            console.log(app.get('onlineUser'))
        }, 2000)
    }

    async test (msg: {uid: string}, session: BackendSession) {
        return {
            code: 200,
            name: 'test',
            age: 10
        }
    }
    /**
     * Gate handler that dispatch user to connectors.
     *
     * @param {Object} msg message from client
     * @param {Object} session
     *
     */
    // async queryEntry(msg: {uid: string}, session: BackendSession) {
    //     let uid = msg.uid;
    //     if (!uid) {
    //         return {
    //             code: 500
    //         };
    //     }
    //     // get all connectors
    //     let connectors = this.app.getServersByType('connector');
    //     if (!connectors || connectors.length === 0) {
    //         return {
    //             code: 500
    //         };
    //     }
    //     // select connector
    //     let res = dispatch(uid, connectors);
    //     return {
    //         code: 200,
    //         host: res.host,
    //         port: res.clientPort
    //     };
    // }
}