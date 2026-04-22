package domain

import "errors"

var (
	// Organization errors
	ErrBranchNotFound      = errors.New("BRANCH_NOT_FOUND")
	ErrDeptNotFound        = errors.New("DEPARTMENT_NOT_FOUND")
	ErrBranchDeptNotFound  = errors.New("BRANCH_DEPT_NOT_FOUND")
	ErrDuplicateBranchDept = errors.New("DUPLICATE_BRANCH_DEPT")
	ErrInsufficientPermission = errors.New("INSUFFICIENT_PERMISSION")

	// Salary history errors
	ErrSalaryHistoryNotFound = errors.New("SALARY_HISTORY_NOT_FOUND")

	// Input validation error
	ErrValidation = errors.New("VALIDATION_ERROR")
)
