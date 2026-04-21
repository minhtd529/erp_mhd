// Package repository provides the PostgreSQL implementation of the HRM domain
// repository interfaces.
package repository

type scanner interface {
	Scan(dest ...any) error
}
