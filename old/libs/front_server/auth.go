package front_server

import (
	"fmt"
	"github.com/go-redis/redis"
	"go-connector/interfaces"
)

var rdb *redis.Client
var scriptHash = make(map[string]*redis.Script)
var userToken = "token-%d"

func init() {
	rdb = redis.NewClient(&redis.Options{
		//Addr:     *conf.RedisAddr,
		//Password: *conf.RedisPwd, // no password set
		//DB:       *conf.RedisDB,  // use default DB
	})

	pong, err := rdb.Ping().Result()
	fmt.Println(pong, err)
	script := redis.NewScript(`
local sid = ARGV[1];
local token = ARGV[2];
local serverId = ARGV[3];

local tokenUidKey = KEYS[1];

local lastInfo = redis.pcall('HMGET', tokenUidKey, 'sid', 'token', 'info');

local _sid = lastInfo[1];
local _token = lastInfo[2];
local _info = lastInfo[3];

if _token == false or token == false or _token ~= token then -- 登陆不匹配
    --return cjson.encode({ Ok= false, token=token, _token=_token });
    return cjson.encode({ Ok= false, token=token, _token=_token });
end

local _frontSid = ' ';
if _sid ~= false then
    local f = 'front_' .. _sid;
    local _f = redis.pcall('HGET', tokenUidKey, f);
    if _f ~= false then
        _frontSid = _f;
        redis.pcall('HDEL', tokenUidKey, f);
    end
end


redis.pcall('HMSET', tokenUidKey, 'sid', sid, 'front_' .. sid, serverId);
redis.pcall('EXPIRE', tokenUidKey, 60 * 60 * 24 * 10);
--return cjson.encode({Ok= true, Sid= _sid, Info= _info, Fid= _frontSid })
if _sid == false then
	_sid = ""
end
if _frontSid == false then
	_frontSid = ""
end
return cjson.encode({Ok= true, Sid= _sid, Info= _info, Fid= _frontSid })
	`)
	scriptHash["bindUserSid"] = script
	s := script.Load(rdb)
	fmt.Println(s.Result())
}

func Auth(_strToken *string, sid interfaces.UserId) (user interfaces.UserInfo, logErr loginErr, ok bool) {
	return
	//headers := strings.Split(*_strToken, ", ")
	//strToken := ""
	//if len(headers)%2 == 0 {
	//	for i := 0; i < len(headers); i = i + 2 {
	//		if headers[i] == "access_token" {
	//			strToken = strings.Trim(headers[i+1], " ,")
	//			break
	//		}
	//	}
	//}
	//if strToken == "" {
	//	logErr = loginErr{Code: 1211, Msg: "token not found0"}
	//	return
	//}
	//
	//var uid int64
	//if s := strings.Split(strToken, "."); len(s) == 3 {
	//	if ss, err := base64.RawStdEncoding.DecodeString(s[1]); err == nil {
	//		var d tokenParsed
	//		if err = json.Unmarshal(ss, &d); err == nil {
	//			uid = int64(d.UID)
	//		}
	//	}
	//}
	//if res, err := scriptHash["bindUserSid"].EvalSha(rdb, []string{fmt.Sprintf(userToken, uid)}, sid, strToken, *conf.ServerID).String(); err != nil {
	//	logErr = loginErr{Code: 1411, Msg: "not found1"}
	//	fmt.Println("err => ", err, "res => ", res)
	//} else {
	//	var r BindRes
	//	fmt.Println(res)
	//	if err := json.Unmarshal([]byte(res), &r); err == nil {
	//		if r.Ok && r.Info != "" {
	//			if err := json.Unmarshal([]byte(r.Info), &user); err == nil {
	//				user.UID = uid
	//				ok = true
	//			} else {
	//				logErr = loginErr{Code: 1411, Msg: "not found2"}
	//			}
	//		} else {
	//			fmt.Println(err)
	//			logErr = loginErr{Code: 1411, Msg: "not found3"}
	//		}
	//	} else {
	//		fmt.Println(err)
	//		fmt.Println("fail 10 => ", string(res))
	//	}
	//}
	//return
}

type loginErr struct {
	Code int32
	Msg  string
}
