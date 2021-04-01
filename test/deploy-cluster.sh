#!/bin/bash

scriptdir=$(dirname "$0")
topdir=${scriptdir}/..
source ${scriptdir}/env.sh

action=${1:-up}

if [[ -z "${TEST_CLUSTER_KIND}" ]]; then
  echo "Test Cluster Kind Not Set in Environment Variables. Chaning to env.sh value"
  export TEST_CLUSTER_KIND="${CLUSTER_PROVIDER}"
fi

if [[ -z "${CLUSTER_NODES}" ]]; then
  echo "Test Cluster Nodes Not Set in Environment Variables. Chaning to env.sh value"
  export CLUSTER_NODES="${NUM_NODES}"
fi

if [[ -z "${CLUSTER_WORKERS}" ]]; then
  echo "Test Cluster Workers Not Set in Environment Variables. Chaning to env.sh value"
  export CLUSTER_WORKERS="${NUM_WORKERS}"
fi

# source the script file containing the create/delete cluster implementation
srcfile="${scriptdir}/deploy-${TEST_CLUSTER_KIND}-cluster.sh"

if [ ! -f "$srcfile" ]; then
    echo "Unknown cluster provider '${TEST_CLUSTER_KIND}'"
    exit 1
fi
source $srcfile

case ${action} in
    up)
        createCluster
        ;;
    down)
        deleteCluster
        ;;
esac
