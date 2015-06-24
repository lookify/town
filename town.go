package main

import (
  "strings"
  "log"
  // "fmt"
//  "regexp"
  "bytes"
  "encoding/json"
  "os"
  "time"
  "strconv"
  "regexp"
  "github.com/lookify/town/cluster"
  dockerapi "github.com/fsouza/go-dockerclient"
)

var (
  SCALE_NUM_REG, _ = regexp.Compile("\\$\\{SCALE_NUM:(.+)\\}")
//  SCALE_TOTAL_REG, _ = regexp.Compile("\\$\\{SCALE_NUM:(.+)\\}")
  HOSTS_REG, _ = regexp.Compile("\\$\\{(.+)_HOSTS\\}")
)

type Town struct {
  cluster *cluster.Cluster
  docker *dockerapi.Client
}

func NewTown() *Town {
  return &Town{
    cluster: nil,
    docker: nil,
  }
}

func (t *Town) ReadFile(name string) {
  var pathLocs = [...]string{
    name + ".yml",
    "/etc/town/" + name + ".yml",
  }

  for _, path := range pathLocs {
    if _, err := os.Stat(path); err == nil {
      t.cluster = cluster.NewCluster(path)
      t.cluster.ReadFile()
      return
    }
  }

  log.Println("ERROR: Could not find file ", name, ".yml")
}

func (t *Town) Connect() {
  endpoint := t.cluster.Application.Docker.Hosts[0] // at the moment use only first
  log.Println("Using Docker API endpont: ", endpoint)
  docker, err := dockerapi.NewClient( endpoint )
  if err != nil {
    log.Println("Can't connect to the docker")
  }
  t.docker = docker
}

func (t *Town) Provision(checkChanged bool) {
  // update containers
  for _, node := range t.cluster.Nodes {
    var buf bytes.Buffer
    var image = strings.Split(node.Container.Image, ":")

    opts := dockerapi.PullImageOptions{
        Repository: image[0],
        // Registry:     "docker.tsuru.io",
        // Tag:          "latest",
        OutputStream: &buf,
    }

    if len(image) > 1 {
      opts.Tag = image[1]
    }
    err := t.docker.PullImage(opts, dockerapi.AuthConfiguration{});
    if err != nil {
      log.Println("Could not pull image ", image)
    }
  }

  allContainers, err := t.docker.ListContainers(dockerapi.ListContainersOptions{
    All: true,
  })
  if err == nil {
    for _, listing := range allContainers {
      container, err := t.docker.InspectContainer(listing.ID)
      if err == nil {
        name := container.Name[1:]
        node, index := t.cluster.FindNodeByName(name)
        if node != nil && index > 0 {
          if node.Container.Exist == nil {
            node.Container.Exist = []cluster.ExistContainer{}
          }
          runningContainer := cluster.NewExistContainer(listing.ID, name, index, container.State.Running)
          if checkChanged {
            node.Container.Changed = t.isChangedImage(node, container)
          } else {
            node.Container.Changed = true
          }
          node.Container.Exist = append(node.Container.Exist, runningContainer)
        }
      } else {
        log.Println("[ERROR] Unable to inspect container:", listing.ID[:12], err)
      }
    }

    if checkChanged {
      t.cluster.AddChangeDependant()
    }
  } else {
    log.Println("[ERROR] Can't start provision")
  }
}


/**
 * Check node and running container for changes.
 * TODO: add cache to docker call.
 **/
func (t *Town) isChangedImage(node *cluster.Node, container *dockerapi.Container) bool {
  var imageName = container.Image
  image , error := t.docker.InspectImage(imageName)
  if error == nil {
    secondImage , secondError := t.docker.InspectImage(node.Container.Image)
    if secondError == nil {
      return secondImage.Created.After(image.Created)
    } else {
      log.Println("[ERROR] Could not inspect image ", node.Container.Name)
    }
  } else {
    log.Println("[ERROR] Could not inspect image ", imageName)
  }
  return false
}

