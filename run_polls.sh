#!/bin/bash

nohup nsqlookupd &
nohup nsqd --lookupd-tcp-address=localhost:4160 &

cd counter
go build -o counter
nohup ./counter &

cd ../socailpoll
go build -o twittervotes
nohup ./twittervotes &

cd ../api
go build -o api
nohup ./api &

cd ../web
go build -o web
nohup ./web &
