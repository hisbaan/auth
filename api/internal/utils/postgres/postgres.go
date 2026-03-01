package postgres

import (
	. "github.com/go-jet/jet/v2/postgres"
)

func ARRAY_AGG(expr Expression) Expression {
	return Func("ARRAY_AGG", expr)
}
