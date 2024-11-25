--cmd.Eval(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd
--cmd.Eval(ctx, slideWindowLuaScript, []string{key}, b.interval.Milliseconds(), b.rate, time.Now().UnixMilli())

--限流对象
local key = KEYS[1]
--窗口大小
local window = tonumber(ARGV[1])
--限流阈值
local threshold = tonumber(ARGV[2])
--现在时间
local now = tonumber(ARGV[3])
--计算窗口起始时间
local min = now - window

--redis.call('COMMAND', arg1, arg2, ..., argN)
--ZREMRANGEBYSCORE key min max: 移除有序集中，指定分数（score）区间内的所有成员， -inf 和 +inf：表示无穷范围。
--这里删除过期的请求
redis.call('ZREMRANGEBYSCORE', key, '-inf', min)
--以 0 表示有序集第一个成员，以此类推。以 -1 表示最后一个成员，以此类推。
--redis.call('ZREMRANGEBYSCORE', key, 0, min)

--ZCARD key: 计算集合中元素的数量
local cnt = redis.call('ZCARD', key)
--ZCOUNT key min max: 计算有序集合中指定分数区间的成员数量
--local cnt = redis.call('ZCOUNT', min, '+inf')
--local cnt = redis.call('ZCOUNT', '-inf', '+inf')

if cnt >= threshold then
    return true
else
    --记录当前请求
    --ZADD key score member: 将成员元素及其分数值加入到有序集当中
    redis.call('ZADD', key, now, now)
    --设置过期时间，节省 redis 内存资源
    redis.call('PEXPIRE', key, window)
    return false
end





