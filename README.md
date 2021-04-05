# stone_scissors_paper

`stone_scissors_paper` is a game-service to play in "Rock paper scissors" classic game.

## Service requerements

The service requires Redis database connection. See the sample code that shows how to run Redis in Docker into `redisDockerRun.sh.sample`. NOTE: Don't forget to change the passsword for Redis in this file. Change "`<some very long password ...>`" to some random string without spaces.

And exactly the same random string have to be written into configuration file as value of ConnectOptions.Password.

Configuration file value ConnectOptions.Addrs have to contain at least one value in form "host:port" that points to host and port where the Redis server runs.

Configuration file have to be named as `cnf.json` and it should be in the same folder from where service executable is running.

## Building the docker image

Golang executable can run into docker image created as FROM SCRATCH (see example `dockerfile`). For this purpose the executable have to be build without dependencies to clib (CGO_ENABLED=0).

    CGO_ENABLED=0 go build
    docker build -f dockerfile --tag stone_scissors_paper

 Then you can run service in docker by command (NOTE: change `<full path to configuration file>` to real path to config in this command before run it):

    docker run --name stone_scissors_paper -d -v <full path to configuration file>:/opt/game/cnf.json stone_scissors_paper

## Service API

### Request for new round:

URL: `<host>[:<port>]/new`

Method: `POST`

Response: `HTTP 200 OK` with body containing JSON with following parameters:

- `round`: round id
- `user1`: token of first user
- `user2`: token of second user


### Request for bid:

URL: `<host>[:<port>]/bid`

Method: `POST`

Request body: JSON with following parameter:

- `round`: round id
- `user`: token of the user that makes a bid
- `bid`: one of `paper`|`stone`|`scissors` 

Success response: `HTTP 200 OK` with body containing JSON with following parameter: 

- `response`: one of: `wait` - wait for second bid to be done, `you won ...`|`you lose ...`|`draw ...` - game result.

The game result also contains the current user and the rival's bids. 
 


### Request for results:
URL: `<host>[:<port>]/result` 

Method: `POST`

Request body: JSON with following parameter:

- `round`: round id
- `user`: token of the user that asks for result

Success response: `HTTP 200 OK` with body containing JSON with following parameter: 

- `response`: one of: `wait` - wait for all bids to be done, `you won...`|`you lose...`|`draw...` - game result.

THe game result also contains the current user and the rival's bids. 



