package util

import (
	"fmt"
	"math"
	"strconv"

	v1 "k8s.io/api/core/v1"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
)

const (
	cloudNodeLabel = "node-role.kubernetes.io/cloud"
	fogNodeLabel   = "node-role.kubernetes.io/fog"
)

// GetNodeByName gets the specified node from the snapshot obtainable through the handle.
func GetNodeByName(handle framework.Handle, nodeName string) (*framework.NodeInfo, error) {
	nodeInfo, err := handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil || nodeInfo.Node() == nil {
		return nil, fmt.Errorf("Error getting node %s from snapshot: %s", nodeName, err)
	}
	return nodeInfo, nil
}

// Returns the hourly cost of the specified node or 0 if none can be found.
func GetNodeCost(nodeInfo *framework.NodeInfo) float64 {
	costStr, ok := kubeutil.GetLabel(nodeInfo.Node(), kubeutil.LabelNodeCost)
	if !ok {
		return 0
	}
	if cost, err := strconv.ParseFloat(costStr, 64); err != nil {
		return cost
	}
	return 0
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
