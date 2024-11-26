#!/bin/bash
set -euo pipefail
sqlite3 paste.db '.backup backup.db'
zstd -o backup-paste-"$(date '+%F')".db.zstd backup.db
rm backup.db
