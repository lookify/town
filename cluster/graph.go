package cluster

import (
  "log"
  "sync"
  // dockerapi "github.com/fsouza/go-dockerclient"
)

type Node struct {
  ID string

//  status *Status

  config *Container  // rename it to container

  sync.Mutex
}

// was Data changed to Graph
type Graph struct {
  Nodes []*Node

  sync.Mutex

  Out map[*Node][]*Node
  In map[*Node][]*Node

  nodesIndex map[string]int
}

func NewGraph() *Graph {
  return &Graph{
    Nodes:  []*Node{},
    Out:  make(map[*Node][]*Node),
    In:   make(map[*Node][]*Node),
    nodesIndex: make(map[string]int),
  }
}

func NewNode(id string) *Node {
  return &Node{
    ID:    id,
   // status: &Status {
      // running: 0,
      // exist: []string{ },
      // start: 0,
    //  links: []string{},
    //  scale: 0,
    //  ports: make(map[dockerapi.Port][]dockerapi.PortBinding),
    //},
  }
}


func (g *Graph) AddNode(node *Node) (bool, error) {
  if _, ok := g.nodesIndex[node.ID]; ok {
    return false, log.Println(node.ID, " already exists")
  }
  g.Mutex.Lock()
  g.nodesIndex[node.ID] = len(g.Nodes)
  g.Mutex.Unlock()
  g.Nodes = append(g.Nodes, node)
  return true, nil
}


func (g *Graph) FindNodeByID(id string) *Node {
  if index, ok := g.nodesIndex[id]; ok && index >= 0 {
    return g.Nodes[index]
  }
  return nil
}

func (g *Graph) DeleteNode(node *Node) {
  if idx, ok := g.nodesIndex[node.ID]; ok && idx >= 0 {
    copy(g.Nodes[idx:], g.Nodes[idx+1:])
    g.Nodes[len(g.Nodes)-1] = nil
    g.Nodes = g.Nodes[:len(g.Nodes)-1 : len(g.Nodes)-1]
  }

  g.Mutex.Lock()
  delete(g.nodesIndex, node.ID)
  g.Mutex.Unlock()
}

func (g *Graph) ConnectNodes(src, dst *Node, connection map[*Node][]*Node) {
  if _, ok := connection[src]; ok {
    isDuplicate := false
    for _, node := range connection[src] {
      if node == dst {
        isDuplicate = true
        break
      }
    }

    if !isDuplicate {
      connection[src] = append(connection[src], dst)
    }
  } else {
    connection[src] = []*Node{dst}
  }
}

func (g *Graph) FindConnection(node *Node, connection map[*Node][]*Node) []*Node {
  if _, ok := connection[node]; ok {
    return connection[node]
  } else {
    return nil
  }
}

func (g *Graph) FindOutConnections(root *Node) []string {
  connections := []string{}
  nodes := g.FindConnection(root, g.Out)
  if nodes != nil {
    for _, node := range nodes {
      connections = append(connections, node.ID)
    }
  }
  return connections
}

func (g *Graph) Connect(src, dst *Node) {
  isAdded, _ := g.AddNode(src)
  if !isAdded {
    src = g.FindNodeByID(src.ID)
  }
  isAdded, _ = g.AddNode(dst)
  if !isAdded {
    dst = g.FindNodeByID(dst.ID)
  }

  g.Mutex.Lock()

  g.ConnectNodes(src, dst, g.In)
  g.ConnectNodes(dst, src, g.Out)

  g.Mutex.Unlock()
}

func (g *Graph) Topsort() []*Node {
  sort := []*Node{}
  noIncome := []*Node{}
  income := make(map[*Node]int)

  for _, node := range g.Nodes {
    if _, ok := g.In[node]; ok {
      income[node] = len(g.In[node])
    } else {
      noIncome = append(noIncome, node)
    }
  }

  for len(noIncome) > 0 {
    last := len(noIncome) - 1
    n := noIncome[last]
    noIncome = noIncome[:last]
    sort = append(sort, n)

    for _, m := range g.Out[n] {
      if income[m] > 0 {
        // log.Println(n.ID, " loaded from ", m.ID)
        income[m]--
        if income[m] == 0 {
          noIncome = append(noIncome, m)
        }
      }
    }
  }

  for c, in := range income {
      if in > 0 {
        log.Println("Cyclic ", c.ID, " = ", in);
        // TODO
      }
    }
  return sort
}