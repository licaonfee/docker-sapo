# Docker Sapo

Watch docker events and do things

- [Docker Sapo](#docker-sapo)
  - [Build](#build)
  - [Usage](#usage)

## Build

```bash
go mod download
cd cmd
go build -o docker-sapo .
```

## Usage

```bash
#On container stop start mycontainer
docker-sapo -filter="container=mycontainer" -filter="event=die" -cmd="docker start mycontainer"
```
