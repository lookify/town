package cluster

import (
  "strings"
  "log"
  "regexp"
  "strconv"

  "io/ioutil"
  "path/filepath"
  "gopkg.in/yaml.v2"
  // "os/exec"
  // dockerapi "github.com/fsouza/go-dockerclient"
  // "encoding/json"
)


type Cluster struct {
  filename string
  config  map[string]Container // rename to containers
  graph *Graph
  application *Container
  cluster []string
  nodes []*Node
//  docker *dockerapi.Client
}


func NewCluster() *Cluster {
  return &Cluster{
    filename: "./town.yml",
    config: make(map[string]Container),
    graph : NewGraph(),
  }
}

func CopyContainerConfig(container *Container) *Container {
  copy := &Container{}
  *copy = *container

  return copy;
}

func (c *Cluster) AddChangeDependant() {
  for node := range c.nodes {
    // && len(node.config.Exist)
    if node.config.Changed {
      log.Println("Check ", node.ID)
      parents := c.graph.FindConnection(node, c.graph.In)
      if parents != nil {
        for parent := range parents {
          log.Println("  - ", parent.ID)
          parent.config.Changed = true
        }
      }
    }
  }
}

func (c *Cluster) AddContainer(name string, container Container) {
  container.Name = strings.TrimSpace( name );
  if container.Name == "application" {
    c.application = CopyContainerConfig(&container)
    c.cluster = container.Cluster
  } else {
    node := c.graph.FindNodeByID(container.Name)
    if node == nil {
      node = NewNode(container.Name)
      c.graph.AddNode(node)
    }

    node.config = CopyContainerConfig(&container)

    for _, link := range container.Links {
      link = strings.TrimSpace( link );
      childNode := c.graph.FindNodeByID(link)
      if childNode == nil {
        childNode = NewNode(link)
        c.graph.AddNode(childNode)
      }
      c.graph.Connect(node, childNode)
    }
  }
}

func (c *Cluster) CheckCluster() {
  for _, name := range c.cluster {
    split := strings.Split(name, ":")
    name = strings.TrimSpace(split[0])

    found := false
    for _, node := range c.graph.Nodes {
      if (name == node.config.Name) {
        scale, err := strconv.Atoi( strings.TrimSpace(split[1]) )
        if err == nil {
          node.config.Scale = scale
        } else {
          log.Println("ERROR: Could not parse sclae number ", split[1], " for container ", name)
        }

        found = true
        break
      }
    }
    if (!found) {
      log.Println("ERROR: node '", name, "' defined in application's cluster, but missing configuration")
    }
  }
}

func (c *Cluster) ReadFile() {
  absFileName, _ := filepath.Abs(c.filename)
  yamlFile, err := ioutil.ReadFile(absFileName)

  if err != nil {
    //panic(err)
    log.Fatal("Couldn't read yml: ", err);
  }

  err = yaml.Unmarshal(yamlFile, &c.config)
  if err != nil {
    //panic(err)
    log.Fatal("Couldn't parse yml: ", err);
  }

  for key, container := range c.config {
    c.AddContainer(key, container)
  }

  c.CheckCluster()

  c.nodes = c.graph.Topsort()
}

func (c *Cluster) findNodeByName(name string) (*Node, int) {
  nodeName, index := c.ParseName(name)
  return c.graph.FindNodeByID(nodeName), index
}

// t.cluster.findNodeByNmae(name)
//   containerName, _ := c.graph.Name(name);
//   containerNode := c.graph.FindNodeByID(containerName)

// func (c *Cluster) IsRunning(name string, id string) bool {
//   nodeName, num := c.ParseName(name)
//   node := c.graph.FindNodeByID(nodeName)
//   return node != nil
//   //  {
//   //   //if active {
//   //   //  node.status.active = append(node.status.active, match[2])
//   //   //} else {
//   //   //  node.status.exist = append(node.status.exist, num)
//   //   //}
//   //   // node.status.ids = append(node.status.ids, id)

//   //   return true
//   // } else {
//   //   return false
//   // }
// }

func (c *Cluster) ParseName(name string) (string, int) {
  r, _ := regexp.Compile("([a-z\\-]+)-([0-9]+)")
  match := r.FindStringSubmatch(name)
  if len(match) == 1 {
    index, err := strconv.Atoi( match[2] )
    if err == nil {
      return match[1], index
    }
  }
  return name, -1
}