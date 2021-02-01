package engine

import (
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/publishr/pkg/config"
)

type engineBase struct {
	Config       config.Interface
	PipelineData *pipeline.Data
}
