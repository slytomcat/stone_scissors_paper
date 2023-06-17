[![Go](https://github.com/slytomcat/stone_scissors_paper/actions/workflows/go.yml/badge.svg)](https://github.com/slytomcat/stone_scissors_paper/actions/workflows/go.yml)

# stone_scissors_paper

`stone_scissors_paper` is a game-service to play in "Rock paper scissors" classic game.

## Service requirements

The service requires Redis database connection. See the sample code that shows how to run Redis in Docker into `redisDockerRun.sh.sample`. NOTE: Don't forget to change the password for Redis in this file. Change "`<some very long password ...>`" to some random string without spaces.

And exactly the same random string have to be set as environnement variable SSP_REDIS_PASSWORD.

The environnement variable SSP_REDIS_ADDRS have set to at least one value in form "host:port" that points to host and port where the Redis server runs. When more then one value provided (separated by comma) then the connection will be established to REDIS cluster.

## Configuration

Service configuration is provided through the environment variables.

### Configuration environment variables

- `SSP_HOST_PORT`: the value in form "host:port" that determines the host and port on which the service have to listen for requests. Default value is: `localhost:8080`
- `SSP_REDIS_ADDRS`: array of string values in form "host:port" that points to host and port where the Redis server runs.
- `SSP_REDIS_PASSWORD`: password for secure connection to Redis database
- `SSP_SERVER_SALT`: server salt for hashes (some random string without spaces)

## Building and running the docker image

Golang executable can run into docker image created as FROM SCRATCH (see `dockerfile`). For this purpose the executable have to be build without dependencies to clib (CGO_ENABLED=0). The `build.sh` script provides all necessary options to build the service executable.

    ./build.sh
    docker build -f dockerfile --tag stone_scissors_paper

Prepare `.env` file with the service configuration (see example in `.env.sample` file). Put the .env file in current directory.
Then you can run service in docker by command:

    docker run --name stone_scissors_paper -d -p 8080:8080 --env-file .env stone_scissors_paper

## Service API

### Request for new round:

URL: `<host>[:<port>]/new`

Method: `POST`

Request body: JSON with following parameter:

- `player`: identification for first player

Player can be identified by any string value: some user_id, e-mail or phone number. 

Response: `HTTP 200 OK` with body containing JSON with following parameters:

- `round`: round id


### Request attach to round:

URL: `<host>[:<port>]/attach`

Method: `POST`

Request body: JSON with following parameter:

- `round`: round id
- `player`: identification of second player

Round id must be one of received from `/new` request.
Player can be identified by any string value: some user_id, e-mail or phone number. 

Response: `HTTP 200 OK` with body containing JSON with following parameters:

- `response`: one of: 
    - `place Your bet, please` - response for successful attaching 
    - `You can't play with yourself` - the error message when player trying to attach to the round that was created by the player himself.
    - `this round is already full` - the error message when player trying to attach to the round that was already has two players.


### Request for placing a bet:

URL: `<host>[:<port>]/bet`

Method: `POST`

Request body: JSON with following parameter:

- `round`: round id
- `player`: identification of player that places the bet
- `bet`: hidden bet: hash made from bet 

The `bet` value should be calculated as:
1. Choice some secret. For example: `my secret`.
2. join the bet (one of `paper`|`stone`|`scissors`) and your secret in one string without delimiters. Example: `stonemy secret`.
3. compute sha256 hash from the string.
4. convert the resulting hash bytes to BASE64 URL safe encoding (using symbols `_-` instead of `/+` and without padding). 

For string `stonemy secret` you have to receive `L64zOtDB4yPHkd9ieLH8ghGdzDVn-_2X17Oo2bjDE64`

This also can be done in bash:

    echo -n 'stonemy secret' | openssl dgst -sha256 -binary | base64 | tr '/+' '_-' | tr -d '='

Success response: `HTTP 200 OK` with body containing JSON with following parameter: 

- `response`: one of: 
    - `wait for the rival to place its bet` 
    - `disclose your bet, please`
    - `bet has already been placed` - the error message when player trying to place more than one bet in the round.

When You receive `wait for the rival to place its bet` You should wait a little and make request for result. 
When You receive `disclose your bet, please` (as response form this request or as response from the status request) then You can make request for disclose bet


### Request for disclose bet:

URL: `<host>[:<port>]/disclose`

Method: `POST`

Request body: JSON with following parameter:

- `round`: round id
- `player`: Identification of player that places the bet
- `bet`: open bet (one of `paper`|`stone`|`scissors`) 
- `secret`: your secret, that was used for preparing the hidden bet (`my secret`)

Success response: `HTTP 200 OK` with body containing JSON with following parameter: 

- `response`: one of: 
    - `wait for your rival to disclose its bet` - you have to wait and make the result request to get the game result.
    - `you won ...`|`you lose ...`|`draw ...` - game result, it result also contains the current player and the rival's bets.
    - `Your bet is incorrect` - the error message when player provided not the same secret or bet that was used to calculate the hidden bet. Request for disclose bet can be repeated with the correct information.
 

### Request for results:
URL: `<host>[:<port>]/result` 

Method: `POST`

Request body: JSON with following parameter:

- `round`: round id
- `player`: identification of player that asks for result

Success response: `HTTP 200 OK` with body containing JSON with following parameter: 

- `response`: the same values as in responses on the request for bet and disclosure.

Some additional responses can be received in the requests for bet, disclose and result:

- `place Your bet, please` -  when round has been filled but player hadn't placed bet yet 
- `unauthorized` - the error message when player is not authorized to play in this round.
- `round had been falsificated` - the error message when the round information was falsificated. The falsificated round cannot be continued. 



