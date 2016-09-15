#!/bin/sh

# Get a randomized list of available hosts

num_hosts=$1

rocks list host \
    | grep compute \
    | cut -d" " -f1 \
    | sed 's/.$//' \
    | shuf \
    | head -n "$num_hosts" \
    | tr '\n' ' '
echo
