#!/bin/bash

# Run frontend
cd front
npm start
cd ..

# Run the go program with appropriate env vars
cd back
GOENV=dev go run .
cd ..