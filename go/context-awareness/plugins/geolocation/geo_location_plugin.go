package geolocation

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	geo "github.com/kellydunn/golang-geo"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

const (
	PluginName = "GeoLocation"

	NodeGeoLocationLabel                = "polaris-slo-cloud.github.io/geo-location"
	PodTargetGeoLocationLabel           = "polaris-slo-cloud.github.io/geo-location.target-location"
	PodMaxDistanceToTargetLocationLabel = "polaris-slo-cloud.github.io/geo-location.max-distance-km"

	DefaultMaxDistanceToTargetLocationKm = 10.0

	// If the distance between a node and a pod's targe location is equal to max distance the score should be 1,
	// if the distance is 0km, then the score should be the max score. Thus, at max distance we need to subtract MaxNodeScore - 1.
	distanceScoreSubtractionRange = float64(pipeline.MaxNodeScore - 1)
)

var (
	_ pipeline.PreFilterPlugin               = (*GeoLocationPlugin)(nil)
	_ pipeline.FilterPlugin                  = (*GeoLocationPlugin)(nil)
	_ pipeline.ScorePlugin                   = (*GeoLocationPlugin)(nil)
	_ pipeline.SchedulingPluginFactoryFunc   = NewGeoLocationSchedulingPlugin
	_ pipeline.ClusterAgentPluginFactoryFunc = NewGeoLocationClusterAgentPlugin
)

// This GeoLocationPlugin ensures that a pod is placed in or close to its specified target geo-location.
// The plugin has two main functions:
//   - Filter: filter out nodes that are further away from the pod's desired target location than the max distance specified by the pod.
//   - Score: assign a score, based on how far (as a percentage of the max distance) the node is away from the target location
type GeoLocationPlugin struct {
}

func NewGeoLocationSchedulingPlugin(configMap config.PluginConfig, scheduler pipeline.PolarisScheduler) (pipeline.Plugin, error) {
	return &GeoLocationPlugin{}, nil
}

func NewGeoLocationClusterAgentPlugin(configMap config.PluginConfig, clusterAgentServices pipeline.ClusterAgentServices) (pipeline.Plugin, error) {
	return &GeoLocationPlugin{}, nil
}

func (glp *GeoLocationPlugin) Name() string {
	return PluginName
}

func (glp *GeoLocationPlugin) PreFilter(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo) pipeline.Status {
	targetLocationStr, ok := podInfo.Pod.Labels[PodTargetGeoLocationLabel]
	if !ok {
		// No target location specified, so this pod required no further processing by this plugin.
		return pipeline.NewSuccessStatus()
	}
	targetLocation, err := parseGeoLocation(targetLocationStr)
	if err != nil {
		return pipeline.NewStatus(pipeline.Unschedulable, err.Error())
	}

	state := &geoLocationState{
		podTargetLocation: targetLocation,
		maxDistanceKm:     DefaultMaxDistanceToTargetLocationKm,
	}

	maxDistanceToTargetStr, ok := podInfo.Pod.Labels[PodMaxDistanceToTargetLocationLabel]
	if ok {
		maxDistance, err := strconv.ParseFloat(maxDistanceToTargetStr, 64)
		if err != nil {
			return pipeline.NewStatus(pipeline.Unschedulable, fmt.Sprintf("could not parse %s", PodMaxDistanceToTargetLocationLabel))
		}
		state.maxDistanceKm = maxDistance
	}

	ctx.Write(stateKey, state)
	return pipeline.NewSuccessStatus()
}

func (glp *GeoLocationPlugin) Filter(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, nodeInfo *pipeline.NodeInfo) pipeline.Status {
	state, ok := glp.readState(ctx)
	if !ok {
		// No state for this pod, thus, no target location was specified, so this pod required no further processing by this plugin.
		return pipeline.NewSuccessStatus()
	}

	nodeLocation := readNodeGeoLocation(nodeInfo)
	if nodeLocation == nil {
		return pipeline.NewStatus(pipeline.Unschedulable, "node does not have a (valid) geo location", nodeInfo.Node.Name)
	}

	distance := state.podTargetLocation.GreatCircleDistance(nodeLocation)
	if distance <= state.maxDistanceKm {
		return pipeline.NewSuccessStatus()
	} else {
		return pipeline.NewStatus(pipeline.Unschedulable, "node location is too far from the pod's target location", nodeInfo.Node.Name)
	}
}

func (glp *GeoLocationPlugin) Score(ctx pipeline.SchedulingContext, podInfo *pipeline.PodInfo, nodeInfo *pipeline.NodeInfo) (int64, pipeline.Status) {
	state, ok := glp.readState(ctx)
	if !ok {
		// No state for this pod, thus, no target location was specified, so this pod required no further processing by this plugin.
		return pipeline.MaxNodeScore, pipeline.NewSuccessStatus()
	}

	nodeLocation := readNodeGeoLocation(nodeInfo)
	if nodeLocation == nil {
		// If this happens, an error occurred while updating the location of the node, because getting to the Score method means that
		// the node is sufficiently close to the desired location of the pod.
		// But we nevertheless return a success status, otherwise scheduling will fail.
		return 0, pipeline.NewSuccessStatus()
	}

	distance := state.podTargetLocation.GreatCircleDistance(nodeLocation)
	distancePercentage := distance / state.maxDistanceKm
	score := float64(pipeline.MaxNodeScore) - distanceScoreSubtractionRange*distancePercentage
	return int64(math.Ceil(score)), pipeline.NewSuccessStatus()
}

func (glp *GeoLocationPlugin) ScoreExtensions() pipeline.ScoreExtensions {
	return nil
}

func (glp *GeoLocationPlugin) readState(ctx pipeline.SchedulingContext) (*geoLocationState, bool) {
	state, ok := ctx.Read(stateKey)
	if !ok {
		return nil, false
	}
	resState, ok := state.(*geoLocationState)
	if !ok {
		panic(fmt.Sprintf("invalid object stored as %s", stateKey))
	}
	return resState, true
}

// Parses a geo location string (e.g, "48.22066363087445_16.403747854930955") into a Point object.
// The coordinates must be separated by an underscore (spaces and commas are not permitted in Kubernetes label values).
func parseGeoLocation(locationStr string) (*geo.Point, error) {
	latLong := strings.Split(locationStr, "_")
	if len(latLong) != 2 {
		return nil, fmt.Errorf("the string %s is not a valid geo location string", locationStr)
	}

	lat, err := strconv.ParseFloat(latLong[0], 64)
	if err != nil {
		return nil, fmt.Errorf("the latitude part of %s is not a valid floating point number", locationStr)
	}
	long, err := strconv.ParseFloat(latLong[1], 64)
	if err != nil {
		return nil, fmt.Errorf("the longitude part of %s is not a valid floating point number", locationStr)
	}

	return geo.NewPoint(lat, long), nil
}

func readNodeGeoLocation(nodeInfo *pipeline.NodeInfo) *geo.Point {
	nodeLocationStr, ok := nodeInfo.Node.Labels[NodeGeoLocationLabel]
	if !ok {
		return nil
	}
	nodeLocation, err := parseGeoLocation(nodeLocationStr)
	if err != nil {
		return nil
	}
	return nodeLocation
}
