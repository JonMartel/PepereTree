# PepereTree
Tool for parsing gedcom files and hosting the results


Notes for setting things up
- Installer docker
- Install dgraph via `docker pull dgraph/dgraph`


# Setup for dgraph:

`mkdir -p ~/dgraph`

Now, we need to run dgraph zero
Note: I used 8081 rather than 8080 as I had a service with that port already open. Still use 8080 internally
So the ratel functionality still works

`docker run -it -p 5080:5080 -p 6080:6080 -p 8081:8080 -p 9080:9080 -p 8000:8000 -v ~/dgraph:/dgraph --name dgraph dgraph/dgraph dgraph zero`

In another terminal, now run dgraph

`docker exec -it dgraph dgraph alpha --lru_mb 2048 --zero localhost:5080`

And in another, run ratel (Dgraph UI)

`docker exec -it dgraph dgraph-ratel`

