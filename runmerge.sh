#!/bin/bash
docker run --privileged --rm --name merge -ti -v $(pwd):/go/src/github.com/docker/docker merge-com /bin/bash
