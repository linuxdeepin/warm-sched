package main

type ApplyStatus int

const (
	ApplyStatusUnknown ApplyStatus = iota
	ApplyStatusWaiting
	ApplyStatusApplying
	ApplyStatusSuccessful
	ApplyStatusFailed
)
