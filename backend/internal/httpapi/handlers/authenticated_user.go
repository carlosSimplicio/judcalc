package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

const AuthenticatedUserKey = "authenticated_user"

func AuthenticatedUser(ctx *gin.Context) (domain.User, bool) {
	value, exists := ctx.Get(AuthenticatedUserKey)
	if !exists {
		return domain.User{}, false
	}
	user, ok := value.(domain.User)
	return user, ok && user.ID > 0
}
