## Overview

The dbproto project is a Go application that provides a simple server and utilities for database operations, including table creation, data encryption/decryption, and basic CRUD operations on records. It leverages environment variables for configuration, Protobuf for data serialization, and integrates AES for security.

# Getting Started

Prerequisites

    Go 1.15 or higher
    Protobuf compiler (protoc)
    github.com/joho/godotenv for loading environment variables
    github.com/Malpizarr/dbproto/pkg/data and github.com/Malpizarr/dbproto/pkg/api for the core functionality
    github.com/Malpizarr/dbproto/pkg/utils for encryption utilities
    AES_KEY environment variable for encryption key

# Environment Variable Setup Guide for AES Key

This README guide provides instructions on how to set up an AES key as an environment variable on macOS, Linux, and Windows using Zsh, Bash, and PowerShell. This includes generating the key with OpenSSL.

Requirements

- OpenSSL must be installed on your system.
- Access to the terminal or command line interface.

Setting Up AES Key Environment Variable

### macOS and Linux

#### Bash

Open your terminal and execute the following commands:

```bash
echo "export AES_KEY=$(openssl rand -hex 16)" >> ~/.bashrc
source ~/.bashrc
```

For a system-wide setup (all users), you might prefer:

```bash
echo "export AES_KEY=$(openssl rand -hex 16)" | sudo tee -a /etc/profile
```

#### Zsh

For Zsh users, execute the following commands:

```bash

echo "export AES_KEY=$(openssl rand -hex 16)" >> ~/.zshrc
source ~/.zshrc
```

### Windows

#### PowerShell

Open PowerShell as Administrator and run:

For system-wide environment variable:

```powershell
$aesKey = openssl rand -hex 16
[System.Environment]::SetEnvironmentVariable('AES_KEY', $aesKey, [System.EnvironmentVariableTarget]::Machine)
```

For the current user only:

```powershell

$aesKey = openssl rand -hex 16
[System.Environment]::SetEnvironmentVariable('AES_KEY', $aesKey, [System.EnvironmentVariableTarget]::User)
```

# Installation

Clone the repository:

    git clone https://github.com/Malpizarr/dbproto.git

Navigate to the project directory:

    cd dbproto

Install dependencies:

    go get .

# Configuration

Set up the required environment variables. Create a .env file in the root of your project and specify the following variables:

    AES_KEY: A 32-byte key used for AES encryption.

# Building the Project

Navigate to the project directory:

    cd cmd/dbproto

Compile the project using:

    go build

# Using the package

This package provides a simple server for database operations, including table creation, data encryption/decryption, and basic CRUD operations on records. The server listens for HTTP requests and responds with JSON payloads. The server uses environment variables for configuration and Protobuf for data serialization.

For usage in your project, import the following packages:

    github.com/Malpizarr/dbproto/pkg/data
    github.com/Malpizarr/dbproto/pkg/api
    github.com/Malpizarr/dbproto/pkg/utils

# Running the Server

Start the server by executing the compiled binary:

    ./dbproto

This starts the server on localhost:8080. The server logs will indicate that it is running and listening for requests.

# API Overview

Database Operations

    Create Database: POST /createDatabase
        Requires a JSON payload with the database name.
    List Databases: GET /listDatabases
        Returns a list of all databases.

Table Operations

    Create Table: POST /createTable?dbName=<database_name>
    Requires a JSON payload with the table name and primary key.

    Table Actions: POST /tableAction?dbName=<database_name>
    Requires a JSON payload with the action (insert, update, delete, selectAll), table name, and record data.

Join Operation

    Join Tables: POST /joinTables?dbName=<database_name>
    Requires a JSON payload with the table names, join keys, and join type (innerJoin, leftJoin, rightJoin, fullOuterJoin).

# Encryption Utilities

The Utils module in utils package provides methods for:

    Encrypting and decrypting data using AES in Counter mode (CTR).
    Initialization of data encryption keys and their secure storage after encryption.

# Protobuf Definitions

Defines records and record collections for serialization:

    Record: A simple map of string pairs.
    Records: A collection of Record objects.

# Transaction Management

The data package includes a transaction mechanism for performing CRUD operations on tables. The Transaction struct stores the original records before any changes, and the provided methods (InsertWithTransaction, UpdateWithTransaction, DeleteWithTransaction) ensure that either all changes are committed or rolled back, maintaining data consistency.

```

```

```

```

```

```

```

```
