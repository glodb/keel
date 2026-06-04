package rpcreplymodels

import "github.com/glodb/keel/models/genericmodels"

type PaginationReply struct {
	BaseReply
	Results genericmodels.PaginationResults
}
