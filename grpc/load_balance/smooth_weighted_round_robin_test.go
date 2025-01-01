package load_balance

import (
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ecodeclub/ekit/slice"
)

type Node struct {
	name           string
	originalWeight int
	currentWeight  int
}

func (n *Node) Invoke() {
	log.Print("模拟发起了 rpc 调用，选中了：", n.name)
}

// 平滑加权轮询算法，测试每次选出来的节点
// 防止权重高的服务端节点连续收到多个请求
func TestSmoothWeightedRoundRobin(t *testing.T) {
	nodes := []*Node{
		{name: "A", originalWeight: 10, currentWeight: 0},
		{name: "B", originalWeight: 20, currentWeight: 0},
		{name: "C", originalWeight: 30, currentWeight: 0},
	}
	b := &Balancer{
		nodes: nodes,
		mu:    sync.Mutex{},
	}

	for i := 1; i <= 6; i++ {
		targetNode := b.pick(t, i)
		targetNode.Invoke()
		t.Log("=======================================================")
	}
}

type Balancer struct {
	nodes []*Node
	mu    sync.Mutex

	idx *atomic.Int32
}

func (b *Balancer) pick(t *testing.T, i int) *Node {
	b.mu.Lock()
	defer b.mu.Unlock()
	// 计算总权重
	totalWeight := 0
	for _, node := range b.nodes {
		totalWeight += node.originalWeight
	}

	// 每次选择前加上初始权重
	t.Logf("开始前，nodes：%v\n", slice.Map(b.nodes, func(idx int, src *Node) Node { return *src }))
	t.Logf("第%d次请求来了", i)
	for _, node := range b.nodes {
		node.currentWeight += node.originalWeight
	}
	t.Logf("加上初始权重后，nodes：%v\n", slice.Map(b.nodes, func(idx int, src *Node) Node { return *src }))

	// 挑选 current weight 最大的节点
	var targetNode *Node
	for _, node := range b.nodes {
		// 注意这里取等也替换，相当于两者相等时让新的节点轮替
		if targetNode == nil || node.currentWeight >= targetNode.currentWeight {
			targetNode = node
		}
	}
	t.Log("选中了节点：", targetNode.name)
	targetNode.currentWeight -= totalWeight
	t.Logf("选中节点减去总权重后，nodes：%v\n", slice.Map(b.nodes, func(idx int, src *Node) Node { return *src }))
	return targetNode
}

// ===========================================================================================
// 加权随机算法不需要 currentWeight
func (b *Balancer) weightedRandomPick() *Node {
	totalWeight := 0
	for _, node := range b.nodes {
		totalWeight += node.originalWeight
	}
	// 如果输入有序这里就不需要排序
	// sort.Slice(b.nodes, func(i, j int) bool {
	// 	return b.nodes[i].originalWeight < b.nodes[j].originalWeight
	// })
	// 随机数落在 [0,10) 之间选 A，落在 [10,30) 之间选 B，落在 [30,60) 之间选 C
	r := rand.Intn(totalWeight)
	for _, node := range b.nodes {
		r -= node.originalWeight
		if r < 0 {
			return node
		}
	}
	panic("impossible")
}

func (b *Balancer) random() *Node {
	r := rand.Int()
	return b.nodes[r%len(b.nodes)]
}

func (b *Balancer) roundRobin() *Node {
	idx := b.idx.Add(1)
	return b.nodes[int(idx)%len(b.nodes)]
}
