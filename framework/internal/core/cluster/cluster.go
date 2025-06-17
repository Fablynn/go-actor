package cluster

import (
	"poker_server/common/pb"
	"poker_server/library/mlog"
	"poker_server/library/random"
	"sync"
)

type pool struct {
	mutex sync.RWMutex
	nodes []*pb.Node // 节点
}

type Cluster struct {
	pools map[pb.NodeType]*pool
}

func New() *Cluster {
	pools := make(map[pb.NodeType]*pool)
	for i := pb.NodeType_NodeTypeBegin + 1; i < pb.NodeType_NodeTypeEnd; i++ {
		pools[i] = new(pool)
	}
	return &Cluster{pools: pools}
}

func (c *Cluster) GetCount(nodeType pb.NodeType) int {
	nn := c.pools[nodeType]
	return len(nn.nodes)
}

func (c *Cluster) List(nodeType pb.NodeType) (rets []*pb.Node) {
	nn := c.pools[nodeType]
	nn.mutex.RLock()
	defer nn.mutex.RUnlock()
	for _, val := range nn.nodes {
		rets = append(rets, val)
	}
	return
}

// 随机获取节点
func (c *Cluster) Get(nodeType pb.NodeType, nodeId int32) *pb.Node {
	nn := c.pools[nodeType]
	if nn == nil {
		return nil
	}
	nn.mutex.RLock()
	defer nn.mutex.RUnlock()
	for _, val := range nn.nodes {
		if val.GetId() == nodeId {
			return val
		}
	}
	return nil
}

// 删除节点
func (c *Cluster) Del(nodeType pb.NodeType, nodeId int32) bool {
	if nn := c.Get(nodeType, nodeId); nn == nil {
		return false
	}

	nn := c.pools[nodeType]
	nn.mutex.Lock()
	defer nn.mutex.Unlock()
	j := -1
	for _, val := range nn.nodes {
		if val.GetId() == nodeId {
			mlog.Debugf("删除服务节点：%s", val.String())
			continue
		}
		j++
		nn.nodes[j] = val
	}
	nn.nodes = nn.nodes[:j+1]
	return true
}

// 添加节点
func (c *Cluster) Add(node *pb.Node) bool {
	// 如果节点已存在，则不添加
	if nn := c.Get(node.Type, node.Id); nn != nil {
		return false
	}

	nn := c.pools[pb.NodeType(node.Type)]
	if nn == nil {
		return false
	}
	nn.mutex.Lock()
	defer nn.mutex.Unlock()
	for i, item := range nn.nodes {
		if item.GetId() == node.GetId() {
			nn.nodes[i] = node
			return true
		}
	}
	nn.nodes = append(nn.nodes, node)
	return true
}

// 随机获取节点
func (c *Cluster) Random(nodeType pb.NodeType, seed uint64) *pb.Node {
	nn := c.pools[nodeType]
	llen := len(nn.nodes)
	if llen <= 0 {
		return nil
	}

	nn.mutex.RLock()
	defer nn.mutex.RUnlock()
	if seed <= 0 {
		return nn.nodes[random.Int32n(int32(llen))]
	}
	return nn.nodes[seed%uint64(llen)]
}
