# Chit Chat readme

## Technical requirements

- The system must be implemented in Go

- The gRPC framework must be used for client–server communication,

  using Protocol Buffers (you must provide a .proto file) for message

  definitions.

- The Go log standard library must be used for structured logging of

  events (e.g. client join/leave notifications, message delivery, and

  server startup/shutdown).

- Concurrency (local to the server or the clients) must be handled

  using Go routines and channels for synchronised communication

  between components

- Every client and the server must be deployed as separate processes

- Each client connection must be served by a dedicated goroutine

  managed by the server.

- The system must support multiple concurrent client connections

  without blocking message delivery.

- The system must log the following events:

  - Server startup and shutdown

  - Client connection and disconnection events

  - Broadcast of join/leave messages

  - Message delivery events

- Log messages must include:

  - Timestamp

  - Component name (Server/Client)

  - Event type

  - Relevant identifiers (e.g. Client ID).

- The system can be started with at least three (3) nodes (two client

  and a server) and it must be able to handle "join" of at least one

  client and "leave" of at least one client

## Hand-in requirements

- you must hand in a report (single pdf file) via LearnIT

- you must provide a link to a Git repo with your source code in the

  report

- in the report, you must

  - discuss, whether you are going to use server-side streaming,

    client-side streaming, or bidirectional streaming?

  - describe your system architecture - do you have a server-client

    architecture, peer-to-peer, or something else?

  - describe what RPC methods are implemented, of what type, and what

    messages types are used for communication

  - describe how you have implemented the calculation of the timestamps

  - provide a diagram, that traces a sequence of RPC calls together

    with the Lamport timestamps, that corresponds to a chosen sequence

    of interactions: Client X joins, Client X Publishes, ..., Client X

    leaves.

- you must include system logs that document the requirements are met,

  in both the appendix of your report and your repo

- your repo must include a readme.md file that describes how to run

  your program.

- you repo must structured as follows:

  > project-root/

   >>client/ # contains the client code

  >>grpc/ # contains .proto file

  >>server/ # contains the server code

  >readme.md # readme file
