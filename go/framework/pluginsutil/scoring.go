package pluginsutil

import "polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"

// Normalizes node scores using the following generic procedure:
// 1. Find the currently highest score and treat it as 100.
// 2. Recalculate all other scores as a percentage (from 0 to 100) of the highest score.
func NormalizeScoresGeneric(scores []pipeline.NodeScore) {
	var maxScore float64 = float64(findMaxScore(scores))
	var highestPossibleScore float64 = float64(pipeline.MaxNodeScore)

	for i := range scores {
		nodeScore := &scores[i]
		percentage := float64(nodeScore.Score) / maxScore
		nodeScore.Score = int64(highestPossibleScore * percentage)
	}
}

func findMaxScore(scores []pipeline.NodeScore) int64 {
	var maxScore int64 = 0
	for i := range scores {
		currScore := scores[i].Score
		if currScore > maxScore {
			maxScore = currScore
		}
	}
	return maxScore
}
