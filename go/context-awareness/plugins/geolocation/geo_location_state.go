package geolocation

import (
	geo "github.com/kellydunn/golang-geo"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.StateData = (*geoLocationState)(nil)
)

const (
	stateKey = PluginName + ".state"
)

type geoLocationState struct {
	podTargetLocation *geo.Point
	maxDistanceKm     float64
}
