#!/usr/bin/env bash

datasette analytics.db --plugins-dir=../ds/plugins --metadata=../ds/metadata.json --setting base_url $BASE_URL --host 0.0.0.0 --setting sql_time_limit_ms 5000 &
DS_PID=$!
../ingest
kill $DS_PID
