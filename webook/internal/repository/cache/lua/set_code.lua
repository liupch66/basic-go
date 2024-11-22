-- cmd.Eval(ctx, scriptSetCode, []string{cache.key(biz, phone)}, code).Int()
-- key 为 phone_code:$biz:$phone
local key = KEYS[1]
-- Integer reply: TTL in seconds.
--            -1: the key exists but has no associated expiration.
--            -2: if the key does not exist.
-- 验证码的有效期是 10 分钟
local ttl = tonumber(redis.call("ttl", key))
-- 验证码剩余可使用次数，这是个字符串连接操作( code:$biz:$phone:cnt)
local cntKey = key..":cnt"
local code = ARGV[1]

-- key 存在但是没有过期时间，异常状态
if ttl == -1 then
    return -2
    -- ttl == -2: key 不存在或已过期（键过期就会被删除）（没有发送过验证码，或验证码已过期）
    -- ttl <540: key 过期时间小于 9 分钟（验证码发送超过 1 分钟，可以重新发送）
elseif ttl == -2 or ttl <540 then
    -- setex key seconds value
    redis.call("setex", key, 600, code)
    redis.call("setex", cntKey, 600, 3)
    return 0
else
    -- ttl >= 540: 已经发送过验证码，且不超过 1 分钟
    return -1
end
