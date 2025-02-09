# Outbox Pattern Implementation

This repository contains an implementation of the Outbox Pattern, designed to ensure reliable message delivery in distributed systems. The implementation is modular, allowing for easy integration with different SQL databases and message brokers. In this example, **PostgreSQL** is used as the database store, and **NATS** is used as the message broker, as specified in the assignment.

## Repository Structure

The repository is divided into three main parts:

1. **Core Outbox Library**: The main library that implements the Outbox Pattern.
2. **Example Application**: Located in `cmd/example`, this application demonstrates how to use the Outbox library.
3. **Subscriber Application**: Located in `cmd/subscriber`, this application tests message reception from the NATS message broker.

## Features

- **Modular Design**: The implementation is designed to easily introduce new SQL databases (stores) and message brokers (brokers).
- **Test Coverage**: Comprehensive test cases are written for the core library to ensure reliability and correctness.
- **Leader/Replica Mode**: The example application supports leader and replica modes to simulate leader election and message processing.

## Prerequisites

Before running the application, ensure the following are installed on your system:

- **Docker** (with Docker Compose)
- **Go** (latest version)
- **Make** (for running commands)

## Setup and Running the Application

1. **Start Docker Containers**:
   - Run `make up` to start PostgreSQL and NATS Docker containers.
   - Run `make down` to stop both containers.

2. **Run Tests**:
   - Execute `make test` to run the test cases for the core Outbox library.

3. **Run the Example Application**:
   - Use `make run-leader` to run the example application in leader mode (simulating leader election and message processing).
   - Use `make run-replica` to run the example application in replica mode (no message processing).
   - Use `make run-subscriber` to run the subscriber application and verify that messages are being received correctly from the NATS message broker.

4. **Build the Application**:
   - Run `make build` to build the example application.

## Example Application Workflow

- The **leader instance** processes messages and writes them to the Outbox table in PostgreSQL.
- Messages are then published to the NATS message broker.
- The **subscriber application** listens to the NATS broker to verify that messages are being received correctly.

## Testing

The core library includes a suite of test cases to ensure the reliability and correctness of the Outbox implementation. Run the tests using the `make test` command.
