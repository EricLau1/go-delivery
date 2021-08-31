#!/bin/bash

docker run -it --rm --name mongodb_container \
-e MONGO_INITDB_ROOT_USERNAME=admin \
-e MONGO_INITDB_ROOT_PASSWORD=admin \
-v mongodata:/data/db -d -p 27017:27017 mongo
