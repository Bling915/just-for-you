package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/nanfang/just-for-you/server/internal/couple"
)

type RelationshipRegistrar struct{ service *couple.Service }

func NewRelationshipRegistrar(service *couple.Service) RelationshipRegistrar {
	return RelationshipRegistrar{service: service}
}
func (RelationshipRegistrar) RegisterPublic(*gin.Engine) {}
func (r RelationshipRegistrar) RegisterProtected(group *gin.RouterGroup) {
	handler := coupleHandler{s: r.service}
	group.POST("/couples/invitations", handler.create)
	group.GET("/couples/invitations/:code", handler.preview)
	group.POST("/couples/bind", handler.accept)
	group.GET("/couples/me", handler.get)
	group.DELETE("/couples/me", handler.unbind)
}
