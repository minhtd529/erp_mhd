package domain

import "errors"

var (
	ErrBranchNotFound     = errors.New("BRANCH_NOT_FOUND")
	ErrDepartmentNotFound = errors.New("DEPARTMENT_NOT_FOUND")
	ErrDuplicateCode      = errors.New("DUPLICATE_CODE")
)
