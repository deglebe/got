# got

a terminal-based git staging manager built with the charm tui stack and golang

## features

- select and (un)stage files
- integration with github via pat/oauth for remote repository initialization
- local repo init
- create well-formatted commits

## installation

```sh
git clone https://github.com/deglebe/stage
cd stage
make install
```

you should review the install location before using the quick install

## config

right now the only config is your github pat. when creating a pat, alow repo and workflow fields.
