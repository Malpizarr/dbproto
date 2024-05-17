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

# Building the Project

Navigate to the project directory:

    cd cmd/dbproto

Compile the project using:

    go build

# Using the package

This project is designed to be used as a package in other projects.

For installation, use the go get command:

    go get github.com/Malpizarr/dbproto

For usage in your project, import the following packages:

    github.com/Malpizarr/dbproto/pkg/data
    github.com/Malpizarr/dbproto/pkg/api
    github.com/Malpizarr/dbproto/pkg/utils

#### For better understanding on the usage of the package in other projects, an example can be found on [here](https://github.com/Malpizarr/dbprototests?tab=readme-ov-file).

# Running the CMD

Navigate to the cmd/dbproto directory:

    cd cmd/dbproto

Run the compiled binary:

    ./dbproto

### How to use the CMD

The CMD prvodes a CLI interface in order to see the information on the databases and tables, and to export the data to a file.

The following commands are available:

    list: Lists all databases.
        list [database]: Lists the tables on a database.
        list [database] [table]: Lists all the information on the table.

    export: Exports the data from a table to a file.
        export [database] [table] [filename] --format=[csv|xml]: Exports the data from a table to a file.


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


