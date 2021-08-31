# MongoDB 

## Run

```bash
docker run -it --rm --name mongodb_container \
-e MONGO_INITDB_ROOT_USERNAME=admin \
-e MONGO_INITDB_ROOT_PASSWORD=admin \
-v mongodata:/data/db -d -p 27017:27017 mongo
```

### Login

```bash
docker exec -it mongodb_container bash

mongo -u admin -p admin --authenticationDatabase admin
```

### Create Database

```bash
use godelivery
````

### Create User

```bash
db.createUser({user: 'root', pwd: 'root', roles:[{'role': 'readWrite', 'db': 'godelivery'}]});

show users
```

### Login with New User

```bash
mongo -u root -p root --authenticationDatabase godelivery

use godelivery

show collections
```

### Clear Docker Volumes

```bash
docker volume rm $(docker volume ls -qf dangling=true)
```