#!/bin/bash

### sample command file to run service in docker

docker run --name stone_scissors_paper -d -v "<full path to config file>":/opt/game/cnf.json "<image name>"