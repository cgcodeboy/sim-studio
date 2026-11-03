package simulation

import (
	"strings"

	model "github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
)

func invalid() error { return &model.Error{Code: model.CodeInvalid} }

func unavailable() error { return &model.Error{Code: model.CodeUnavailable} }

func banned(props map[string]any) bool {
	for _, value := range props {
		text, ok := value.(string)
		if !ok {
			continue
		}
		folded := strings.ToLower(text)
		for _, needle := range []string{"import ", " class ", ".jar", "jdbc:", "python", "rscript", ".xlsx", ".xls"} {
			if strings.Contains(folded, needle) {
				return true
			}
		}
	}
	return false
}

func refuse(p model.Project) error {
	for _, agent := range p.Agents {
		for _, el := range agent.Elements {
			if banned(el.Props) {
				return invalid()
			}
		}
	}
	for _, exp := range p.Experiments {
		if banned(exp.Props) {
			return invalid()
		}
	}
	return nil
}
