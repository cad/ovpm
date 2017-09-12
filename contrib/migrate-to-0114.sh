#!/bin/bash

#set -x
set -e

DIR="/var/db/ovpm"
SQLITEBIN=`which sqlite3`

TABLE_NAME_PAIRS="db_networks,db_network_models db_revokeds,db_revoked_models db_servers,db_server_models db_users,db_user_models"

# backup
echo "backing up $DIR/db.sqlite3 to /tmp/bak-db.sqlite3"
cp -f $DIR/db.sqlite3 /tmp/bak-db.sqlite3

for i in $TABLE_NAME_PAIRS; do
    IFS=","
    set $i
    echo "migrating table '$1' to '$2'"
    $SQLITEBIN $DIR/db.sqlite3 "ALTER TABLE $2 RENAME TO old_$2;"  # move the tables
    $SQLITEBIN $DIR/db.sqlite3 "ALTER TABLE $1 RENAME TO $2;"  # migrate
    unset IFS
done
echo "done!"
