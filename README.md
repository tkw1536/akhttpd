# akhttpd - Authorized Keys HTTP Daemon

![CI Status](https://github.com/tkw1536/akhttpd/workflows/CI/badge.svg)

This repository contains a small go daemon that serves authorized_keys files for every GitHub user. 

This Daemon has two GET-only endpoints:

- `/<user>` - when called from a browser, same as `/<user>.html`, else `/<user>/authorized_keys`
- `/<user>/authorized_keys` - gets the keys of the user `user` in a format ready for `authorized_keys`
- `/<user>.html` - gets the keys of the user `user` and shows them in niceish html
- `/<user>.sh` - gets a shell script the writes the file `$HOME/.ssh/authorized_keys` with the content above. 

This is intended to be used inside of Docker, and can be found as [a GitHub Package](https://github.com/users/tkw1536/packages/container/package/akhttpd). 
To start it up run:

```
docker run -p 8080:8080 ghcr.io/tkw1536/akhttpd:latest
```

You can also use GitHub OAuth Token like so:

```
docker run -p 8080:8080 -e GITHUB_TOKEN=my-super-secret-token ghcr.io/tkw1536/akhttpd:latest
```

For a more detailed documentation, see [the godoc page](https://pkg.go.dev/github.com/tkw1536/akhttpd). 

## License
The code is licensed under the CC0 License, hence in the public domain. 