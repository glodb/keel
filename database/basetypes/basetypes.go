package basetypes

import "fmt"

// CollectionName represents a database collection/table name
type CollectionName string

// DBName represents a database name
type DBName string

// DbType represents the type of database being used
type DbType int

const (
	// MYSQL represents MySQL database type
	MYSQL DbType = iota + 1
	// PSQL represents PostgreSQL database type  
	PSQL
	// MONGO represents MongoDB database type
	MONGO
)

// String returns the string representation of the database type
func (dt DbType) String() string {
	switch dt {
	case MYSQL:
		return "mysql"
	case PSQL:
		return "postgresql"
	case MONGO:
		return "mongodb"
	default:
		return fmt.Sprintf("unknown_db_type_%d", int(dt))
	}
}

// IsValid checks if the database type is valid
func (dt DbType) IsValid() bool {
	return dt >= MYSQL && dt <= MONGO
}

// GetSupportedTypes returns all supported database types
func GetSupportedTypes() []DbType {
	return []DbType{MYSQL, PSQL, MONGO}
}

// ParseDbType parses a string to DbType
func ParseDbType(s string) (DbType, error) {
	switch s {
	case "mysql":
		return MYSQL, nil
	case "postgresql", "psql", "postgres":
		return PSQL, nil
	case "mongodb", "mongo":
		return MONGO, nil
	default:
		return 0, fmt.Errorf("unsupported database type: %s", s)
	}
}