func (t *Town) StopContainers(checkChanged bool) {
  log.Println("Stop...")
  //for node := range t.cluster.nodes {
  for i := len(t.cluster.Nodes) - 1; i >= 0; i-- {
    node := t.cluster.Nodes[i]
    if (!checkChanged || node.Container.Changed) && len(node.Container.Exist) > 0 {
      for _, container := range node.Container.Exist {
        if container.Running {
          err := t.docker.StopContainer(container.ID, 10)
          if err == nil {
            log.Println("   -  ", container.Name)
          } else {
            log.Println("   -  ", container.Name, " failed ", err)
          }
        }
      }
    }
  }
  log.Println("=============================")
}

func (t *Town) RemoveContainers(checkChanged bool) {
  log.Println("Remove...")
  //for node := range t.cluster.nodes {
  for i := len(t.cluster.Nodes) - 1; i >= 0; i-- {
    node := t.cluster.Nodes[i]
    if (!checkChanged || node.Container.Changed) && len(node.Container.Exist) > 0 {
      for _, container := range node.Container.Exist {
        err := t.docker.RemoveContainer(dockerapi.RemoveContainerOptions{
          ID: container.ID,
          RemoveVolumes: false,
        })
        if err == nil {
          log.Println("   -  ", container.Name)
        } else {
          log.Println("   -  ", container.Name, " failed ", err)
        }
      }
    }
  }
  log.Println("=============================")
}



func (t *Town) CreateContainer(node *cluster.Node, index int) (string, string) {
  containerName := node.Container.Name + "-" + strconv.Itoa(index)

  log.Println("   -  ", containerName)

  node.Container.Hostname = containerName // ?? Help !!!!

  env :=  make([]string, 0, cap(node.Container.Environment))
  for _, e := range node.Container.Environment {
    env = append(env, t.exec(e, index))
  }

  volumes := make(map[string]struct{})
  binds := make([]string, 0, cap(node.Container.Volumes))
  if len(node.Container.Volumes) > 0 {
    for _, volume := range node.Container.Volumes {
      volume = t.exec(volume, index)
      vol := strings.Split(volume, ":")
      if len(vol) > 1 {
        volumes[vol[1]] = struct{}{}
      } else {
        volumes[vol[0]] = struct{}{}
      }
      binds = append(binds, volume)
    }
  }

  dockerConfig := dockerapi.Config{
      Image: node.Container.Image,
      Hostname: node.Container.Hostname,
      PortSpecs: node.Container.Ports,
      Env: env,
      Volumes: volumes,

      AttachStdout: false,
      AttachStdin: false,
      AttachStderr: false,

      Tty: false,

      //Cmd: []
  }

  if len(node.Container.Command) > 0 {
    cmd := t.exec(node.Container.Command, index)
    dockerConfig.Cmd = []string{ cmd }
  }

  // just info
  //for _, l := range node.status.links {
  //  log.Println("     * ", l)
  //}

  // create links
  links := t.cluster.GetLinks(node)

  portBindings := map[dockerapi.Port][]dockerapi.PortBinding{}
  // create ports
  for _, ports := range node.Container.Ports {

    port := strings.Split(ports, ":")
    var p dockerapi.Port

    if len(port) > 1 {
      p = dockerapi.Port(port[1] + "/tcp")
    } else {
      p = dockerapi.Port(port[0] + "/tcp")
    }

    if portBindings[p] == nil {
      portBindings[p] = [] dockerapi.PortBinding {}
    }

    portBindings[p] = append(portBindings[p], dockerapi.PortBinding{
      HostIP: "",
      HostPort: port[0],
    })
  }

  hostConfig := dockerapi.HostConfig{
    Binds: binds,
    Links: links, //, [],
    PortBindings: portBindings,
    NetworkMode: "bridge",
    PublishAllPorts: false,
    Privileged: node.Container.Privileged,
  }

  opts := dockerapi.CreateContainerOptions{Name: containerName, Config: &dockerConfig, HostConfig: &hostConfig}
  container, err := t.docker.CreateContainer(opts)
  if err == nil {
    retry := 5
    for retry > 0 {
      error := t.docker.StartContainer(container.ID, &hostConfig)
      if error != nil {
        // log.Println("start error: ", error);

        out, err := json.Marshal(container)
        if err != nil {
            panic (err)
        }
        // fmt.Println(string(out))

        retry--;
        if retry == 0 {
          log.Println(" Start failed after 5 retries: ", string(out))
        }
        // log.Println("retry: ", retry);
      } else {
        inspect, inspectError := t.docker.InspectContainer(container.ID)
        if inspectError == nil {
          //links = append(links, inspect.NetworkSettings.IPAddress + "  " + containerName)
          //ids = append(ids, container.ID)
          return container.ID, inspect.NetworkSettings.IPAddress + "  " + containerName
        } else {
          log.Println("Inpect ", container.ID, " error ", inspectError)
        }
        //retry = 0
        break;
      }
    }
  } else {
    log.Println("error: ", err);
  }

  return "", ""
}

