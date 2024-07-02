#!/bin/bash


cd ./docker || exit

docker compose down -d

docker compose up -d