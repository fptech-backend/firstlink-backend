#!/bin/bash

set -e

# Directory where backups are stored
BACKUP_DIR=/docker-entrypoint-initdb.d

# Database credentials
DB_USER=${POSTGRES_USER}
DB_PASSWORD=${POSTGRES_PASSWORD}
DB_NAME=${POSTGRES_DB}

# Check if RESTORE_DB is set to true
if [ "$RESTORE_DB" = "true" ]; then
  echo "Restore is enabled."

  # Find the latest backup file
  LATEST_BACKUP=$(ls -1t $BACKUP_DIR | head -n 1)

  if [ -z "$LATEST_BACKUP" ]; then
    echo "No backup file found to restore."
    exit 1
  fi

  BACKUP_FILE="$BACKUP_DIR/$LATEST_BACKUP"
  echo "Restoring from backup file: $BACKUP_FILE"

  # Restore the database
  PGPASSWORD=$DB_PASSWORD pg_restore -h localhost -U $DB_USER -d $DB_NAME -F c $BACKUP_FILE
  echo "Database restored from $BACKUP_FILE"
else
  echo "Restore is not enabled."
fi
