package service

import "strings"

func canon(name string) string { return strings.ReplaceAll(name, " ", "") }
