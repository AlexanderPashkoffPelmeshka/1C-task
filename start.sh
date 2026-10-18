#!/bin/bash

cd bin
go mod tidy
go build -o ../diary .
cd ..
./diary
