A TCPNode is an independent peer which can operate in peer-2-peer environment.

The idea is to have both, a server and a client on same machine.

Imagine Node 1 (server1, client1) and a Node 2 (server2, client2)

- Initially, through a discovery mechanism, client1 dials to server2 and it sends the address of his server (server1) to server2.
- Upon receiving the server1's address, server2 passes it to his client2 through a channel so the client2 can Dial to server1 as well.
- This results into two node connected over TCP, able to process and share data.
- Each node maintains a map of active peers.
- illustration:
  ![overview](https://gist.github.com/user-attachments/assets/831fd408-5e98-4118-b784-75ba4b4d291c)
