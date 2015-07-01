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

const DEFAULT_ENDPOINT = "unix:///var/run/docker.sock"

type Cluster struct {
  filename string

  config  map[string]Container // rename to containers
  Application *Application

  graph *Graph
  //application *Container
  Nodes []*Node
//  docker *dockerapi.Client
}

func NewCluster(conf string) *Cluster {
  return &Cluster{
    filename: conf,
    config: make(map[string]Container),
    graph : NewGraph(),
    Application: &Application{
      Docker: Docker{
       Hosts: []string{ DEFAULT_ENDPOINT },
      },
    },
  }
}

func CopyContainerConfig(container *Container) *Container {
  copy := &Container{}
  *copy = *container

  return copy;
}

func doLink(name string, num int) string {
  index := strconv.Itoa(num)
  return name + "-" + index + ":" + name + "-" + index
}

func (c *Cluster) GetLinks(node *Node) []string {
  links := []string{}
  parents := c.graph.In[node]
  for _, parent := range parents {
    for i := 1; i <= parent.Container.Scale; i++ {
      link := doLink(parent.Container.Name, i)
      links = append(links, link);
    }
  }
  return links
}

func (c *Cluster) AddChangeDependant() {
  for _, node := range c.Nodes {
    // && len(node.Container.Exist)
    if node.Container.Changed {
      log.Println("Check ", node.ID)
      parents := c.graph.FindConnection(node, c.graph.Out)
      if parents != nil {
        for _, parent := range parents {
          log.Println("  - ", parent.ID)
          parent.Container.Changed = true
        }
      }
    }
  }
}

func (c *Cluster) AddContainer(name string, container Container) {
  container.Name = strings.TrimSpace( name );
  if container.Name == "application" {
    if container.Cluster != nil {
      c.Application.Cluster = container.Cluster
    }
    if container.Docker.Hosts != nil {
      c.Application.Docker.Hosts = container.Docker.Hosts
    }
  } else {
    node := c.graph.FindNodeByID(container.Name)
    if node == nil {
      node = NewNode(container.Name)
      c.graph.AddNode(node)
    }

    node.Container = CopyContainerConfig(&container)

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
  for name, scale := range c.Application.Cluster {
    found := false
    for _, node := range c.graph.Nodes {
      if (name == node.Container.Name) {
        node.Container.Scale = scale
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

  c.Nodes = c.graph.Topsort()
}


func (c *Cluster) FindNodeByID(name string) (*Node) {
  return c.graph.FindNodeByID(name)
}

func (c *Cluster) FindNodeByName(name string) (*Node, int) {
  nodeName, index := c.ParseName(name)
  return c.FindNodeByID(nodeName), index
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
  if len(match) == 3 {
    index, err := strconv.Atoi( match[2] )
    if err == nil {
      return match[1], index
    }
  }
  return name, -1
}