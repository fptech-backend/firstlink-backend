#!/bin/bash

set -e

# Directory to store backups
BACKUP_DIR=/docker-entrypoint-initdb.d

# Database credentials
DB_USER=${POSTGRES_USER}
DB_PASSWORD=${POSTGRES_PASSWORD}
DB_NAME=${POSTGRES_DB}

# Ensure the backup directory exists
mkdir -p $BACKUP_DIR

# Print environment variables for debugging
echo "POSTGRES_USER: $DB_USER"
echo "POSTGRES_PASSWORD: $DB_PASSWORD"
echo "POSTGRES_DB: $DB_NAME"

# Function to perform the backup
perform_backup() {
  # Generate the backup filename with a timestamp
  BACKUP_FILE="$BACKUP_DIR/backup_$(date +'%Y%m%d%H%M%S').sql"

  # Perform the backup
  echo "Performing backup..."
  PGPASSWORD=$DB_PASSWORD pg_dump -h localhost -U $DB_USER -d $DB_NAME -F c > $BACKUP_FILE

  # Check if the backup was successful
  if [ $? -eq 0 ]; then
    echo "Backup completed: $BACKUP_FILE"
  else
    echo "Backup failed"
    rm -f $BACKUP_FILE
  fi

  # Remove backups older than the 15 latest ones
  cd $BACKUP_DIR
  ls -1t | tail -n +16 | xargs rm -f --

  echo "Old backups cleaned up."
}

# Wait for PostgreSQL to start
until pg_isready -h localhost -U "$DB_USER"; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

# Perform the backup
perform_backup
