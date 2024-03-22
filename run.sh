#!/bin/bash

# docker run --name=redis -p 6379:6379 -d --rm redis

# if [ $? -eq 0 ]; then
#     echo "Redis build completed successfully."
# else
#     echo "Error during redis build. Exiting script."
#     exit 1
# fi

go build -o ./bin/main main.go

if [ $? -eq 0 ]; then
    echo "App build completed successfully."
else
    echo "Error during app build. Exiting script."
    exit 1
fi

./bin/main