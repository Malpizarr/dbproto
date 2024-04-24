Overview

The dbproto project is a Go application that provides a simple server and utilities for database operations, including table creation, data encryption/decryption, and basic CRUD operations on records. It leverages environment variables for configuration, Protobuf for data serialization, and integrates AES for security.
Getting Started
Prerequisites

    Go 1.15 or higher
    Protobuf compiler (protoc)
    github.com/joho/godotenv for loading environment variables
    github.com/Malpizarr/dbproto/data and github.com/Malpizarr/dbproto/api for the core functionality
    github.com/Malpizarr/dbproto/utils for encryption utilities

Installation

    Clone the repository:

    bash

git clone https://github.com/Malpizarr/dbproto.git

Navigate to the project directory:

bash

cd dbproto

Install dependencies:

bash

    go get .

Configuration

Set up the required environment variables. Create a .env file in the root of your project and specify the following variables:

    MASTER_AES_KEY: A 32-byte key used for AES encryption.

Building the Project

Compile the project using:

bash

go build

Running the Server

Start the server by executing the compiled binary:

bash

./dbproto

This starts the server on localhost:8080. The server logs will indicate that it is running and listening for requests.
API Overview
Database Operations

    Create Database: POST /createDatabase
        Requires a JSON payload with the database name.
    List Databases: GET /listDatabases
        Returns a list of all databases.

Encryption Utilities

The Utils module in utils package provides methods for:

    Encrypting and decrypting data using AES in Counter mode (CTR).
    Initialization of data encryption keys and their secure storage after encryption.

Protobuf Definitions

Defines records and record collections for serialization:

    Record: A simple map of string pairs.
    Records: A collection of Record objects.
