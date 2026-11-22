// Package domain 是会话的领域接口。它不引用其他领域，也不保存工程正文。
package domain

import (
	"context"
	"errors"
)

// Turn 是一个会话里的一轮。结果是已经返回给页面的那句话。
type Turn struct {
	User   string `json:"user" bson:"user"`
	Reply  string `json:"reply" bson:"reply"`
	Intent string `json:"intent" bson:"intent"`
	Status string `json:"status" bson:"status"`
}

// Session 是一个工程上的一次对话。它不属于某一个智能体。摘要只覆盖被挤出窗口的更早轮次。
type Session struct {
	ID      string `json:"id" bson:"_id"`
	Project string `json:"project" bson:"project"`
	Title   string `json:"title" bson:"title"`
	Named   bool   `json:"named" bson:"named"`
	Updated int64  `json:"updated" bson:"updated"`
	Turns   []Turn `json:"turns" bson:"turns"`
	Summary string `json:"summary,omitempty" bson:"summary,omitempty"`
	Covered int    `json:"covered,omitempty" bson:"covered,omitempty"`
}

// Card 是列表里的一条会话，不含轮次。
type Card struct {
	ID      string `json:"id"`
	Project string `json:"project"`
	Title   string `json:"title"`
	Named   bool   `json:"named"`
	Updated int64  `json:"updated"`
}

// Fact 是一个工程上跨会话仍成立的一条事实。
type Fact struct {
	Text string `json:"text" bson:"text"`
}

// View 是计划前要读到的记忆。
type View struct {
	Recent  []Turn
	Summary string
	Facts   []string
}

// Commit 是一轮结束后要记下的内容。
type Commit struct {
	Project string
	Session string
	Turn    Turn
}

// Job 是需要压成摘要的更早轮次。没有轮次时不用调用模型。
type Job struct {
	Turns   []Turn
	Prior   string
	Covered int
}

var (
	ErrNotFound = errors.New("notfound")
	ErrInvalid  = errors.New("invalid")
)

// Repository 保存会话和工程级事实。
type Repository interface {
	Load(ctx context.Context, id string) (Session, error)
	Save(ctx context.Context, session Session) error
	List(ctx context.Context, project string) ([]Session, error)
	Delete(ctx context.Context, id string) error
	Facts(ctx context.Context, project string) ([]Fact, error)
	ReplaceFacts(ctx context.Context, project string, facts []Fact) error
}

// Service 是会话的入口。
type Service interface {
	List(ctx context.Context, project string) ([]Card, error)
	Create(ctx context.Context, project string) (Session, error)
	Open(ctx context.Context, id string) (Session, error)
	Rename(ctx context.Context, id, title string) (Session, error)
	Delete(ctx context.Context, id string) error
	Recall(ctx context.Context, project, session string) (View, error)
	Commit(ctx context.Context, in Commit) (Job, error)
	SaveSummary(ctx context.Context, session, summary string, covered int) error
}
