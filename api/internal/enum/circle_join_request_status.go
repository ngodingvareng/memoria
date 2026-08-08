package enum

type CircleJoinRequestStatus string

const (
	CircleJoinRequestStatusPending   CircleJoinRequestStatus = "pending"
	CircleJoinRequestStatusApproved  CircleJoinRequestStatus = "approved"
	CircleJoinRequestStatusRejected  CircleJoinRequestStatus = "rejected"
	CircleJoinRequestStatusCancelled CircleJoinRequestStatus = "cancelled"
)
