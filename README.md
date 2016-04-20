# Git-Appraise Web UI

This repository contains a read-only web UI for git-appraise reviews.

## Disclaimer

This is not an official Google product.

## Prerequisites

Building requires the Go tools and GNU Make. Running the built binary requires the git command line tool.

## Building the source code

First checkout the code from the git repo:

    git clone git@github.com:google/git-appraise-web.git

Build the binary:

    make

And then launch it:

    ${GOPATH}/bin/git-appraise-web

The tool requires that it be started in a directory that contains at least one git repo, and it shows the
reviews from every git repo under that directory.

The UI is a webserver which defaults to listening on port 8080. To use a different port, pass it as an argument to the "--port" flag:

    ${GOPATH}/bin/git-appraise-web --port=12345
