DBdeployer
==========

DBdeployer is a tool that deploys MySQL database servers easily.

# Overview

## Goals:
* Allow easy deployment of MySQL database instances, either for testing or for production
* Use a simple and intuitive syntax

## Main features:
* Deploy a single or composite instance of MySQL for either testing or production
* Easily administer the instance with auto generated tools

## Common features:
* Easy automated installation (through Ruby Gems)
* Easy manual installation (should not require gem or other infrastructure)
* Easy remote installation (as part of remote deployment)

## Installation sources:
* a tarball (a .tar.gz file)
* an expanded tarball from $HOME/opt/mysql/x.xx.xx
* installed server (e.g. what is already in /usr/local/mysql/{bin,lib} /usr/{bin,lib}
* source path (after the server has been compiled and built)
* .rpm or .deb files

## Testing instances (same as MySQL Sandbox)
* install a single server instance 
* install multiple server instances in isolation
* install several server instances in replication
* install complex replication topologies (circular, multi-source,etc, when supported)
* install remote single instances
* install remote single instances with replication
* install a single (local or remote) instance as slave of another one

## Production instances:
* install a single instance in recommended location (/usr/local/mysql, /var/lib/mysql)
* install a single instance in custom location 
* install remote single instance
* install remote single instances with replication

## Common custom scripts for each deployment:
* start [options] 
* restart  [options]
* stop
* status 
* clear 
* kill 
* use 
* my
* msb
* test-instance
* test-topology

## Common configuration features:
* pass some options on the command line
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
* test basic instance functionality
* test replication
* make a single instance ready for replication (as master)
* make an instance a slave of an existing instance
* promote a slave to master
* move, copy, or delete an instance

# Implementation

## CLI Interface:
* It will be a command suite (similar to git or svn, with main commands and contextual options)
* The help is a drill-down (general commands first, and help on single commands)

## Implementation requirements:
* All features should be accessible from a library
* All features must be implemented as abstract ones
* Implementation for specific operating systems or environment will be created as subclasses
* every task among the ones needed to complete the deployment can be skipped on demand
* There will be trigger commands that can run before or after each task 
* The application will help defining triggers by listing all the tasks when called in dry-run mode
* All features should be implemented as actions repeatable with portable shell scripts.
* A CLI interface will expose all API in such a way that the library can be called from any programming language

## Special features (require some more thought):
* Easy integration with Percona Toolkit, OpenArk, CommonSchema, xtrabackup, ps_helper
* Easy integration with major testing frameworks (e.g. sysbench)
* Install MySQL Cluster
* Install Tungsten Replicator

## Graphical interface:
A simple web-based GUI will allow the choice of

* Source
* Destination
* Configuration issues
* Replication Topology
* Triggers to run before or after stages

