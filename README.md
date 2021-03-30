# stone_scissors_paper

`stone_scissors_paper` is a game-service to play in "Rock paper scissors" classic game.

The service requires Redis database connection. See example how to run Redis in Docker in redisDockerRun.sh


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

- `response`: one of: `wait` - wait for second bid to be done, `you won`|`you lose`|`draw` - game result.
 


### Redirect for results:
URL: `<host>[:<port>]/result` 

Method: `POST`

Request body: JSON with following parameter:

- `round`: round id
- `user`: token of the user that asks for result

Success response: `HTTP 200 OK` with body containing JSON with following parameter: 

- `response`: one of: `wait` - wait for second bid to be done, `you won`|`you lose`|`draw` - game result.


