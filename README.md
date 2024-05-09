# PureGamer: An Experimental Distributed Game Accelerator with Automatic Routing

## Introduction
PureGamer is an experimental distributed game accelerator with automatic routing.
It is designed to accelerate online games by routing game traffic through the shortest path between the game server and the player.
PureGamer is built on top of the [LibP2P](https://libp2p.io/) networking stack.

## Features
- Automatic routing: PureGamer automatically choose the route with the lowest latency between the game server and the player.
- Distributed: PureGamer is a distributed system that can be run on multiple nodes.
- Scalable: PureGamer can be easily scaled by adding more nodes to the network.
- Secure: PureGamer uses end-to-end encryption to secure the game traffic.
- Open-source: PureGamer is an open-source project that is free to use and modify.

## How it works
PureGamer use LibP2P to create a peer-to-peer network.
It used Gossip PubSub to broadcast the latency information between the nodes.
Nodes receive the latency information from other nodes and use it to calculate the shortest path between the game server and the entry node.
It supports multiple transport protocols, such as TCP and QUIC, thanks to LibP2P.

## Getting started
### Build from source
```bash
go build
```

### Run the PureGamer node
```bash
./puregamer
```

## License
PureGamer is licensed under the [MIT License](LICENSE).
