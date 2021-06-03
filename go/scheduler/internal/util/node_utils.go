package util

import (
	"fmt"
	"math"

	v1 "k8s.io/api/core/v1"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	cloudNodeLabel       = "node-role.kubernetes.io/cloud"
	cloudNodeSmallLabel  = "node-role.kubernetes.io/small"
	cloudNodeMediumLabel = "node-role.kubernetes.io/medium"
	cloudNodeLargeLabel  = "node-role.kubernetes.io/large"
	fogNodeLabel         = "node-role.kubernetes.io/fog"
)

// GetNodeByName gets the specified node from the snapshot obtainable through the handle.
func GetNodeByName(handle framework.Handle, nodeName string) (*framework.NodeInfo, error) {
	nodeInfo, err := handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil || nodeInfo.Node() == nil {
		return nil, fmt.Errorf("Error getting node %s from snapshot: %s", nodeName, err)
	}
	return nodeInfo, nil
}

// IsCloudNode returns true if the node is a cloud node, otherwise false.
func IsCloudNode(node *v1.Node) bool {
	_, found := node.Labels[cloudNodeLabel]
	return found
}

// IsFogNode returns true if the node is a fog node, otherwise false.
func IsFogNode(node *v1.Node) bool {
	_, found := node.Labels[fogNodeLabel]
	return found
}

// GetCloudNodeType returns a string representing the type of cloud node (small, medium, large).
func GetCloudNodeType(node *v1.Node) (string, error) {
	if _, found := node.Labels[cloudNodeSmallLabel]; found {
		return "small", nil
	}
	if _, found := node.Labels[cloudNodeMediumLabel]; found {
		return "medium", nil
	}
	if _, found := node.Labels[cloudNodeLargeLabel]; found {
		return "large", nil
	}
	return "", fmt.Errorf("No cloud size label found on node %s", node.Name)
}

// NormalizeNodeScores normalizes the nodeScores to a range between 0 and 100
func NormalizeNodeScores(nodeScores framework.NodeScoreList) {
	length := len(nodeScores)
	if length == 0 {
		return
	}

	maxScore := findMaxScore(nodeScores)
	if maxScore.Score == 0 {
		return
	}
	maxScoreF := float64(maxScore.Score)

	for i, score := range nodeScores {
		if score.Score == maxScore.Score {
			nodeScores[i].Score = 100
			continue
		}
		if score.Score == 0 {
			continue
		}

		fractionOfMax := float64(score.Score) / maxScoreF
		normalizedScore := int64(math.Ceil(fractionOfMax * 100))
		nodeScores[i].Score = normalizedScore
	}
}

func findMaxScore(scores framework.NodeScoreList) framework.NodeScore {
	var maxScore framework.NodeScore = framework.NodeScore{Name: "", Score: math.MinInt64}
	for _, score := range scores {
		if score.Score > maxScore.Score {
			maxScore = score
		}
	}
	return maxScore
}
