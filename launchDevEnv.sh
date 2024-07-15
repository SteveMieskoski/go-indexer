#!/bin/bash


if ! hash docker 2>/dev/null; then
  echo "Docker Missing - Please install docker to run the development environment"
  exit
fi

if ! pgrep docker 2>/dev/null; then
  echo "Please start Docker before running this script. Docker Process Not Found"
  exit
fi

#cd ./docker || exit
#
#docker compose down
#
#docker compose up -d
BLOCKCHAIN=0
INFRA=0
FUZZ=0
FUZZTWO=0
FUZZBLOB=0
FUZZOTHER=0

POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -b)
    BLOCKCHAIN=1
    shift # past argument
    ;;
    -i)
    INFRA=1
    shift # past argument
    ;;
    -f)
    FUZZ=1
    shift # past argument
    ;;
    -f2)
    FUZZTWO=1
    shift # past argument
    ;;
    -fb)
    FUZZBLOB=1
    shift # past argument
    ;;
    -fo)
    FUZZOTHER=1
    shift # past argument
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done

set -- "${POSITIONAL[@]}"

if [ $BLOCKCHAIN -eq 0 ]; then
  if [ $INFRA -eq 0 ]; then
    if [ $FUZZ -eq 0 ]; then
      if [ $FUZZBLOB -eq 0 ]; then
        if [ $FUZZOTHER -eq 0 ]; then
          BLOCKCHAIN=1
          INFRA=1
        fi
      fi
    fi
  fi
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

if [[ $FUZZ -ne 0 ]]; then

  cd ./fuzz || exit

  ./livefuzzer spam --sk 0x62014863ea920a2fe27631a7f6b2732e2a554787e3f564a2f9113b351dc7c4ec --slot-time 20 --gaslimit 50000 --accounts 20 --txcount 2

  cd .. || exit
fi

if [[ $FUZZTWO -ne 0 ]]; then

  cd ./fuzz || exit

  ./livefuzzer spam --sk 0x4f9c794dba16f0a3e252b8e202c824c17ad6252b8424d1d17094717184461e8b --slot-time 20 --gaslimit 50000 --accounts 20 --txcount 2

  cd .. || exit
fi

if [[ $FUZZBLOB -ne 0 ]]; then

  cd ./spammer || exit

   ./blob-spammer combined -p 62014863ea920a2fe27631a7f6b2732e2a554787e3f564a2f9113b351dc7c4ec -h http://localhost:8545 -b 2 -t 3 --max-pending 3

  cd .. || exit
fi

if [[ $FUZZOTHER -ne 0 ]]; then

  cd ./fuzz || exit

  ./livefuzzer blobs --accounts 20 --txcount 2

  cd .. || exit
fi



#cd ..
#
#cd ./docker/infrastructure || exit
#
#docker compose down
#
#docker compose up -d

