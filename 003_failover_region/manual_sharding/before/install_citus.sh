#!/bin/bash
apt-get update
apt-get install curl -y

curl https://install.citusdata.com/community/deb.sh | bash

apt-get install postgresql-16-citus-12.1 -y

echo "wal_level = 'logical'" >> /var/lib/postgresql/data/postgresql.conf
echo "shared_preload_libraries = 'citus'" >> /var/lib/postgresql/data/postgresql.conf
echo "host all all 0.0.0.0/0 trust" >> /var/lib/postgresql/data/pg_hba.conf
echo "host all all 0.0.0.0/0 md5" >> /var/lib/postgresql/data/pg_hba.conf

sed -i 's/\(^host all all all scram-sha-256$\)/#\1/' /var/lib/postgresql/data/pg_hba.conf