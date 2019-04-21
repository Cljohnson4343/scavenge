package pgsql

// ColumnMap is a type that maps database column names to their values
type ColumnMap map[string]interface{}

// TableColumnMap is a type that maps a database table, the first string key,
// to a column, the second string key, value.
type TableColumnMap map[string]ColumnMap

// TableColumnMapper is an interface that wraps the GetTableColumnMap method.
// This method is used to generate a mapping between data and the associated
// table, column, and value in a database.
type TableColumnMapper interface {
	GetTableColumnMap() TableColumnMap
}
