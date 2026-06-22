#!/usr/bin/env bash
# Seed the Labzyn research data model (Experiment, Run, Result, Dataset,
# Artifact, Environment, Allocation) into a workspace via the REST metadata API.
# Idempotent-ish: re-running returns 409s for existing objects/fields.
#
# Usage: API=http://localhost:4000/api/v1 EMAIL=adama@labzyn.dev PASSWORD=labzyn123 ./seed-research-model.sh
set -euo pipefail

API="${API:-http://localhost:4000/api/v1}"
EMAIL="${EMAIL:-adama@labzyn.dev}"
PASSWORD="${PASSWORD:-labzyn123}"

TOK=$(curl -s -X POST "$API/auth/login" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['token'])")

H() { curl -s -H "Authorization: Bearer $TOK" -H 'Content-Type: application/json' "$@"; }
obj() { H -X POST "$API/metadata/objects" -d "$1" -o /dev/null -w "  obj $2 -> %{http_code}\n"; }
fld() { H -X POST "$API/metadata/objects/$1/fields" -d "$2" -o /dev/null -w "    + $3 -> %{http_code}\n"; }

echo "Experiment"
obj '{"nameSingular":"experiment","namePlural":"experiments","labelSingular":"Experiment","labelPlural":"Experiments","icon":"IconFlask"}' experiment
fld experiment '{"name":"name","type":"TEXT","label":"Name","isNullable":false}' name
fld experiment '{"name":"hypothesis","type":"TEXT","label":"Hypothesis"}' hypothesis
fld experiment '{"name":"status","type":"SELECT","label":"Status","options":[{"value":"PLANNED","label":"Planned"},{"value":"RUNNING","label":"Running"},{"value":"DONE","label":"Done"}]}' status

echo "Run"
obj '{"nameSingular":"run","namePlural":"runs","labelSingular":"Run","labelPlural":"Runs","icon":"IconPlayerPlay"}' run
fld run '{"name":"name","type":"TEXT","label":"Name","isNullable":false}' name
fld run '{"name":"script","type":"TEXT","label":"Script"}' script
fld run '{"name":"status","type":"SELECT","label":"Status","options":[{"value":"PENDING","label":"Pending"},{"value":"RUNNING","label":"Running"},{"value":"COMPLETED","label":"Completed"},{"value":"FAILED","label":"Failed"}]}' status
fld run '{"name":"partition","type":"TEXT","label":"Partition"}' partition
fld run '{"name":"nodeHours","type":"NUMERIC","label":"Node Hours"}' nodeHours
fld run '{"name":"hpcJobId","type":"TEXT","label":"HPC Job Id"}' hpcJobId
fld run '{"name":"exitCode","type":"NUMBER","label":"Exit Code"}' exitCode

echo "Result"
obj '{"nameSingular":"result","namePlural":"results","labelSingular":"Result","labelPlural":"Results","icon":"IconChartHistogram"}' result
fld result '{"name":"name","type":"TEXT","label":"Name","isNullable":false}' name
fld result '{"name":"fileName","type":"TEXT","label":"File Name"}' fileName
fld result '{"name":"fileSizeBytes","type":"NUMBER","label":"File Size (bytes)"}' fileSizeBytes
fld result '{"name":"stdout","type":"TEXT","label":"Stdout"}' stdout

echo "Dataset"
obj '{"nameSingular":"dataset","namePlural":"datasets","labelSingular":"Dataset","labelPlural":"Datasets","icon":"IconDatabase"}' dataset
fld dataset '{"name":"name","type":"TEXT","label":"Name","isNullable":false}' name
fld dataset '{"name":"storageUri","type":"TEXT","label":"Storage URI"}' storageUri
fld dataset '{"name":"format","type":"TEXT","label":"Format"}' format
fld dataset '{"name":"sizeBytes","type":"NUMBER","label":"Size (bytes)"}' sizeBytes
fld dataset '{"name":"checksum","type":"TEXT","label":"Checksum"}' checksum
fld dataset '{"name":"version","type":"TEXT","label":"Version"}' version

echo "Artifact"
obj '{"nameSingular":"artifact","namePlural":"artifacts","labelSingular":"Artifact","labelPlural":"Artifacts","icon":"IconFileCode"}' artifact
fld artifact '{"name":"name","type":"TEXT","label":"Name","isNullable":false}' name
fld artifact '{"name":"kind","type":"SELECT","label":"Kind","options":[{"value":"SCRIPT","label":"Script"},{"value":"NOTEBOOK","label":"Notebook"},{"value":"CONTAINER","label":"Container"},{"value":"MODEL","label":"Model"}]}' kind
fld artifact '{"name":"uri","type":"TEXT","label":"URI"}' uri
fld artifact '{"name":"language","type":"TEXT","label":"Language"}' language
fld artifact '{"name":"gitCommit","type":"TEXT","label":"Git Commit"}' gitCommit

echo "Environment"
obj '{"nameSingular":"environment","namePlural":"environments","labelSingular":"Environment","labelPlural":"Environments","icon":"IconBox"}' environment
fld environment '{"name":"name","type":"TEXT","label":"Name","isNullable":false}' name
fld environment '{"name":"kind","type":"SELECT","label":"Kind","options":[{"value":"MODULES","label":"Modules"},{"value":"CONTAINER","label":"Container"}]}' kind
fld environment '{"name":"spec","type":"TEXT","label":"Spec"}' spec
fld environment '{"name":"imageDigest","type":"TEXT","label":"Image Digest"}' imageDigest

echo "Allocation"
obj '{"nameSingular":"allocation","namePlural":"allocations","labelSingular":"Allocation","labelPlural":"Allocations","icon":"IconWallet"}' allocation
fld allocation '{"name":"name","type":"TEXT","label":"Name","isNullable":false}' name
fld allocation '{"name":"status","type":"SELECT","label":"Status","options":[{"value":"ACTIVE","label":"Active"},{"value":"EXHAUSTED","label":"Exhausted"},{"value":"EXPIRED","label":"Expired"}]}' status
fld allocation '{"name":"partition","type":"TEXT","label":"Partition"}' partition
fld allocation '{"name":"nodeHoursGranted","type":"NUMERIC","label":"Node Hours Granted"}' nodeHoursGranted
fld allocation '{"name":"nodeHoursUsed","type":"NUMERIC","label":"Node Hours Used"}' nodeHoursUsed

echo "Done. Restart the server so the dynamic records schema picks up the new objects."
