# DBdeployer

DBdeployer is a tool that deploys MySQL database servers easily.

# Overview

## Goals:
* Replace [MySQL-Sandbox](https://github.com/datacharmer/mysql-sandbox)
* Allow easy deployment of MySQL database instances for testing.
* Use a simple and intuitive syntax

## Main features:
* Deploy a single or composite instance of MySQL 
* Easily administer the instance with auto generated tools
* Easy installation (Simple deployment of a binary file without dependencies)

## Installation sources:
* an expanded tarball from $HOME/opt/mysql/x.xx.xx (default)
* a tarball (a .tar.gz file: requires additional command)

## Common custom scripts for each deployment:
* start [options] 
* restart  [options]
* stop
* status 
* clear 
* send\_kill 
* use 
* test\_sb

## Common configuration features:
* Pass some options on the command line
* Allow using a my.cnf as template
* Produce automatic or custom server-ids
* Produce custom server UUIDs
* keep an easily accessible configuration file to inspect the instance

## Common administrative tasks:
* Update configuration files with new/different options
* start or restart an instance with different options
* test basic instance functionality
* test replication
* make a single instance ready for replication (as master)

Future:
* make an instance a slave of an existing instance
* move, copy, or delete an instance

# Implementation

## CLI Interface:
* It will be a command suite (similar to git or docker, with main commands and contextual options)
* The help is a drill-down (general commands first, and help on single commands)

## Implementation requirements:
* Unix operating system (Linux or MacOS tested so far)
* All features are accessible from a single binary file
* All features are implemented as easily extendable
* The only hard requirement is that the OS is ready to run a MySQL server

See features.md for a detailed list of implemented features.
