[![Go](https://github.com/slytomcat/stone_scissors_paper/actions/workflows/go.yml/badge.svg)](https://github.com/slytomcat/stone_scissors_paper/actions/workflows/go.yml)

# stone_scissors_paper

`stone_scissors_paper` is a game-service to play in "Rock paper scissors" classic game.

## Service requerements

The service requires Redis database connection. See the sample code that shows how to run Redis in Docker into `redisDockerRun.sh.sample`. NOTE: Don't forget to change the passsword for Redis in this file. Change "`<some very long password ...>`" to some random string without spaces.

And exactly the same random string have to be set as environnment variable SSP_REDISPASSWORD.

The environnment variable SSP_REDISADDRS have set to at least one value in form "host:port" that points to host and port where the Redis server runs.

## Configuration

Service configuration is provided through the environment variables.

### Configuration environment variables

- `SSP_HOSTPORT`: the value in form "host:port" that determines the host and port on which the service have to listen for requests. Default value is: `localhost:8080`
- `SSP_REDISADDRS`: array of string values in form "host:port" that points to host and port where the Redis server runs.
- `SSP_REDISPASSWORD`: password for secure connection to Redis database

## Building and running the docker image

Golang executable can run into docker image created as FROM SCRATCH (see example `dockerfile`). For this purpose the executable have to be build without dependencies to clib (CGO_ENABLED=0). The `bild.sh` script provides all necessary options to build the service executable.

    ./build.sh
    docker build -f dockerfile --tag stone_scissors_paper

Prepare `.env` file with the service configuration (see example in `.env.sample` file). Put the .env file in current directory.
Then you can run service in docker by command:

    docker run --name stone_scissors_paper -d -p 8080:8080 --env-file .env stone_scissors_paper

## Service API

### Request for new round:

URL: `<host>[:<port>]/new`

Method: `POST`

Response: `HTTP 200 OK` with body containing JSON with following parameters:

- `round`: round id
- `user1`: token of first user
- `user2`: token of second user


### Request for placing a bet:

URL: `<host>[:<port>]/bet`

Method: `POST`

Request body: JSON with following parameter:

- `round`: round id
- `user`: token of the user that places a bet
- `bet`: one of `paper`|`stone`|`scissors` 

Success response: `HTTP 200 OK` with body containing JSON with following parameter: 

- `response`: one of: 
    - `wait` - wait for the second bet to be placed
    - `you won ...`|`you lose ...`|`draw ...` - game result, it result also contains the current user and the rival's bets.
    - `bet has already been placed` - the error message when player trying to place more than one bet in the round.
    - `unauthorized` - the error message when player is not authorized to place a bet.
 


### Request for results:
URL: `<host>[:<port>]/result` 

Method: `POST`

Request body: JSON with following parameter:

- `round`: round id
- `user`: token of the user that asks for result

Success response: `HTTP 200 OK` with body containing JSON with following parameter: 

- `response`: the same values as in response on the request for bet.



