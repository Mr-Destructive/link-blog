package embedsql

import (
	_ "embed"
)

//go:embed schema.sql
var DDL string

//go:embed list.html
var ListHTML string

//go:embed link.html
var LinkHTML string

//go:embed edit.html
var EditHTML string
