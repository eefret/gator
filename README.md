# Gator CLI

Gator is a command-line application for managing feeds. With Gator, you can add feeds, follow or unfollow them, and view the feeds you follow. This tool uses PostgreSQL as its database backend and is written in Go.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go:** [Download and install Go](https://golang.org/dl/) (version 1.17 or later is recommended).
- **PostgreSQL:** [Download and install PostgreSQL](https://www.postgresql.org/download/) and make sure your PostgreSQL server is running.

## Installation

To install the Gator CLI, use the `go install` command. For example, if your repository is hosted on GitHub under your username, run:

```bash
go install github.com/eefret/gator@latest
```

## Configuration

Gator uses a configuration file to connect to your PostgreSQL database and manage other settings. Follow these steps to set up your configuration:

1. Create a file named config.yaml in the root of your project.

2. Populate it with your configuration settings. For example:
```yaml
# config.yaml

database:
  host: "localhost"
  port: 5432
  user: "your_db_user"
  password: "your_db_password"
  dbname: "gator_db"
```
Adjust the database settings as needed for your environment.

## Running the Program
Once installed and configured, you can run the Gator CLI from your terminal. For example, to start the CLI with your config file, run:

```bash
gator --config config.yaml
```

## Available Commands
Gator provides several commands to help you manage feeds. Here are a few common ones:

addfeed:
Adds a new feed and automatically follows it.
```bash
gator addfeed <feed-arguments>
```

follow:
Follow an existing feed by its URL.
```bash

gator follow <feed_url>
```

following:
Lists all feeds you are currently following.
```bash
gator following
```
unfollow:
Unfollow a feed by its URL.

```bash
gator unfollow <feed_url>
```
## Contributing


For additional command details, you can run:

License MIT
