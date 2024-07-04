#!/bin/bash


if ! hash docker 2>/dev/null; then
  echo "Docker Missing - Please install docker to run the development environment"
fi

cd ./docker || exit

docker compose down

docker compose up -d