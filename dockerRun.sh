#!/bin/bash

### sample command file to run service in docker

docker run --name stone_scissors_paper -d -p 8080:8080 --env-file .env stone_scissors_paper