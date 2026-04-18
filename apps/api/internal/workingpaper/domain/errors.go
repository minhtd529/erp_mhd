package domain

import "errors"

var (
	ErrWorkingPaperNotFound    = errors.New("WORKING_PAPER_NOT_FOUND")
	ErrWorkingPaperLocked      = errors.New("WORKING_PAPER_LOCKED")
	ErrWorkingPaperNotEditable = errors.New("WORKING_PAPER_NOT_EDITABLE")
	ErrInvalidStateTransition = errors.New("INVALID_STATE_TRANSITION")
	ErrReviewChainIncomplete  = errors.New("REVIEW_CHAIN_INCOMPLETE")
	ErrInvalidReviewSequence  = errors.New("INVALID_REVIEW_SEQUENCE")
	ErrCommentsNotResolved    = errors.New("COMMENTS_NOT_RESOLVED")
	ErrTemplateNotFound       = errors.New("TEMPLATE_NOT_FOUND")
	ErrReviewNotFound         = errors.New("REVIEW_NOT_FOUND")
	ErrCommentNotFound        = errors.New("COMMENT_NOT_FOUND")
	ErrFolderNotFound         = errors.New("FOLDER_NOT_FOUND")
)
