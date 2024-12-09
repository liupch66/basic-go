--Eval(ctx, "", []string{"interact:$biz:$bizId"}, "read_cnt", 1)
local key = KEYS[1]
local cntKey = ARV[1]
local delta = tonumber(ARV[2])
local exists = redis.call("EXISTS", key)
--hincrby key field increment
if exists == 1 then
    redis.call(HINCRBY, key, cntKey, delta)
    return 1
else
    return 0
end