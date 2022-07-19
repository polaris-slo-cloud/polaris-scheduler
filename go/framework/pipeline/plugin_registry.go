package pipeline

// Combines a ScorePlugin with its ScoreExtensions.
type ScorePluginWithExtensions struct {

	// The actual score plugin instance.
	ScorePlugin

	// The ScoreExtensions supplied by the ScorePlugin or nil, if the plugin does not have any.
	ScoreExtensions
}

// Serves as an ordered list of all plugin instances for the scheduling pipeline.
type PluginRegistry struct {

	// The configured SortPlugin used for determining the scheduling order of incoming pods.
	Sort SortPlugin

	SampleNodes []SampleNodesPlugin

	PreFilter []PreFilterPlugin

	Filter []FilterPlugin

	PreScore []PreScorePlugin

	Score []ScorePluginWithExtensions

	Reserve []ReservePlugin
}
