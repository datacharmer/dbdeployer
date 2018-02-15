# dbdeployer features 

This table compares features from [MySQL-Sandbox](https://github.com/datacharmer/mysql-sandbox) and [dbdeployer](https://github.com/datacharmer/dbdeployer).

Feature                     | MySQL-Sandbox   | dbdeployer  | dbdeployer planned
--------------------------- | :-------------- | :---------- | :-----------------
Single sandbox deployment   | yes             | yes         |
unpack command              | sort of [^1]    | yes         |
multiple sandboxes          | yes             | yes         |
master-slave replication    | yes             | yes         |
"force" flag                | yes             | yes         |
pre-post grants SQL action  | yes             | yes         |
initialization options      | yes             | yes         |
my.cnf options              | yes             | yes         |   
custom my.cnf               | yes             | yes         |
friendly UUID generation    | yes             | yes         |
global commands             | yes             | yes         |
test replication flow       | yes             | yes         |
delete command              | yes [^2]        | yes         |
group replication  SP       | no              | yes         |
group replication  MP       | no              | yes         |
prevent port collision      | no              | yes  [^3]   |
visible initialization      | no              | yes  [^4]   |
visible script templates    | no              | yes  [^5]   |
replaceable templates       | no              | yes  [^6]   |
list of source binaries     | no              | yes  [^7]   |
list of installed sandboxes | no              | yes  [^8]   |
test script per sandbox     | no              | yes  [^9]   |
integrated usage help       | no              | yes  [^10]  |
custom abbreviations        | no              | yes  [^11]  |
version flag                | no              | yes  [^12]  |
fan-in                      | no              | no          | yes [^13]
all-masters                 | no              | no          | yes [^14]
finding free ports          | yes             | no          | yes
pre-post grants shell action| yes             | no          | maybe
getting remote tarballs     | yes             | no          | yes
circular replication        | yes             | no          | no [^15]
master-master  (circular)   | yes             | no          | no

[^1]: It's achieved using --export_binaries and then abandoning the operation.

[^2]: Uses the sbtool command

[^3]: dbdeployer sandboxes store their ports in a description JSON file, which allows the tool to get a list of used ports and act before a conflict happens.

[^4]: The initialization happens with a script that is generated and stored in the sandbox itself. Users can inspect the *init_db* script and see what was executed.

[^5]: All sandbox scripts are generated using templates, which can be examined and eventually changed.

[^6]: See also note 5. Using the flag --use-template you can replace an existing template on-the-fly. 

[^7]: This is little more than using an O.S. file listing, with the added awareness of the source directory.

[^8]: Using the description files, this command lists the sandboxes with their topology and used ports.

[^9]: It's a basic test that checks whether the sandbox is running and is using the expected port.

[^10]: The "usage" command will show basic commands for single and multiple sandboxes.

[^11]: the abbreviations file allows user to define custom shortcuts for frequently used commands.

[^12]: Strangely enough, this simple feature was never implemented for MySQL-Sandbox, while it was one of the first additions to dbdeployer.

[^13]: Will use the multi source technology introduced in MySQL 5.7.

[^14]: Same as n. 13.

[^15]: Circular replication should not be used anymore. There are enough good alternatives (multi-source, group replication) to avoid this old technology.
