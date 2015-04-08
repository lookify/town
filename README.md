# Town

Town help you run multi-container application in Docker. Defining a single cluster configuration file.

Town is small and simple tool help small startups run they infrastructure very quickly.

# Example

Main configuration file is /etc/town/town.yaml It describe separate containers and container relationship itself.

```yml
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
  command: ${SCALE_NUM}
  links:
   - redis
  volumes:
   - /var/log/www-${SCALE_NUM}/log/:/var/log/nodejs

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

# Config Reference
TBD

# Command Reference
TBD





