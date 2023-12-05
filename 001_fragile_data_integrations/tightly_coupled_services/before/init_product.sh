#!/bin/bash
echo Wait for servers to be up
sleep 10

HOSTPARAMS="--host product_db --insecure"
SQL="/cockroach/cockroach.sh sql $HOSTPARAMS"

$SQL -e "CREATE DATABASE product;"
$SQL -d product -e "CREATE TABLE product.products(id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name STRING NOT NULL);"
$SQL -e "CREATE USER product;"
$SQL -e "GRANT ALL ON product.* to product;"

$SQL -e "INSERT INTO product.products (id, name) VALUES ('ac9384f7-12f7-4431-8a78-c9ccc6d321af', 'a')"
$SQL -e "INSERT INTO product.products (id, name) VALUES ('bf6569f0-08fc-4a01-a3a6-2d353cdda01d', 'b')"
$SQL -e "INSERT INTO product.products (id, name) VALUES ('c9803ecd-04f2-44e4-87ff-e3e5725f93bd', 'c')"
$SQL -e "INSERT INTO product.products (id, name) VALUES ('d394b123-5673-425c-9240-e9fc697bc7fa', 'd')"
$SQL -e "INSERT INTO product.products (id, name) VALUES ('e41a7113-ac42-4ff1-aa0c-c631c1310396', 'e')"