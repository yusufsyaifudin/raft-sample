# Raft Example

Straightforward implementation of Raft Consensus

## Why?

Copying statement from [another raft repo example](https://github.com/yongman/leto) that

> You can have better comprehension about how `raft protocol` works if you use it. 

And, yes! This is another example of implementing Raft using BadgerDB.

## What the difference with another example repo?

Here I try to create the very basic example, where you can create multi leader server only by running it separately using different port.
Then, you can select one server as the Raft Leader and connect it manually using CURL or via Postman.
This will give you better understanding how to create Raft Cluster rather than another repo which already included `join` command via program's argument or config.

This example also show you how we can create the multiple server read with eventual consistency. 
This means that we can read data from any server, but only can do writes and deletes operation through leader server.

## How to run?

`git clone` this project or download prebuilt executable files in Release, then set three (3) different config, for example:

Then run 3 server with different program in different terminal tab:

```bash
$ SERVER_PORT=2221 RAFT_NODE_ID=node1 RAFT_PORT=1111 RAFT_VOL_DIR=node_1_data go run ysf/raftsample/cmd/api
$ SERVER_PORT=2222 RAFT_NODE_ID=node2 RAFT_PORT=1112 RAFT_VOL_DIR=node_2_data go run ysf/raftsample/cmd/api
$ SERVER_PORT=2223 RAFT_NODE_ID=node3 RAFT_PORT=1113 RAFT_VOL_DIR=node_3_data go run ysf/raftsample/cmd/api
```

Or using prebuilt executable:

```bash
$ SERVER_PORT=2221 RAFT_NODE_ID=node1 RAFT_PORT=1111 RAFT_VOL_DIR=node_1_data ./raftsample
$ SERVER_PORT=2222 RAFT_NODE_ID=node2 RAFT_PORT=1112 RAFT_VOL_DIR=node_2_data ./raftsample
$ SERVER_PORT=2223 RAFT_NODE_ID=node3 RAFT_PORT=1113 RAFT_VOL_DIR=node_3_data ./raftsample
```

## Creating clusters

After running the each server, we have 3 servers:

* http://localhost:2221 with raft server localhost:1111
* http://localhost:2222 with raft server localhost:1112
* http://localhost:2223 with raft server localhost:1113

We can check using `/raft/stats` for each server and see that all server initiated as Leader.

Now, manually pick one server as the real Leader, for example http://localhost:2221 with raft server localhost:1111.
Using Postman, we can register http://localhost:2222 as a Follower to http://localhost:2221 as a Leader.

```curl
curl --location --request POST 'localhost:2221/raft/join' \
--header 'Content-Type: application/json' \
--data-raw '{
	"node_id": "node_2", 
	"raft_address": "127.0.0.1:1112"
}'
```

And doing the same to register http://localhost:2223 as a Follower to http://localhost:2221 as a Leader:

```curl
curl --location --request POST 'localhost:2221/raft/join' \
--header 'Content-Type: application/json' \
--data-raw '{
	"node_id": "node_3", 
	"raft_address": "127.0.0.1:1113"
}'
```

> What happen when we do cURL?
>
> When we running the cURL, we send the data of `node_id` and `raft_address` that being registered as a Voter.
> We say `Voter` because we don't know the real Leader yet.
> 
>
> In server http://localhost:2221 it will add the configuration stating that http://localhost:2222 and http://localhost:2223
> now is a Voter.
> After add the Voter, raft will choose the server http://localhost:2221 as the Leader.
>
> Adding Voter must be done in Leader server, that's why we always send to the same server for adding server.
> You can see that we always call port 2221 both for adding port 2222 or 2223

Then, check each of this endpoint, it will return the status that the port 2221 is now the only leader and the other is just a follower:

* http://localhost:2221/raft/stats
* http://localhost:2222/raft/stats
* http://localhost:2223/raft/stats

Now, raft cluster already created!

## Using Docker

First, build the image using command: `docker build -t ysf/raftsample .`

Then, run using docker compose `docker-compose up`.

To connect between cluster, use docker gateway IP, see using `docker network inspect bridge`,
so instead of 

```curl
curl --location --request POST 'localhost:2221/raft/join' \
--header 'Content-Type: application/json' \
--data-raw '{
	"node_id": "node_2", 
	"raft_address": "127.0.0.1:1112"
}'
```

You must change the `127.0.0.1` to Bridge IP from docker inspect command, for example:

```curl
curl --location --request POST 'localhost:2221/raft/join' \
--header 'Content-Type: application/json' \
--data-raw '{
	"node_id": "node_2", 
	"raft_address": "172.17.0.1:1112"
}'
```

## Store, Get and Delete Data

As already mentioned before, this cluster will create a simple distributed KV storage with eventual consistency in read.
This means, all writes command (Store and Delete) **must** redirected to the Leader server, since the Leader server is the only one
that can do `Apply` in raft protocol. After doing Store and Delete, we can make sure that the Raft already committed the message to all Follower servers.

Then, in `Get` method in order to fetch data, we can use the internal database instead calling `raft.Apply`. 
This makes all Get command can be targeted to any server, not only the Leader.

So, why we call it _eventual consistency in read_ while we can make sure that every after Store and Delete response returned it means that the raft already applied the logs to n quorum servers?

That is because while reading data directly in badgerDB we only use read transaction. From BadgerDB's Readme:

> You cannot perform any writes or deletes within this transaction. Badger ensures that you get a consistent view of the database within this closure. Any writes that happen elsewhere after the transaction has started, will not be seen by calls made within the closure.

To do store data, use this cURL (change `raft.leader.server` to the Leader HTTP address, in this example http://localhost:2221):

```curl
curl --location --request POST 'raft.leader.server/store' \
--header 'Content-Type: application/json' \
--data-raw '{
	"key": "key",
	"value": "value"
}'
```

To get data, use this (change `any.raft.server` to any HTTP address, it can be port 2221, 2222 or 2223):

```curl
curl --location --request GET 'any.raft.server/store/key'
```

To delete data, use this (change `raft.leader.server` to the Leader HTTP address, in this example http://localhost:2221):

```curl
curl --location --request DELETE 'raft.leader.server/store/key'
```
