package enums

type PeriodStatus string

const (
	PeriodStatusDraft    PeriodStatus = "draft"
	PeriodStatusOpen     PeriodStatus = "open"
	PeriodStatusClosed   PeriodStatus = "closed"
	PeriodStatusArchived PeriodStatus = "archived"
)

type TemplateStatus string

const (
	TemplateStatusActive   TemplateStatus = "active"
	TemplateStatusInactive TemplateStatus = "inactive"
)

type QuestionType string

const (
	QuestionTypeRating   QuestionType = "rating"
	QuestionTypeText     QuestionType = "text"
	QuestionTypeNumber   QuestionType = "number"
	QuestionTypeBoolean  QuestionType = "boolean"
	QuestionTypeMultiple QuestionType = "multiple_choice"
)

type QuestionStatus string

const (
	QuestionStatusActive   QuestionStatus = "active"
	QuestionStatusInactive QuestionStatus = "inactive"
)

type AssignmentStatus string

const (
	AssignmentStatusPending   AssignmentStatus = "pending"
	AssignmentStatusStarted   AssignmentStatus = "started"
	AssignmentStatusSubmitted AssignmentStatus = "submitted"
	AssignmentStatusCompleted AssignmentStatus = "completed"
	AssignmentStatusOverdue   AssignmentStatus = "overdue"
)

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)