func (t *Town) CreateContainers(checkChanged bool) {
  log.Println("Create...")
  for _, node := range t.cluster.Nodes {

    if !checkChanged || node.Container.Changed {
      ids := make([]string, 0, node.Container.Scale )

      hosts := make([]string, 0, node.Container.Scale)

      for i := 1; i <= node.Container.Scale; i++ {
        log.Println(node.Container.Name, "  image: ", node.Container.Image)
        id, host := t.CreateContainer(node, i)
        ids = append(ids, id)
        hosts = append(hosts, host)
      }

      if len(ids) > 1 {
        for index, id := range ids {
          var buffer bytes.Buffer

          buffer.WriteString("echo -e '")
          for i := 0; i < len(hosts); i++ {
            if i != index {
              buffer.WriteString("\n")
              buffer.WriteString(hosts[i])
            }
          }
          buffer.WriteString("' >> /etc/hosts; touch /tmp/host-generated")
          t.bashCommand(id, buffer.String() )
        }
      }

      time.Sleep(1000 * time.Millisecond)
    }
  }
}

func (t *Town) bashCommand(id string, command string)  {
  config := dockerapi.CreateExecOptions{
    Container:    id,
    AttachStdin:  true,
    AttachStdout: true,
    AttachStderr: false,
    Tty:          false,
    Cmd:          []string{"bash", "-c", buffer.String()},
  }
  execObj, err := t.docker.CreateExec(config)
  if err == nil {
    config := dockerapi.StartExecOptions{
      Detach: true,
    }
    err := t.docker.StartExec(execObj.ID, config)
    if err != nil {
      log.Println("Start exec failed ", id, " error: ", err)
    }
  } else {
    log.Println("Create exec failed ", id, " error: ", err)
  }
}

func (t *Town) exec(text string, scale int) string {
  replace := strings.Replace(text, "${SCALE_NUM}", strconv.Itoa(scale), -1)
  match := SCALE_NUM_REG.FindAllStringSubmatch(replace, -1)
  hostMatch := HOSTS_REG.FindAllStringSubmatch(replace, -1)
  if len(match) > 0 {
    if len(match[0]) > 1 {
      nums := strings.Split(match[0][1], ",")
      if len(nums) > (scale - 1) {
        replace = strings.Replace(replace, match[0][0], nums[scale - 1], -1)
      }
    }
  }
  if len(hostMatch) > 0 {
    if len(hostMatch[0]) > 1 {
      //nums := strings.Split(, ",")
      name := strings.ToLower(hostMatch[0][1])
      node := t.cluster.FindNodeByID(name)

      var buffer bytes.Buffer
      for i := 1; i <= node.Container.Scale; i++ {
        buffer.WriteString(name)
        buffer.WriteString("-")
        buffer.WriteString(strconv.Itoa( i ))
        if i != node.Container.Scale {
          buffer.WriteString(",")
        }
      }
      replace = strings.Replace(replace, hostMatch[0][0], buffer.String(), -1)
    }
  }
  return replace
}
