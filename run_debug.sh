#!/bin/sh
exec gin --port 0 --appPort ${PORT:-3000} --immediate --bin $(basename $(pwd))
