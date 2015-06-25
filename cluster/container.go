package cluster

import (
//  dockerapi "github.com/fsouza/go-dockerclient"
)

type ExistContainer struct {
  ID string
  Name string
  Index int
  Running bool
  Pid int
  User string
}

func NewExistContainer(id string, name string, index int, running bool) ExistContainer {
  return ExistContainer{
    ID: id,
    Name: name,
    Index: index,
    Running: running,
    Pid: -1,
    User: "",
  }
}

type Application struct {
  Cluster map[string]int
  // `yaml:"cluster,inline"`
  Docker Docker
}

type Docker struct {
  Hosts []string
}

type Container struct {
  Name string
  Image string
  Hostname string
  Ports []string
  Environment []string
  Links []string
  Volumes []string
  Command string
 
  Post string
  Privileged bool

  Scale int
  // Links []string
  // Ports map[dockerapi.Port][]dockerapi.PortBinding

  Exist []ExistContainer

  Changed bool

  // Application level 
  Cluster map[string]int
  // `yaml:"cluster,inline"`
  Docker Docker
}


// check for containers and application mixing  `yaml:",inline"`