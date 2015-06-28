package annolityx

import (
	"fmt"
	"github.com/euforia/simplelog"
	//"github.com/metrilyx/annolityx/annolityx/annotations"
	"github.com/metrilyx/annolityx/annolityx/config"
	"net/http"
	"os"
)

type ConfigHandle struct {
	Path   string
	cfg    *config.Config
	logger *simplelog.Logger
}

func NewConfigHandle(path string, cfg *config.Config, logger *simplelog.Logger) *ConfigHandle {
	return &ConfigHandle{Path: path, cfg: cfg, logger: GetLogger(logger)}
}

func (h *ConfigHandle) RegisterHandle() {
	http.Handle(h.Path, h)
}

func (h *ConfigHandle) GetHandler(r *http.Request) (interface{}, int) {
	if h.cfg.Http.WebsocketHostname == "" {

		var err error
		if h.cfg.Http.WebsocketHostname, err = os.Hostname(); err != nil {
			return err, 500
		}
	}

	return map[string]interface{}{
		"websocket": map[string]string{
			"url": fmt.Sprintf(`ws://%s:%d%s`, h.cfg.Http.WebsocketHostname, h.cfg.Http.Port,
				h.cfg.Http.WebsocketEndpoint),
		},
		"endpoints": map[string]string{
			"types":      h.cfg.Http.TypesEndpoint,
			"annotation": h.cfg.Http.AnnoEndpoint,
		},
	}, 200
}

func (h *ConfigHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		resp interface{}
		code int
	)
	switch r.Method {
	case "GET":
		resp, code = h.GetHandler(r)
		break
	default:
		resp = map[string]string{
			"error": fmt.Sprintf("Method not supported: %s", r.Method)}
		code = 405
		break
	}
	WriteJsonResponse(w, r, resp, code)
	h.logger.Info.Printf("%s %d %s\n", r.Method, code, r.URL.RequestURI())
}
