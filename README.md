# stone_scissors_paper

`stone_scissors_paper` is a game-service to play in "Rock paper scissors" classic game.

## Service requerements

The service requires Redis database connection. See the sample code that shows how to run Redis in Docker into `redisDockerRun.sh.sample`. NOTE: Don't forget to change the passsword for Redis in this file. Change "`<some very long password ...>`" to some random string without spaces.

And exactly the same random string have to be written into configuration file as value of ConnectOptions.Password.

Configuration file value ConnectOptions.Addrs have to contain at least one value in form "host:port" that points to host and port where the Redis server runs.

## Configuration

Configuration file can be made from `cnf.json.sample` sample file. But You have to change the values in it according to the running environment.

Configuration file have to be named as `cnf.json` and it should be in the same folder from where service executable is running.

### Configuration values

- `HostPort`: the value in form "host:port" that determines the host and port on which the service have to listen for requests.
- `ConnectOptions`: Redis database connection options:
    - `Addrs`: array of string values in form "host:port" that points to host and port where the Redis server runs.
    - `Password`: password for secure connection to Redis database
    - ... all possible connection options can be found [here](https://godoc.org/github.com/go-redis/redis#UniversalOptions)

## Building and running the docker image

Golang executable can run into docker image created as FROM SCRATCH (see example `dockerfile`). For this purpose the executable have to be build without dependencies to clib (CGO_ENABLED=0).

    CGO_ENABLED=0 go build
    docker build -f dockerfile --tag stone_scissors_paper
or

    ./build.sh
    docker build -f dockerfile --tag stone_scissors_paper

Then you can run service in docker by command (NOTE: Change `<full path to configuration file>` to real path to config in this command before run it):

    docker run --name stone_scissors_paper -d -p 8080:8080 -v <full path to configuration file>:/opt/game/cnf.json stone_scissors_paper

You also should correct exposed port (`-p 8080:8080`) according to `HostPort` configuration value.

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



