# Multiple Services Example

This is a fairly bare-bones example of how you would structure a
project with multiple services. In this example, we've got a `GameService`
which just manages a catalog of games and a `ScoreService` that tracks
high score leaderboard info for those games.

In addition to our main function using the `GameService`, the
`ScoreService` is actually dependent on it as well. When
posting/listing high scores, it uses the game service to verify that
the game in question actually exists before doing its "real" work.

To run this example, just execute the following commands from
this directory. You can look at the `makefile` to see what commands
it's running under the hood... it's fairly simple.

```shell
# In terminal 1
make game-service

# In terminal 2
make score-service

# In terminal 3
make run
```

If everything is working properly, you should see the output:

```
Game 1 = Super Mario Bros.
Game 2 = The Witcher 3: Wild Hunt
New Score = 9393311
High Score 1 = 9393311 (Red Luigi)
High Score 2 = 20332 (Dog Man)
High Score 3 = 999 (Wolverine)
```

## Consume Using curl

When you execute `make run`, it's just using the strongly-typed Go clients
to make API calls under the hood. You can make the exact same requests
using raw HTTP if you want. Just make sure that both services are running
first:

```shell
# In terminal 1
make game-service

# In terminal 2
make score-service

# In terminal 3
curl http://localhost:9001/v2/game/1
curl -d '{"Name":"The Witcher 3: Wild Hunt", "Publisher":"CD Projekt RED"}' http://localhost:9001/v2/game
curl -d '{"PlayerName":"Red Luigi", "Score":9393311}' http://localhost:9002/v2/game/1/highscore
curl http://localhost:9002/v2/game/1/highscore?howMany=3
```

When in your Go code, the clients are much more convenient, but this
shows you that you can customize your routes and consume your service using any language.

## What This Example Shows

* How to structure your code by putting services in their own packages.
* How one service can invoke functions on another using the Frodo-generated client.
* How you can consume multiple services in the same program.
* Use doc options to customize RESTful paths for all operations.
* You can use json tags to customize marshaling structure.
* How to generate 4XX status error codes using the `rpc/errors` package.
* Reduce verbosity by using aliases when creating identical request/response structs.
