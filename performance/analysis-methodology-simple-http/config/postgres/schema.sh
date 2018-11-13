#!/usr/bin/env sh

psql -c "create database people"

psql -c "create schema people"

psql ipsets -c "CREATE TABLE people (
    last_updated_time TIMESTAMP WITH TIME ZONE default current_timestamp,
    address text NOT NULL,
    full_name text NOT NULL,
    age integer NOT NULL
    PRIMARY KEY (full_name, address)
)"
