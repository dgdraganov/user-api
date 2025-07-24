# user-api

Simple API to manage users and their files.


## What is the service tech description? 

`go service` that exposes a REST API 

`mysql database` to keep the users and the files metadata

`minio` service to use as a blobstore for the actual files

`rabbitMQ` to send user events

`c# consumer` to consume events


## How to run it?

There is a docker copose file with all the needed services. Use the make target `up`:

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

There are two admin users that are seeded with on app start (if the DB is empty):

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


Operations `update`, `delete` for user records and `upload`, `remove` for files are allowed if the user is updating his/her record or files or the user is admin. Admins can delete or update users and their files.


## Any tests?

The tests can be run with the following make target:

```bash
    make test
```