package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
)

func PlaceSettings(w domain.Repository, el domain.Block) (bool, error) {
	switch canon(el.Type) {
	case "MHLSettings":
		Tune(w.Params(), el)
		return true, nil
	default:
		return false, nil
	}
}

// Tune 把设置元件上的数值抄进参数表。
func Tune(params map[string]float64, el domain.Block) {
	if params == nil {
		return
	}
	for key, value := range el.Props {
		switch value.(type) {
		case float64, int, int32, int64, string:
			params[el.ID+"/"+key] = Number(value, 0)
		}
	}
}
