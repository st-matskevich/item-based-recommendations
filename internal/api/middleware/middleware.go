package middleware

import (
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/firebase"
)

func AuthMiddleware(inner utils.BaseHandler) utils.BaseHandler {
	return utils.BaseHandler(func(r *http.Request) utils.HandlerResponse {
		uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
		if err != nil {
			return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
		}
		ctx := utils.SetUserID(r.Context(), uid)

		return inner(r.WithContext(ctx))
	})
}
