# Town

Town help you run multi-container application in Docker. Defining a single cluster configuration file.

Town is small and simple tool help small startups run they infrastructure very quickly.

# Create Cluster Configuration

Main configuration file is town.yaml It describe separate containers and container relationship itself.

Example

redis:
  image: lookify/redis:latest

node:
  image: lookify/www:latest
  links:
   - redis

nginx:
  image: lookify/nginx-www:latest
  ports:
   - "80:80"
  links:
   - node
  volumes:
    - /opt/public:/opt/public/


