package model

import (
	"github.com/shippinAI/nomen/internal/domain"
)

type LabelPolicySearchKey int32

const (
	LabelPolicySearchKeyUnspecified LabelPolicySearchKey = iota
	LabelPolicySearchKeyAggregateID
	LabelPolicySearchKeyState
	LabelPolicySearchKeyInstanceID
	LabelPolicySearchKeyOwnerRemoved
)

type LabelPolicySearchQuery struct {
	Key    LabelPolicySearchKey
	Method domain.SearchMethod
	Value  interface{}
}
