package embedsql

import (
	_ "embed"
)

//go:embed schema.sql
var DDL string
