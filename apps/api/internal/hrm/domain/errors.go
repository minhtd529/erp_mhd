package domain

import "errors"

var (
	// Organization errors
	ErrBranchNotFound      = errors.New("BRANCH_NOT_FOUND")
	ErrDeptNotFound        = errors.New("DEPARTMENT_NOT_FOUND")
	ErrBranchDeptNotFound  = errors.New("BRANCH_DEPT_NOT_FOUND")
	ErrDuplicateBranchDept = errors.New("DUPLICATE_BRANCH_DEPT")
	ErrInsufficientPermission = errors.New("INSUFFICIENT_PERMISSION")
)
