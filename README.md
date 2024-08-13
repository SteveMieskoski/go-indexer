
**A Golang familiarization project**

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
- Run `make` to build go-indexer
- Use `./build/bin/go-indexer` to start go-indexer as a producer extracting blockchain data from the RPC endpoints in the env file
- Use `./build/bin/go-indexer --run-as-producer=false` to start go-indexer as a consumer to parse the extracted blockchain data into the DBs 


