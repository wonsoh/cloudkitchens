# Order Simulation System
**Author:** Won Oh (wonsoh@live.com)
## Setup
### Setting up Go
This environment assumes you are using latest version of Go at the time of writing (version 1.18+).

Please visit [Go installation page](https://go.dev/doc/install).

#### Setting up `GOPATH`
Please make sure your environment variables are properly set for your shell profile (`.bashrc`, `.bash_profile`, `.zshrc`, `.profile`, etc.)
```sh
GOPATH=$HOME/go
```

### Setting up directories
The directory for this project needs to be set up under the source directory (`src`) of your `$GOPATH`. The project directory for this project is `$GOPATH/src/wonsoh.private/cloudkitchens`.
```sh
mkdir -p $GOPATH/src/wonsoh.private/cloudkitchens
```

Then, please copy all the contents in this directory into the project directory.
```sh
cp -r ./ $GOPATH/src/wonsoh.private/cloudkitchens
```

### Installing modules
Install all dependencies by running the following command:
```sh
go mod download
```

### Experimental Script for Project Setup
The project directory setup can be automated by running the following script
```sh
./install.sh
```

All the projects are set up.

## Running the project
There are two strategies you can use to run the simulation.
### Matched Order Strategy
This will run a simulation where a courier is dispatched for a specific order and may only pick up that order.

To run this strategy, run the following command:
```sh
./run_matched.sh
```

#### Design decision
We used a thread-safe map (concurrent hash map, or [`sync.Map`](https://pkg.go.dev/sync#Map)) for fast lookup (`O(1)`) (dictionaries, or maps, are good data structures for looking up). We also used channels for notifying between entities (order-to-courier and courier-to-order) and utilized [`sync.Mutex`](https://pkg.go.dev/sync#Mutex) to prevent deadlocks and for thread-safety.

Alternatively, we could use an infinite-loop (polling) with a sentinel value that breaks upon discovering a courier or order to be picked up from the map (by constantly checking if the value corresponding to the key exists in the map).

### FIFO Order Strategy
This will run a simulation where a courier picks up the next available order upon arrival. If there are multiple orders available, pick up an arbitrary order. If there are no available orders, couriers wait for the next available one. When there are multiple couriers waiting, the next available order is assigned to the earliest arrived courier.

To run this strategy, run the following command:
```sh
./run_fifo.sh
```

#### Design decision
We used a doubly-linked list ([`container/list`](https://pkg.go.dev/container/list) package) for fast-eviction (`O(1)`) and the nature of ordering (queues are good for FIFO ordering, and a linked-list is an optimal data structure to represent a queue). We also used channels for notifying between entities (order-to-courier and courier-to-order) and utilized [`sync.Mutex`](https://pkg.go.dev/sync#Mutex) to prevent deadlocks and for thread-safety.

Alternatively, we could use an infinite-loop (polling) with a sentinel value that breaks upon discovering a courier or order to be picked up from the queue (by constantly checking if an element exists in the queue).

## Testing
You can run comprehensive unit-tests that will run all unit tests and report the coverage for this project.

To run all the unit-tests and view the coverage, run the following command:
```sh
./test_with_coverage.sh
```
