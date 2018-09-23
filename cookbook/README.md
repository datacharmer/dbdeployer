# dbdeployer usage recipes

The scripts in this directory show some of the most interesting ways of using dbdeployer.

Each script runs with this syntax:

```bash
./script_name.sh version
```
where `version` is `5.7.23`, or `8.0.12`, or any other recent version of MySQL. For this to work, you ned to have unpacked the tarball binaries for the corresponding version. For example:

```bash
$ ./single.sh 8.0.13
Directory $HOME/opt/mysql/8.0.13 not found
To install the binaries, use:
    dbdeployer unpack mysql-8.0.13-YOUR-OPERATING-SYSTEM.tar.gz
```
Then, you need to download `mysql-8.0.13-linux-glibc2.12-x86_64.tar.gz` (or `mysql-8.0.13-macos10.13-x86_64.tar.gz` if you are on a Mac) and run the command.

```bash
dbdeployer unpack /downloads/mysql-8.0.13-linux-glibc2.12-x86_64.tar.gz
```

After that, all operations on version name `8.0.13` will work.

You can run the same command several times, provided that you use a different version at every call.

```bash
./single 5.7.24
./single 8.0.13
```

## deployment 

`single.sh` deploys a single sandbox, then redeploys it with `--force` and finally shows how dbdeployer computes a new port for the same version.

The following scripts will simply deploy the corresponding replication topology.

* `replication-master-slave.sh` (This is needed to run `operations-replication.sh`)
* `replication-group-multi-primary.sh`
* `replication-group-single-primary.sh`
* `replication-all-masters.sh`
* `replication-fan-in.sh`

## operations

These scripts run operations on sandboxes already deployed.

* `show-sandboxes.sh` shows the list of installed sandboxes.

* `operations-single.sh` runs various read-only and writing operations on a single sandbox.
* `operations-replication.sh` shows how to deal with master/slave deployment, by running a command in a single node or in all nodes at once, or all slaves only.
* `operations-restart.sh` shows how to alter temporarily sandboxes with a restart that includes new options, and how to alter them permanently.

* `delete-all.sh` will delete all sandboxes (includes the ones that you may have deployed on your own, outside this cookbook). WARNING: the effects of this command are not recoverable! The deletion command asks for confirmation. After that, there is no going back.

* `upgrade.sh` is the most entertaining of them all. It runs an upgrade from 5.5 to 5.6, then the upgraded database is upgraded to 5.7, and finally to 8.0. Along the way, each database writes to the same table, so that you can see the effects of the upgrade.
Here's an example.
```
+----+-----------+------------+----------+---------------------+
| id | server_id | vers       | urole    | ts                  |
+----+-----------+------------+----------+---------------------+
|  1 |      5553 | 5.5.53-log | original | 2018-09-23 11:25:34 |
|  2 |      5641 | 5.6.41-log | upgraded | 2018-09-23 11:25:44 |
|  3 |      5641 | 5.6.41-log | original | 2018-09-23 11:25:49 |
|  4 |      5723 | 5.7.23-log | upgraded | 2018-09-23 11:26:00 |
|  5 |      5723 | 5.7.23-log | original | 2018-09-23 11:26:05 |
|  6 |      8012 | 8.0.12     | upgraded | 2018-09-23 11:26:18 |
+----+-----------+------------+----------+---------------------+
```
If you don't have the same versions expanded in your system, you should edit the script and change the variables `ver_55`, `ver_56`, `ver_57`, and `ver_80`.
