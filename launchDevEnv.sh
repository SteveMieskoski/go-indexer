#!/bin/bash


if ! hash docker 2>/dev/null; then
  echo "Docker Missing - Please install docker to run the development environment"
  exit
fi

if ! pgrep docker 2>/dev/null; then
  echo "Please start Docker before running this script. Docker Process Not Found"
  exit
fi


help(){
  echo "Script to launch docker for local geth blockchain and indexer infrastructure components "
  echo "Usage: launchDevEnv.sh [-a] [-i] [-b]"
  echo " -a - launch both local blockchain and indexer infrastructure in docker"
  echo " -b - only launch local blockchain in docker"
  echo " -i - only launch local indexer infrastructure in docker"
  echo " -h - show this help"
}

BLOCKCHAIN=0
INFRA=0
HELP=0

POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -a)
    BLOCKCHAIN=1
    INFRA=1
    shift
    ;;
    -b)
    BLOCKCHAIN=1
    shift
    ;;
    -i)
    INFRA=1
    shift
    ;;
    -h)
    HELP=1
    shift
    ;;
    *)
    POSITIONAL+=("$1")
    shift
    ;;
esac
done

set -- "${POSITIONAL[@]}"

if [ $BLOCKCHAIN -eq 0 ]; then
  if [ $INFRA -eq 0 ]; then
    help
    exit
  fi
fi

if [[ HELP -ne 0 ]]; then
    help
    exit
fi

cd ./docker || exit

if [ $BLOCKCHAIN -ne 0 ]; then
  cd ./blockchain || exit

  sudo rm -rf ./consensus/beacondata
  sudo rm -rf ./consensus/validatordata
  sudo rm -rf ./consensus/genesis.ssz
  sudo rm -rf ./execution/geth

  docker compose down
  docker compose up -d

  cd .. || exit
fi

if [[ $INFRA -ne 0 ]]; then

  cd ./infrastructure || exit

  docker compose down
  docker compose up -d

  cd .. || exit
fi

