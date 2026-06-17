#!/bin/sh
set -u

api_pid=
web_pid=

stop_children() {
  if [ -n "${api_pid}" ]; then
    kill "${api_pid}" 2>/dev/null || true
  fi
  if [ -n "${web_pid}" ]; then
    kill "${web_pid}" 2>/dev/null || true
  fi
  wait 2>/dev/null || true
}

shutdown() {
  stop_children
  exit 143
}

trap shutdown INT TERM

/usr/local/bin/monitor &
api_pid=$!

node /app/.output/server/index.mjs &
web_pid=$!

while :; do
  if ! kill -0 "${api_pid}" 2>/dev/null; then
    wait "${api_pid}"
    status=$?
    kill "${web_pid}" 2>/dev/null || true
    wait "${web_pid}" 2>/dev/null || true
    exit "${status}"
  fi

  if ! kill -0 "${web_pid}" 2>/dev/null; then
    wait "${web_pid}"
    status=$?
    kill "${api_pid}" 2>/dev/null || true
    wait "${api_pid}" 2>/dev/null || true
    exit "${status}"
  fi

  sleep 1
done
