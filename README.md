# UNO Game

Online UNO Game.

## Features

- Create new game
- Join to a game
- Lobby to wait until game start
- Play card
- Draw card
- Exit game



## Test game
**Run server**

```
go run .
```

**Run client**

```
go run client/*
```

- Create a new game using command `create`
- Join a game using the command `join`
- Once the play started, use the command `play` plus `<card index>` number enclosed in []


## TODO

- Update GameState
  - Players information: Name, Id, Number of cards
  - Whow created the game
- Handle case when player reconnect to the game
- Send Game state on event "GameStarted"
- Send event when Deck is empty and game is draw
