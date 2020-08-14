# akhttpd -- Authorized Keys HTTP Daemon

This repository contains a small golang daemon that serves authorized_keys files for every GitHub user. 

This Daemon has two GET-only endpoints:

- `/<user>` - gets the keys of the user `user` in a format ready for `authorized_keys`
- `/<user>.sh` - gets a shell script the writes the file `$HOME/.ssh/authorized_keys` with the content above. 

This is intended to be used inside of Docker, and can be found as on DockerHub as [tkw01536/akhttpd](https://hub.docker.com/r/tkw01536/akhttpd/) as an automated build. 
To start it up run:

```
docker run -p 8080:8080 tkw01536/akhttpd
```

You can also use GitHub OAuth Token like so:

```
docker run -p 8080:8080 -e GITHUB_TOKEN=my-super-secret-token tkw01536/akhttpd
```

The code is licensed under the Unlicense, hence in the public domain. 