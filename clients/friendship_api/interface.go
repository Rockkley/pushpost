package friendship_api

import (
	"context"
	"github.com/google/uuid"
)

type Client interface {
	GetRelationship(ctx context.Context, viewerID, targetID uuid.UUID) (RelationshipResponse, error)
}
