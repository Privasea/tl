package dbx

import "strings"

func setDefaultDatabase(dsn string) {
	s := strings.Split(dsn, "/")
	if len(s) > 1 {
		s1 := strings.Split(s[1], "?")
		if len(s1) > 0 {
			defaultDatabase = s1[0]
		}
	}
}
