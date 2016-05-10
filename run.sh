#!/bin/bash
docker run --privileged --rm --name git -ti -v $(pwd):/go/src/github.com/docker/docker git-fsdriver /bin/bash
