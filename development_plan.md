DBdeployer
==========

DBdeployer is a tool that deploys MySQL database servers easily.

# Overview

## Goals:
* Replace [MySQL-Sandbox](https://github.com/datacharmer/mysql-sandbox)
* Allow easy deployment of MySQL database instances, both for testing and production
* Use a simple and intuitive syntax

## Main features:
* Deploy a single or composite instance of MySQL for testing 
* Easily administer the instance with auto generated tools

## Common features:
* Easy installation (Simple deployment of a binary file)
* Easy manual installation (should not require any infrastructure)

## Installation sources:
* an expanded tarball from $HOME/opt/mysql/x.xx.xx (default)
* a tarball (a .tar.gz file: requires additional command)

In the near future:

* installed server (e.g. what is already in /usr/local/mysql/{bin,lib} /usr/{bin,lib}
* source path (after the server has been compiled and built)
* .rpm or .deb files

## Production instances (future dev):
* install a single instance in recommended location (/usr/local/mysql, /var/lib/mysql)
* install a single instance in custom location
* store an instance for further usage (will save and eventually compress binaries, data, and configuration)
* resume an instance from storage (we call it "go live")
* install remote single instance
* install remote single instances with replication

## Common custom scripts for each deployment:
* start [options] 
* restart  [options]
* stop
* status 
* clear 
* send\_kill 
* use 

## Common configuration features:
* pass some options on the command line

Future:

* allow different options for master and slave
* allow using a my.cnf as template
* produce automatic or custom server-ids
* produce custom server UUIDs
* install plug-ins on demand
* keep an easily accessible configuration file to inspect or modify the instance
* produce connection hooks to connect in various languages

## Common administrative tasks:
* Update configuration files with new/different options
* start or restart an instance with different options

Future:
* test basic instance functionality
* test replication
* make a single instance ready for replication (as master)
* make an instance a slave of an existing instance
* promote a slave to master
* move, copy, or delete an instance

# Implementation

## CLI Interface:
* It will be a command suite (similar to git or docker, with main commands and contextual options)
* The help is a drill-down (general commands first, and help on single commands)

## Implementation requirements:
* All features will be accessible from a single binary file
* All features must be implemented as easily extendable
* Implementation for specific operating systems or environment will be possible

Future:

* Every task among the ones needed to complete the deployment can be skipped on demand
* There will be trigger commands that can run before or after each task 
* The application will help defining triggers by listing all the tasks when called in dry-run mode

## Special features (require some more thought):
* Easy integration with popular tools, such as Percona Toolkit, OpenArk, xtrabackup, ProxySQL.
* Easy integration with major testing frameworks (e.g. sysbench)
* Install Galera or MySQL Group Replication
* (low priority) Install MySQL Cluster

## Graphical interface (low priority):
A simple web-based GUI will allow the choice of

* Source
* Destination
* Configuration issues
* Replication Topology
* Triggers to run before or after stages

