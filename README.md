# Town

Town help you run multi-container application in Docker. Defining a single cluster configuration file.

Town is small and simple tool help small startups run they infrastructure very quickly.

# Install
Download the latest release:
 * [Linux](https://github.com/lookify/town/tree/master/release/linux)
 * [Freebsd](https://github.com/lookify/town/tree/master/release/freebsd)
 * [Windows](https://github.com/lookify/town/tree/master/release/windows)

# Example

Main configuration file is /etc/town/town.yaml It describe separate containers and container relationship itself.

```yml
application:
  cluster:
   - "www: 1"
   - "redis: 1"
   - "nginx: 1"

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
The configuration file /etc/town/town.yaml has list of containers names. The *application* item is reserved keyword and used for cluster definition. Here you can define how many instances of containers town needs to create.

Each container must have *image* key. The keys *command*, *environment* and *volume* can containe dynamic variables:
 * ${SCALE_INDEX} - Index of given container. Start from 1.
 * ${(container_name)_HOSTS} - Comma separated list of hosts.

## image
Docker image definition.

## command
Override default command of the container.

## links
Link to container in another service. Link is the name of other containers and will be dynamicly added depends of the container scale.

## ports
Expose ports.

## environment
List of environments variables passed to container.

## volumes
Mount paths as volumes.

# Commands Reference

## run
Run containers. If the container is already running nothing will happend. But if the conatiner has new image, town will catch it and restart this container and all refer to it.

## restart
Restart all cluster. This operation will gracefully shutdown, remove containers and start from sratch whole cluster.

## stop
Stop and remove all containers.
