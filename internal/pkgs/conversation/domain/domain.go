// Package domain 是会话的领域接口。它不引用其他领域，也不保存工程正文。
package domain

import "context"

// Turn 是一个会话里的一轮。结果是已经返回给页面的那句话。
type Turn struct {
	User   string `json:"user" bson:"user"`
	Reply  string `json:"reply" bson:"reply"`
	Intent string `json:"intent" bson:"intent"`
	Status string `json:"status" bson:"status"`
}

// Session 是一次对话。摘要只覆盖被挤出窗口的更早轮次。
type Session struct {
	ID      string `json:"id" bson:"_id"`
	Project string `json:"project" bson:"project"`
	Turns   []Turn `json:"turns" bson:"turns"`
	Summary string `json:"summary" bson:"summary"`
	Covered int    `json:"covered" bson:"covered"`
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

// Repository 保存会话和工程级事实。
type Repository interface {
	Load(ctx context.Context, id string) (Session, error)
	Save(ctx context.Context, session Session) error
	Facts(ctx context.Context, project string) ([]Fact, error)
	ReplaceFacts(ctx context.Context, project string, facts []Fact) error
}

// Service 是会话的入口。
type Service interface {
	Recall(ctx context.Context, project, session string) (View, error)
	Commit(ctx context.Context, in Commit) (Job, error)
	SaveSummary(ctx context.Context, session, summary string, covered int) error
}
