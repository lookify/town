# Town

Town help you run multi-container application in Docker. Defining a single cluster configuration file.
It is simple, just like Docker Compose tool. However, it has several advantages:
 * partial cluster update - in case images were updated it will restart only the dependant part of the cluster
 * dynamic configuration support - IP address or number of the scale (i.e., creating number of images for this instance) can be passed

Town is small and simple tool help small startups run they infrastructure very quickly.

# Install
Download the latest release:
 * [Mac x86](https://raw.github.com/lookify/release/master/town/darwin/386/town)
 * [Mac x64](https://raw.github.com/lookify/release/master/town/darwin/amd64/town)
 * [Linux x86](https://raw.github.com/lookify/release/master/town/linux/386/town)
 * [Linux x64](https://raw.github.com/lookify/release/master/town/linux/amd64/town)
 * [Linux ARM](https://raw.github.com/lookify/release/master/town/linux/arm/town)
 * [Freebsd x86](https://raw.github.com/lookify/release/master/town/freebsd/386/town)
 * [Freebsd x64](https://raw.github.com/lookify/release/master/town/freebsd/amd64/town)
 * [Freebsd ARM](https://raw.github.com/lookify/release/master/town/freebsd/arm/town)
 * [Windows x86](https://raw.github.com/lookify/release/master/town/windows/386/town.exe)
 * [Windows x64](https://raw.github.com/lookify/release/master/town/windows/amd64/town.exe)

# Example

Configuration files must be in current directoy or /etc/town/ 
Configuration file describe containers, container relationship itself and scale of the cluster.

demo.yaml
```yml
application:
  cluster:
    www: 1
    redis: 1
    nginx: 1
  docker:
    hosts:
      - unix:///var/run/docker.sock

redis:
  image: redis
  environment:
   - REDIS_PASS=secretpassword

www:
  image: node
  command: ${SCALE_INDEX}
  links:
   - redis
  volumes:
   - /var/log/www-${SCALE_INDEX}/log/:/var/log/nodejs

nginx:
  image: nginx
  privileged: true
  environment:
    - WWW_HOSTS=${WWW_HOSTS}
  ports:
   - "80:80"
  links:
   - node
  volumes:
    - /opt/public:/opt/static/
```

To Run town simple execute command:

```sh
town run
```

# Configuration Reference
The configuration file /etc/town/town.yaml has list of containers names. The _application_ item is reserved keyword and used for cluster definition. Here you can define how many instances of containers town needs to create.

Each container must have *image* key. The keys _command_, _environment_ and _volume_ can containe dynamic variables:
 * ${SCALE_INDEX} - Index of given container. Start from 1.
 * ${(container_name)_HOSTS} - Comma separated list of hosts.

#### image
Docker image definition.

#### command
Override default command of the container.

#### links
Link to container in another service. Link is the name of other containers and will be dynamicly added depends of the container scale.

#### ports
Expose ports.

#### environment
List of environments variables passed to container.

#### volumes
Mount paths as volumes.

# Commands Reference

#### run
Run containers. If the container is already running nothing will happend. But if the conatiner has new image, town will catch it and restart this container and all refer to it.

#### restart
Restart all cluster. This operation will gracefully shutdown, remove containers and start from sratch whole cluster.

#### stop
Stop and remove all containers.

# Creators

_Kiril Menshikov_ - https://twitter.com/kiril

_Arturs Licis_ - https://twitter.com/arturs_li

# Copyright and license
Code and documentation copyright 2015 [Lookify.co](http://www.lookify.co) Code released under the Apache 2.0 license.

