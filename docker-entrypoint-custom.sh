#!/bin/bash
set -e

# Start the cron service
service cron start

# Check if RESTORE_DB is set to true and run the restore script if it is
if [ "$RESTORE_DB" = "true" ]; then
  /usr/local/bin/restore.sh
fi

# Execute the original entrypoint script with the given arguments
exec docker-entrypoint.sh "$@"
