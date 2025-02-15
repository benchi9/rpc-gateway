package node

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/scroll-tech/rpc-gateway/util"
	"github.com/scroll-tech/rpc-gateway/util/rpc"
)

// NewServer creates node management RPC server
func NewServer(nf nodeFactory, groupConf map[Group]UrlConfig) *rpc.Server {
	managers := make(map[Group]*Manager)
	for k, v := range groupConf {
		managers[k] = NewManager(k, nf, v.Nodes)
	}

	return rpc.MustNewServer("node", map[string]interface{}{
		"node": &api{managers},
	})
}

// api node management RPC APIs.
type api struct {
	managers map[Group]*Manager
}

func (api *api) Add(group Group, url string) {
	if m, ok := api.managers[group]; ok {
		m.Add(url)
	}
}

func (api *api) Remove(group Group, url string) {
	if m, ok := api.managers[group]; ok {
		m.Remove(url)
	}
}

// List returns the URL list of all nodes.
func (api *api) List(group Group) []string {
	m, ok := api.managers[group]
	if !ok {
		return nil
	}

	var nodes []string

	for _, n := range m.List() {
		nodes = append(nodes, n.Url())
	}

	return nodes
}

func (api *api) Status(group Group, url *string) (res []Status) {
	mgr := api.managers[group]
	if mgr == nil { // no group found
		return
	}

	if url != nil { // get specific node status
		if n := mgr.Get(*url); !util.IsInterfaceValNil(n) {
			res = append(res, n.Status())
		}

		return
	}

	// get all group node status
	for _, n := range mgr.List() {
		res = append(res, n.Status())
	}

	return
}

// List returns the URL list of all nodes.
func (api *api) ListAll() map[Group][]string {
	result := make(map[Group][]string)

	for group := range api.managers {
		result[group] = api.List(group)
	}

	return result
}

// Route implements the Router interface. It routes the specified key to any node
// and return the node URL.
func (api *api) Route(group Group, key hexutil.Bytes) string {
	if m, ok := api.managers[group]; ok {
		return m.Route(key)
	}

	return ""
}
