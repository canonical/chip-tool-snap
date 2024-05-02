#!/bin/bash

export TMPDIR=$SNAP_USER_COMMON

exec "$@"