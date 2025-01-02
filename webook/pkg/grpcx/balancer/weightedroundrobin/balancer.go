package weightedroundrobin

import (
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const name = "custom_weighted_round_robin"

// 参考 google.golang.org/grpc@v1.69.2/balancer/roundrobin/roundrobin.go
// 可以知道要实现 base.PickerBuilder 接口，而 Build 函数又要生成 balancer.Picker 接口，所以也要实现
func init() {
	balancer.Register(base.NewBalancerBuilder(name, &wrrPickerBuilder{}, base.Config{HealthCheck: true}))
}

type wrrPickerBuilder struct{}

func (*wrrPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	conns := make([]*weightedConn, 0, len(info.ReadySCs))
	for subConn, subConnInfo := range info.ReadySCs {
		conn := weightedConn{subConn: subConn}

		// 打断点发现 md 类型是 interface{} | map[string]interface{}
		// map: key => string, value => interface{} | float64
		// 这里一定要根据 resolver 的不同找准信息在哪，也有可能在 subConnInfo.Address.Attributes
		// 包括信息的具体类型转化，都可以先断点打印查看
		// weight 本来是 int，这里却变成了 float64
		// 这是因为 etcd 注册之后是一个 json 串，反序列化回来数字默认类型就是 float64
		md := subConnInfo.Address.Metadata
		// 防止 md == nil;用 if md != nil 判断也行
		mdVal, ok := md.(map[string]any)
		if ok {
			weightVal := mdVal["weight"]
			weight, _ := weightVal.(float64)
			conn.weight = int(weight)
		} else {
			// set default value
		}

		conns = append(conns, &conn)
	}

	return &wrrPicker{conns: conns}
}

type wrrPicker struct {
	conns []*weightedConn
	mu    sync.Mutex
}

// Pick 实现负载均衡算法的地方，这里实现（基于权重的轮询）负载均衡算法
func (p *wrrPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight int
	var targetConn *weightedConn
	for _, conn := range p.conns {
		totalWeight += conn.weight
		conn.currentWeight += conn.weight
		if targetConn == nil || conn.currentWeight >= targetConn.currentWeight {
			targetConn = conn
		}
	}
	targetConn.currentWeight -= totalWeight
	return balancer.PickResult{
		SubConn: targetConn.subConn,
		Done: func(info balancer.DoneInfo) {
			// RPC 完成时调用
			// 很多动态算法，在这里根据调用结果来调整权重
		},
	}, nil
}

// 我们自定义算法需要权重信息，所以将 balancer.SubConn 包装了一下，再作为 wrrPicker 成员，
// 这和 roundrobin.go 直接嵌入 balancer.SubConn 中不同
type weightedConn struct {
	subConn       balancer.SubConn
	weight        int
	currentWeight int
}
