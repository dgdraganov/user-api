# user-api

Simple API to manage users and their files.


## What is the service architecture? 

In the center there is a `go service` that exposes a REST API - `user-api`. It provides several endpoints for performing CRUD operations with users and user files: 

```
    POST    /api/auth               - authenticates an user by generating a JWT token 
	GET     /api/users/{guid}       - get a single user record from the DB
	GET     /api/users              - get paginated list of users from the DB
	POST    /api/users              - create/register a new user
	PUT     /api/users/{guid}       - update an existing user record
	DELETE  /api/users/{guid}       - delete an user
	POST    /api/users/file         - uploads a file
	GET     /api/users/{guid}/files - get a list of all user files
	DELETE  /api/users/{guid}/files - delete all the files for a scpecific user
```

The users and the files' metadata are stored in `MySQL` DB via the `gorm` ORM package. The files themselves are stored in a blobstore - `minio`. `Minio` is very convenient tool as later in the development phase it can easily be swapped with `Amazon S3 bucket` due to the APIs being very similar. You can check the Web UI on `http://localhost:9001/`. 

Creating, updating and deleting an user will trigger an event to be sent to `rabbitMQ`. Web UI version available on `http://localhost:15672/`. Finally there is the `c# consumer`. It is a limple console application that consumes messages from the respective topic.


## How to run it?

All of the services are described in the `docker-compose.yaml` file. Simply run the make target `up`:

```bash
    make up
```

To bring the services down: 

```bash
    make down
```


## How to use it?

In directory `curl_requests` you can find all the needed curl requests to play with the API. 

```bash
curl_requests
├── authenticate.sh
├── delete_user_files.sh
├── delete_user.sh
├── get_user.sh
├── list_files.sh
├── list_users.sh
├── register_user.sh
├── test1.txt
├── test.txt
├── update_user.sh
└── upload_file.sh

```

Listing users does not require authentication. You can list existing users with: 

```bash
    # page 1 with 33 users per page
    ./list_users.sh 1 33
```

Use the below script to authenticate with a specific user. 

```bash
    # this user is an admin user and it is seeded by default
    source ./authenticate alice@example.com alice
```

The script will load the JWT token as an env variable for the other scripts to use.

There are two admin users that are seeded on app start (if the DB is empty). You can use them to authenticate with an admin role. 

```
Email               Password

alice@example.com   alice
bob@example.com     bob
```

Run a script with no arguments and it will give you a hint on how to use it: 

```bash
    ./update_user.sh
```
The above command will print the correct usage: 

```bash
    Usage: ./update_user.sh <user_id> <first_name> <age>
```

This update_user script requires first name and age just to illustrate how the endpoint updates user data. The actual endpoint can update all user parameters except for the user `role` and `id`. 

Operations `update`, `delete` for user records and `upload`, `remove` for files are allowed if the user is updating his/her record or files or if the user is an admin. Admins can delete or update users and their files.


## Any tests?

The tests can be run with the following make target:

```bash
    make test
```