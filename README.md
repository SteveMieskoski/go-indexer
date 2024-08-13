
A Golang familiarization project

## Ethereum Indexer in GO

A basic project to serve dual purposes of learning Golang and potentially building useful tool for the web3 community

## Project Structure
- `src` contains the indexer 
- `docker` contains docker compose files for building the docker environment
- `backend` contains an api server to serve indexed content to a frontend


## Docker Components
### Infrastructure
- Kafka broker
- Postgres DB
- Mongo DB
- Redis
### Blockchain
- Geth execution node
- Prysm validator node



## Getting Started
- `git clone https://github.com/SteveMieskoski/go-indexer.git`
- `cd go-indexer`
- Run script to start all the docker components
- `launchDevEnv.sh -a`
- If you have nodes to connect via RPC you can use `launchDevEnv.sh -i` to only launch the infrastructure docker components
- 

