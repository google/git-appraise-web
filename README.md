# Git-Appraise Web UI

This repository contains a read-only web UI for git-appraise reviews.

## Try it now

You can see a live demo of the UI running [here](https://git-appraise-web.appspot.com)

## Disclaimer

This is not an official Google product.

## Prerequisites

Building requires the Go tools and GNU Make. Running the built binary requires the git command line tool.

## Building the source code

Assuming you have the [Go tools installed](https://golang.org/doc/install), run
the following command:

    go get github.com/google/git-appraise-web/git-appraise-web

### Manual steps

Assuming you have not run the above command, first checkout the code from the git repo:

    mkdir -p ${GOPATH}/src/github.com/google
    git clone https://github.com/google/git-appraise-web.git

Build the binary:

    cd ${GOPATH}/src/github.com/google/git-appraise-web
    make

## Running the application

Binary is placed into `${GOPATH}/bin`:

    ${GOPATH}/bin/git-appraise-web

The tool requires that it be started in a directory that contains at least one git repo, and it shows the
reviews from every git repo under that directory.

The UI is a webserver which defaults to listening on port 8080. To use a different port, pass it as an argument to the "--port" flag:

    ${GOPATH}/bin/git-appraise-web --port=12345
