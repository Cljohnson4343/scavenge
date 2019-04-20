package pgsql

// TableColumnMap is a type that maps a database table, the first string key,
// to a column, the second string key, value.
type TableColumnMap map[string]map[string]interface{}

// TableColumnMapper is an interface that wraps the GetTableColumnWrapper method.
// This method is used to generate a mapping between data and its associated
// table and column in a database
type TableColumnMapper interface {
	GetTableColumnMapper() TableColumnMap
}
