#!/bin/bash


cd ./docker || exit

docker compose down -v

docker compose up -d