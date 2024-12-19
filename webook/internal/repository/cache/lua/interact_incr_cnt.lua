--Eval(ctx, luaIncrCnt, []string{"interact:$biz:$bizId"}, "read_cnt", 1)
local key = KEYS[1]
local cntKey = ARGV[1]
local delta = tonumber(ARGV[2])
local exists = redis.call("EXISTS", key)
--hincrby key field increment
if exists == 1 then
    -- 坑：这里的 HINCRBY 必须打引号，不能直接写，直接写会被认为是变量
    -- 这是具体报错：ERR user_script:8: Script attempted to access nonexistent global variable 'HINCRBY' script: 9249508cdaee624bc9dc0155bcf77557a1591e3f, on @user_script:8.
    redis.call("HINCRBY", key, cntKey, delta)
    return 1
else
    return 0
end