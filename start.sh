#!/bin/sh

cd /root/web && python3 -m http.server 8080 &

exec /root/tui-server
