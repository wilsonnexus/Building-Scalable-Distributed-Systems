# Distributed Databases using Replication

This repo contains two implementations of an in-memory distributed key-value database:

- `leader_follower/`
- `leaderless/`

It also contains:

- load testing code
- unit tests
- Dockerfiles
- docker-compose configs
- report notes
- AI usage statement

## API

### Write

`POST /set`

Body:

```json
{ "key": "k1", "value": "v1" }
```
