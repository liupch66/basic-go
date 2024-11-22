-- cmd.Eval(ctx, luaVerifyCode, []string{key}, inputCode)
local key =KEYS[1]
local cntKey = key..":cnt"
local inputCode = ARGV[1]
-- 剩余可使用次数
local cnt = tonumber(redis.call("get", cntKey))
local code = redis.call("get", key)
-- 检查验证码是否过期
local ttl = tonumber(redis.call("ttl", key))
if ttl <= 0 then
    return -3
end
-- 验证码可使用次数用完了，验证码无效
if cnt <= 0 then
    return -1
end
-- 验证成功，将验证码置为无效
if inputCode == code then
    redis.call("set", cntKey, -1)
    return 0
else
    -- decr key,验证失败，验证码可使用次数减少 1 次
    redis.call("decr", cntKey)
    return -2
end