package service

import (
	"github.com/cgcodeboy/sim-studio/internal/pkgs/pedestrian/domain"
)

// PlaceSettings 装上行人库的设置。
func PlaceSettings(w domain.Repository, el domain.Block) (bool, error) {
	if canon(el.Type) != "PedSettings" {
		return false, nil
	}
	Tune(w.Params(), el)
	return true, nil
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
