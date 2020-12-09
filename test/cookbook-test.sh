#!/usr/bin/env bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2020 Giuseppe Maxia
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

if [ -n "$GITHUB_ACTIONS" ]
then
    exit 0
fi

if [ -d recipes ]
then
    rm -rf recipes
fi

dbdeployer cookbook make all
if [ "$?" != "0" ] ; then exit 1 ; fi

g0=(admin-single.sh)

g1=(single-deployment.sh
single-custom-users.sh
single-reinstall.sh
tidb-deployment.sh
)

g2=(master-slave-deployment.sh
repl-operations.sh
repl-operations-restart.sh
)

g3=(upgrade.sh)

g4=(ndb-deployment.sh
fan-in-deployment.sh
all-masters-deployment.sh
)

g5=(circular-replication.sh
replication-multi-versions.sh
custom-named-replication.sh
)

g6=(group-multi-primary-deployment.sh
group-single-primary-deployment.sh
)

g7=(replication-between-groups.sh
replication-between-master-slave.sh
replication-between-ndb.sh
replication-between-single.sh
replication-group-master-slave.sh
replication-group-single.sh
replication-master-slave-group.sh
replication-single-group.sh
)

for s in ${g0[*]}
do
    echo "# -- $s"
    ./recipes/$s
    if [ "$?" != "0" ] ; then exit 1 ; fi
done
dbdeployer delete all --concurrent --skip-confirm

for s in ${g1[*]}
do
    echo "# -- $s"
    ./recipes/$s
    if [ "$?" != "0" ] ; then exit 1 ; fi
done
dbdeployer delete all --concurrent --skip-confirm

for s in ${g2[*]}
do
    echo "# -- $s"
    ./recipes/$s
    if [ "$?" != "0" ] ; then exit 1 ; fi
done
dbdeployer delete all --concurrent --skip-confirm

for s in ${g3[*]}
do
    echo "# -- $s"
    ./recipes/$s
    if [ "$?" != "0" ] ; then exit 1 ; fi
done
dbdeployer delete all --concurrent --skip-confirm

for s in ${g4[*]}
do
    echo "# -- $s"
    ./recipes/$s
    if [ "$?" != "0" ] ; then exit 1 ; fi
done
dbdeployer delete all --concurrent --skip-confirm

for s in ${g5[*]}
do
    echo "# -- $s"
    ./recipes/$s
    if [ "$?" != "0" ] ; then exit 1 ; fi
done
dbdeployer delete all --concurrent --skip-confirm

for s in ${g6[*]}
do
    echo "# -- $s"
    ./recipes/$s
    if [ "$?" != "0" ] ; then exit 1 ; fi
done
dbdeployer delete all --concurrent --skip-confirm

for s in ${g7[*]}
do
    echo "# -- $s"
    ./recipes/$s
    if [ "$?" != "0" ] ; then exit 1 ; fi
    dbdeployer delete all --concurrent --skip-confirm
done

if [ -d recipes ]
then
    rm -rf recipes
fi

