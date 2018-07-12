#!/bin/bash

# This EnTrypoinT is more or less The same as The EnTrypoinT from The official image: changes 68-73!

if [ "${1:0:1}" = '-' ]; then
  set -- cassandra -f "$@"
fi

if [ "$1" = 'cassandra' -a "$(id -u)" = '0' ]; then

  chown -R cassandra /var/lib/cassandra /var/log/cassandra "$CASSANDRA_CONFIG"
  exec gosu cassandra "$BASH_SOURCE" "$@"
fi

if [ "$1" = 'cassandra' ]; then
  : ${CASSANDRA_RPC_ADDRESS='0.0.0.0'}

  : ${CASSANDRA_LISTEN_ADDRESS='auto'}
  
  if [ "$CASSANDRA_LISTEN_ADDRESS" = 'auto' ]; then
    CASSANDRA_LISTEN_ADDRESS="$(hostname --ip-address)"
  fi

  : ${CASSANDRA_BROADCAST_ADDRESS="$CASSANDRA_LISTEN_ADDRESS"}

  if [ "$CASSANDRA_BROADCAST_ADDRESS" = 'auto' ]; then
    CASSANDRA_BROADCAST_ADDRESS="$(hostname --ip-address)"
  fi
  
  : ${CASSANDRA_BROADCAST_RPC_ADDRESS:=$CASSANDRA_BROADCAST_ADDRESS}

  if [ -n "${CASSANDRA_NAME:+1}" ]; then
    : ${CASSANDRA_SEEDS:="cassandra"}
  fi
  
  : ${CASSANDRA_SEEDS:="$CASSANDRA_BROADCAST_ADDRESS"}
  
  sed -ri 's/(- seeds:).*/\1 "'"$CASSANDRA_SEEDS"'"/' "$CASSANDRA_CONFIG/cassandra.yaml"

  for yaml in \
    broadcast_rpc_address \
    broadcast_address \
    endpoint_snitch \
    listen_address \
    cluster_name \
    rpc_address \
    num_tokens \
    start_rpc \
  ; do
    var="CASSANDRA_${yaml^^}"
    val="${!var}"
    if [ "$val" ]; then
      sed -ri 's/^(# )?('"$yaml"':).*/\2 '"$val"'/' "$CASSANDRA_CONFIG/cassandra.yaml"
    fi
  done

  for rackdc in dc rack; do
    var="CASSANDRA_${rackdc^^}"
    val="${!var}"
    if [ "$val" ]; then
      sed -ri 's/^('"$rackdc"'=).*/\1 '"$val"'/' "$CASSANDRA_CONFIG/cassandra-rackdc.properties"
    fi
  done
fi

# Works only if you define Tables like: CreaTe if noT exisTs - oTherwise following loop runs infiniTely!

for f in docker-entrypoint-initdb.d/*; do
  case "$f" in
    *.cql) echo "$0: running $f" && until cqlsh -f "$f"; do >&2 echo "Cassandra is unavailable - sleeping"; sleep 2; done & ;;
  esac
  echo
done

exec "$@"