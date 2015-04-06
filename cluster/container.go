package cluster

import (
//  dockerapi "github.com/fsouza/go-dockerclient"
)

type ExistContainer struct {
  ID string
  Name string
  Index int
  Running bool
}

func NewExistContainer(id string, name string, index int, running bool) *ExistContainer {
  return &ExistContainer{
    ID: id,
    Name: name,
    Index: index,
    Running: running,
  }
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
  Cluster []string
  Post string
  Privileged bool

  Scale int
  // Links []string
  // Ports map[dockerapi.Port][]dockerapi.PortBinding

  Exist []ExistContainer

  Changed bool
}
