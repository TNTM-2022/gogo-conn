const EventEmitter = require('events');
const User = require('./user');
const redis = require('redis');
const MongoClient = require('mongodb').MongoClient;
const bluebird = require('bluebird');
const Koi = require('./koi');
const debug = require('debug');

const config = require('config');

bluebird.promisifyAll(redis.RedisClient.prototype);
bluebird.promisifyAll(redis.Multi.prototype);

const logoutDebugger = debug("user:logout");
const connDebugger = debug("user:conn");

const serverConfig = {};
const dbConn = {};

class UserEx extends User { // 如果整个过程出错, 记得直接 emit error
    constructor(...loginInfo) {
        super(...loginInfo);
        if (config.has('pomelo.host') && config.has('pomelo.port')) {
            User.config(config.get('pomelo.host'), config.get('pomelo.port'));
        }
    }

    static async close() {
        if (dbConn.mongo) {
            await dbConn.mongo.close();

        }
        if (dbConn.redis) {
            await dbConn.redis.quit();
        }
    }

    static async getAndLogin(count) {
        await getDBConnection();

        const u = await dbConn.mongo.db('fish').collection('users').find({loginToken: /.+/}).limit(count).toArray();

        const users = [];
        const _us = [];
        for (const user of u) {
            const u = new this(user);
            users.push(u);
            _us.push(u.login());
        }
        await Promise.all(_us);

        return users;
    }

    static redisConn(){
        if(!dbConn.redis)
            throw 'get redis after ?'
        return dbConn.redis;
    }


    static getRoomMasterGold(...p) {
        return getRoomMasterGold (...p);
    }

    static async getPumpGold(...p) {
        return getPumpGold(...p)
    }

    static getUserGold(...p) {
        return getUserGold(...p);
    }

    async getUserGold() {
        return getUserGold(this.user.uid);
    }
}

async function getRoomMasterGold(gameId, roomId) {
    await getDBConnection();

    let uid = (`${gameId}`.padEnd(3, '0') + `${roomId}`.padStart(6, '0')) | 0;
    return await getUserGold(uid);
}

async function getPumpGold(gameId) {
    await getDBConnection();

    let uid = (`${gameId}`.padEnd(3, '0') + ''.padStart(6, '0')) | 0;
    return await getUserGold(uid);
}

async function getUserGold(uid) {
    await getDBConnection();

    let user = await dbConn.mongo.db('fish').collection('users').findOne({uid});
    return user ? user.gold : 0;
}

async function getDBConnection() {
    if (!Object.keys(serverConfig).length) {
        await createConnection();
    }
    return dbConn;
}

async function createConnection() {
    dbConn.connected = true;
    dbConn.mongo = await MongoClient.connect(config.get('mongo'));
    dbConn.redis = redis.createClient({
        db: config.get('redis.db'),
        port: config.get('redis.port'),
        host: config.get('redis.host'),
    });
    if (config.has('redis.password')) {
        await dbConn.redis.auth(config.get('redis.password'));
    }
}

module.exports = UserEx;

// process.on('unhandledRejection', console.error);