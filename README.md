Hack the Valley 3 API
=====================
<p align="center">
  <img src="assets/logo.png"/>
</p>

[_powered by gqlgen_](https://gqlgen.com/)

## Pre-requisites:
1. Spin up MongoDB docker container:
    ```bash
    $ docker volume create mongo-data
    $ docker network create htv-net
    $ docker run -d --network htv-net -p 27017:27017 -v mongo-data:/data/db --name mongodb -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=password --restart always mongo
    ```
2. Spin up Redis docker container:
    ```bash
    $ docker volume create redis-data
    $ docker run --name redis --restart always -d -p 6379:6379 --network htv-net -v redis-data:/data redis redis-server --appendonly yes
    ```

## Instructions:
1. Install and configure Golang for your OS: [Here](https://golang.org/doc/install)
2. Clone this repository master branch
3. `cd` into the project root
4. Run the following commands to download the dependencies, 
verify them and vendor it locally for your projects:
    ```bash
    $ go mod download
    $ go mod tidy
    $ go mod verify
    $ go mod vendor
    ```
5. Set the following environment variables for the MLH Oauth middleware:
    ```bash
    export client_id=some_id 
    export client_secret=some_secret
    export scope=email+education+birthday
    export redirect_uri=http://localhost:8080/v1/auth/callback
    ```
6. Run the following command to start up the graphql playground:
    ```bash
    $ go run ./server/server.go
    ```
7. Modify `resolver.go` as you go along to add/modify features.
8. Now on your frontend, redirect the user to the following url to initiate the flow<br/>
(provided you have added `http://localhost:8080/v1/auth/callback` to the redirect url section for your mlh client id)
    ```
    https://my.mlh.io/oauth/authorize?client_id=some_client_id&redirect_uri=http://localhost:8080/v1/auth/callback&response_type=code&scope=email+education+birthday
    ```
8. For more information on how gqlgen works, check out the: [docs](https://gqlgen.com/getting-started/)
