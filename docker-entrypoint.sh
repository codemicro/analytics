#!/usr/bin/env bash

datasette analytics.db --plugins-dir=../ds/plugins --metadata=../ds/metadata.json --setting base_url $BASE_URL --host 0.0.0.0 &
DS_PID=$!
../ingest
kill $DS_PID
