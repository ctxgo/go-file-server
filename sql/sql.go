package sql

import _ "embed"

//go:embed init.sql
var EmbeddedSQLData []byte
