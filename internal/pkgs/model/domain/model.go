package domain

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	CodeInvalid     = "invalid"
	CodeNotFound    = "notfound"
	CodeConflict    = "conflict"
	CodeRejected    = "rejected"
	CodeUnavailable = "unavailable"
)

type Error struct {
	Code    string
	Project *Project
}

func (e *Error) Error() string { return e.Code }

type Project struct {
	ID          string       `json:"id" bson:"_id"`
	Name        string       `json:"name" bson:"name"`
	Revision    int64        `json:"revision" bson:"revision"`
	Unit        string       `json:"unit" bson:"unit"`
	Agents      []Agent      `json:"agents" bson:"agents"`
	Experiments []Experiment `json:"experiments" bson:"experiments"`
	UpdatedAt   time.Time    `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

type Agent struct {
	ID       string    `json:"id" bson:"id"`
	Name     string    `json:"name" bson:"name"`
	Space    string    `json:"space" bson:"space"`
	W        float64   `json:"w" bson:"w"`
	H        float64   `json:"h" bson:"h"`
	Elements []Element `json:"elements" bson:"elements"`
	Links    []Link    `json:"links" bson:"links"`
}

type Element struct {
	ID    string         `json:"id" bson:"id"`
	Name  string         `json:"name" bson:"name"`
	Type  string         `json:"type" bson:"type"`
	Lib   string         `json:"lib" bson:"lib"`
	X     float64        `json:"x" bson:"x"`
	Y     float64        `json:"y" bson:"y"`
	W     float64        `json:"w" bson:"w"`
	H     float64        `json:"h" bson:"h"`
	Props map[string]any `json:"props" bson:"props"`
}

type Link struct {
	ID       string `json:"id" bson:"id"`
	From     string `json:"from" bson:"from"`
	FromPort int    `json:"fromPort" bson:"fromPort"`
	To       string `json:"to" bson:"to"`
	ToPort   int    `json:"toPort" bson:"toPort"`
}

type Experiment struct {
	ID    string         `json:"id" bson:"id"`
	Name  string         `json:"name" bson:"name"`
	Type  string         `json:"type" bson:"type"`
	Props map[string]any `json:"props" bson:"props"`
}

type Summary struct {
	ID   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
}

func Validate(p Project) error {
	if p.ID == "" || p.Name == "" || len([]rune(p.Name)) > 200 {
		return &Error{Code: CodeInvalid}
	}
	if p.Unit != "秒" && p.Unit != "分钟" && p.Unit != "小时" {
		return &Error{Code: CodeInvalid}
	}
	raw, err := json.Marshal(p)
	if err != nil || len(raw) > 2<<20 {
		return &Error{Code: CodeInvalid}
	}
	agents := map[string]struct{}{}
	var mainOK, simOK bool
	for _, agent := range p.Agents {
		if agent.ID == "" || agent.Name == "" {
			return &Error{Code: CodeInvalid}
		}
		if _, ok := agents[agent.ID]; ok {
			return &Error{Code: CodeInvalid}
		}
		agents[agent.ID] = struct{}{}
		if agent.ID == "main" {
			mainOK = true
		}
		if agent.Space != "连续" && agent.Space != "离散" && agent.Space != "GIS" {
			return &Error{Code: CodeInvalid}
		}
		if agent.W <= 0 || agent.H <= 0 {
			return &Error{Code: CodeInvalid}
		}
		ids := map[string]struct{}{}
		for _, el := range agent.Elements {
			if el.ID == "" {
				return &Error{Code: CodeInvalid}
			}
			if _, ok := ids[el.ID]; ok {
				return &Error{Code: CodeInvalid}
			}
			ids[el.ID] = struct{}{}
		}
		seenLink := map[string]struct{}{}
		for _, link := range agent.Links {
			if link.ID == "" || link.FromPort < 0 || link.ToPort < 0 || link.From == link.To {
				return &Error{Code: CodeInvalid}
			}
			if _, ok := seenLink[link.ID]; ok {
				return &Error{Code: CodeInvalid}
			}
			seenLink[link.ID] = struct{}{}
			if _, ok := ids[link.From]; !ok {
				return &Error{Code: CodeInvalid}
			}
			if _, ok := ids[link.To]; !ok {
				return &Error{Code: CodeInvalid}
			}
		}
	}
	exps := map[string]struct{}{}
	for _, exp := range p.Experiments {
		if exp.ID == "" || exp.Type == "" {
			return &Error{Code: CodeInvalid}
		}
		if _, ok := exps[exp.ID]; ok {
			return &Error{Code: CodeInvalid}
		}
		exps[exp.ID] = struct{}{}
		if exp.Type == "simulation" {
			simOK = true
		}
	}
	if !mainOK || !simOK {
		return &Error{Code: CodeInvalid}
	}
	return nil
}

func Normalize(p *Project) {
	if p.Agents == nil {
		p.Agents = []Agent{}
	}
	if p.Experiments == nil {
		p.Experiments = []Experiment{}
	}
	for i := range p.Agents {
		if p.Agents[i].Elements == nil {
			p.Agents[i].Elements = []Element{}
		}
		if p.Agents[i].Links == nil {
			p.Agents[i].Links = []Link{}
		}
		for j := range p.Agents[i].Elements {
			if p.Agents[i].Elements[j].Props == nil {
				p.Agents[i].Elements[j].Props = map[string]any{}
			}
		}
	}
	for i := range p.Experiments {
		if p.Experiments[i].Props == nil {
			p.Experiments[i].Props = map[string]any{}
		}
	}
}

func NewID(prefix string) string {
	var buf [8]byte
	_, _ = rand.Read(buf[:])
	return prefix + hex.EncodeToString(buf[:])
}

func AsDoc(err error) *Error {
	var de *Error
	if errors.As(err, &de) {
		return de
	}
	return nil
}

func Unavailable(err error) error {
	if err == nil {
		return nil
	}
	if AsDoc(err) != nil {
		return err
	}
	return fmt.Errorf("%w", err)
}
