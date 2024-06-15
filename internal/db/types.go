package db

import "github.com/jackc/pgx/v5/pgtype"

// Text returns a properly initialized pgtype.Text instance based on the passed value. Please note empty strings
// are considered as NULL. Don't use this method if you want to set a column empty instead of NULL.
func Text(value string) pgtype.Text {
	return pgtype.Text{
		String: value,
		Valid:  value != "",
	}
}

func Int4(value int32) pgtype.Int4 {
	return pgtype.Int4{
		Int32: value,
		Valid: true,
	}
}
