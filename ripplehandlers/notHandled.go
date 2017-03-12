package ripplehandlers

import (
	"net/http"

	"github.com/whoisjeremylam/enu/consts"
	"github.com/whoisjeremylam/enu/enulib"
	"github.com/whoisjeremylam/enu/handlers"
	"github.com/whoisjeremylam/enu/internal/golang.org/x/net/context"
	"github.com/whoisjeremylam/enu/log"
)

func Unhandled(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	log.FluentfContext(consts.LOGINFO, c, "Unhandled function called: %s", c.Value(consts.RequestTypeKey).(string))
	handlers.ReturnNotFoundWithCustomError(c, w, consts.GenericErrors.FunctionNotAvailable.Code, consts.GenericErrors.FunctionNotAvailable.Description)

	return nil
}
