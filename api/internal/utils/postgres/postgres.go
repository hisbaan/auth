package postgres

import (
	. "github.com/go-jet/jet/v2/postgres"
)

func ARRAY_AGG_TEXT_NO_NULLS(expr Expression) Array[StringExpression] {
	return CAST(Func("ARRAY_REMOVE", Func("ARRAY_AGG", expr), NULL)).AS_TEXT_ARRAY()
}
