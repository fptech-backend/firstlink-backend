package database

import (
	"fmt"

	"gorm.io/gorm"
)

func IsMigrateTableEmpty(db *gorm.DB, table string) bool {
	var count int
	query := fmt.Sprintf("SELECT COUNT(ID) FROM %s", table)
	db.Raw(query).Scan(&count)

	return count == 0
}
