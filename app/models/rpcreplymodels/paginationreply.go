package rpcreplymodels

import "github.com/glodb/keel/app/models/genericmodels"

type PaginationReply struct {
	BaseReply
	Results genericmodels.PaginationResults
}
