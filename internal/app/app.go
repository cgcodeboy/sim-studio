// Package app 打开数据库，把各领域的服务交给接入。领域之间不互相注入。
package app

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	behaviordom "github.com/cgcodeboy/sim-studio/internal/pkgs/behavior/domain"
	dynamicsdom "github.com/cgcodeboy/sim-studio/internal/pkgs/dynamics/domain"
	experimentdom "github.com/cgcodeboy/sim-studio/internal/pkgs/experiment/domain"
	fluiddom "github.com/cgcodeboy/sim-studio/internal/pkgs/fluid/domain"
	handlingdom "github.com/cgcodeboy/sim-studio/internal/pkgs/handling/domain"
	librarydom "github.com/cgcodeboy/sim-studio/internal/pkgs/library/domain"
	libraryrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/library/repo"
	librarysvc "github.com/cgcodeboy/sim-studio/internal/pkgs/library/service"
	modeldom "github.com/cgcodeboy/sim-studio/internal/pkgs/model/domain"
	modelrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/model/repo"
	modelsvc "github.com/cgcodeboy/sim-studio/internal/pkgs/model/service"
	pedestriandom "github.com/cgcodeboy/sim-studio/internal/pkgs/pedestrian/domain"
	processdom "github.com/cgcodeboy/sim-studio/internal/pkgs/process/domain"
	raildom "github.com/cgcodeboy/sim-studio/internal/pkgs/rail/domain"
	recorddom "github.com/cgcodeboy/sim-studio/internal/pkgs/record/domain"
	recordrepo "github.com/cgcodeboy/sim-studio/internal/pkgs/record/repo"
	recordsvc "github.com/cgcodeboy/sim-studio/internal/pkgs/record/service"
	roaddom "github.com/cgcodeboy/sim-studio/internal/pkgs/road/domain"
)

// App 把领域服务交给 handler。handler 组合它们，领域不组合彼此。
type App struct {
	Model      modeldom.Service
	Record     recorddom.Service
	Library    librarydom.Service
	Process    processdom.Opener
	Handling   handlingdom.Opener
	Pedestrian pedestriandom.Opener
	Road       roaddom.Opener
	Rail       raildom.Opener
	Fluid      fluiddom.Opener
	Dynamics   dynamicsdom.Opener
	Behavior   behaviordom.Opener
	Experiment experimentdom.Opener
}

// Wire 装上不需要数据库的领域。测试和 Open 共用这一份。
func Wire() *App {
	process, handling, pedestrian, road, rail, fluid, dynamics, behavior, experiment := solvers()
	return &App{
		Library:    librarysvc.New(libraryrepo.New()),
		Process:    process,
		Handling:   handling,
		Pedestrian: pedestrian,
		Road:       road,
		Rail:       rail,
		Fluid:      fluid,
		Dynamics:   dynamics,
		Behavior:   behavior,
		Experiment: experiment,
	}
}

func Open(ctx context.Context, uri string) (*App, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetWriteConcern(writeconcern.Journaled()))
	if err != nil {
		return nil, err
	}
	db := client.Database("simstudio")
	projects := modelrepo.New(db)
	records := recordrepo.New(db)
	if err := records.Ensure(ctx); err != nil {
		return nil, err
	}
	documents := modelsvc.New(projects)
	runner := recordsvc.New(records)
	if err := runner.Sweep(ctx); err != nil {
		return nil, err
	}
	ids, err := documents.IDs(ctx)
	if err != nil {
		return nil, err
	}
	if err := runner.DropOrphans(ctx, ids); err != nil {
		return nil, err
	}
	application := Wire()
	application.Model = documents
	application.Record = runner
	return application, nil
}
