# swarm-prune
This is a little weekend project I through together in a few hours. It is Alpha quality, so please be careful if you use this in production.

## Why
With the Docker 1.13.0 was released, it has added a cool new command that allows you to clean up unused space on your Docker server.

The [prune command](https://github.com/docker/docker/pull/26108) allows you to clean up, containers, images, volumes, and networks. The only problem is that it isn't yet supported with swarm mode. This little tool is a way to use that command across your swarm.

There isn't much to this program, it is mostly just calling the Docker API, and triggering the commands for all nodes in a swarm at once.

## Releasing
you can either run the `./release.sh` script or run `make release`. They do the same thing.

## Building
I'm using a docker container to build a linux binary from my mac, you can generate via `make build`. If you want to build locally, you can build with `make go-build`

## Assumptions
Because I threw this together in such a short amount of time, I made a lot of assumptions, and it works for me, but it might not work for you. Here are some of the assumptions. You should also see the requirements section below.
- The swarm is running on a private network, with docker daemon listening on port 2375, and this port is accessible on all nodes by the swarm managers.

## Important
Please do not expose port 2375 on your swarm unless you know what you are doing, if you do this wrong, you could open up your swarm to the world, and it will only be a matter of time before hackers take it over.

## Requirements
- This needs to run on a swarm manager, and it needs access to the Docker daemon in order to run the commands.
- Docker 1.13 or higher needs to be running on all of the hosts on your swarm
- The swarm has a private network, and the docker daemon for each host is running on port 2375 on that private network.
- The manager where this is running, has access to port 2375 on all other nodes in the swarm.

## How to use
Since this is basically a wrapper around the prune commands, the syntax is almost the same as what you would use locally. Here is the output from the help command.

You can use this a couple of different ways, you can run the binary directly, on a manager, or you can use the docker image.

### binary
Build the binary, move it to your service, and then run it like any other binary.

```
$ ./swarm-prune --help
```

### docker image
Run the image passing in the different commands.

#### Help
```
docker run -it -v /var/run/docker.sock:/var/run/docker.sock kencochrane/swarm-prune:v0.1 swarm-prune --help
```
#### df
```
docker run -it -v /var/run/docker.sock:/var/run/docker.sock kencochrane/swarm-prune:v0.1 swarm-prune df
```

#### full swarm prune
```
docker run -it -v /var/run/docker.sock:/var/run/docker.sock kencochrane/swarm-prune:v0.1 swarm-prune system --force --all
```

### help
```
# ./swarm-prune --help
NAME:
   swarm-prune - swarm wide prune command

USAGE:
   swarm-prune [global options] command [command options] [arguments...]

VERSION:
   v0.1

AUTHOR(S):
   Ken Cochrane <@KenCochrane>

COMMANDS:
     system      This will remove: all stopped containers, all volumes not used by at least one container, all dangling images, all unused networks
     containers  prune containers swarm wide
     images      prune images swarm wide
     volumes     prune volumes swarm wide.
     networks    complete a task on the list
     df          This will show disk usage for all nodes.
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --host value, -H value  Docker Swarm manager host url (default: "unix:///var/run/docker.sock")
   --tlscacert value       TLS CA cert
   --tlscert value         TLS cert
   --tlskey value          TLS key
   --tlsverify             True to skip TLS
   --help, -h              show help
   --version, -v           print the version
```

### system
```
# ./swarm-prune system --help
NAME:
   swarm-prune system - This will remove: all stopped containers, all volumes not used by at least one container, all dangling images, all unused networks

USAGE:
   swarm-prune system [command options] [arguments...]

DESCRIPTION:
   WARNING! This will remove all stopped containers, orphaned in your swarm

OPTIONS:
   --force, -F  Do not prompt for confirmation
   --all, -A    This will remove all images without at least one container associated to them
```

### containers
```
./swarm-prune containers --help
NAME:
   swarm-prune containers - prune containers swarm wide

USAGE:
   swarm-prune containers [command options] [arguments...]

DESCRIPTION:
   WARNING! This will remove all stopped containers in your swarm

OPTIONS:
   --force, -F  Do not prompt for confirmation
```

### images
```
./swarm-prune images --help
NAME:
   swarm-prune images - prune images swarm wide

USAGE:
   swarm-prune images [command options] [arguments...]

DESCRIPTION:
   This will remove all dangling images.

OPTIONS:
   --force, -F  Do not prompt for confirmation
   --all, -A    This will remove all images without at least one container associated to them

```

### volumes
```
./swarm-prune volumes --help
NAME:
   swarm-prune volumes - prune volumes swarm wide.

USAGE:
   swarm-prune volumes [command options] [arguments...]

DESCRIPTION:
   WARNING: This will remove all volumes not used by at least one container

OPTIONS:
   --force, -F  Do not prompt for confirmation
```

### networks
```
./swarm-prune networks --help
NAME:
   swarm-prune networks - complete a task on the list

USAGE:
   swarm-prune networks [command options] [arguments...]

DESCRIPTION:
   WARNING: This will remove all networks not being used

OPTIONS:
   --force, -F  Do not prompt for confirmation
```

### df
```
./swarm-prune df --help
NAME:
   swarm-prune df - complete a task on the list

USAGE:
   swarm-prune df [command options] [arguments...]

DESCRIPTION:
   WARNING: This will remove all networks not being used

OPTIONS:
   --verbose, -V  Verbose output
```

## Contributing
Feel free to submit any PRs that you feel will make this better.

## Long term goals
Long term, I hope we won't need this, and the prune commands will be integrated with swarm mode, and then we can retire this tool.

## TODO
- Add golang vendoring
- make more robust
- make sure it works well with tls
