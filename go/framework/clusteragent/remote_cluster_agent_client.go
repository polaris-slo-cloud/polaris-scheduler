package clusteragent

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

var (
	_ client.ClusterClient = (*RemoteClusterAgentClient)(nil)
)

// ClusterClient implementation that connects via REST to a remote PolarisClusterAgent.
type RemoteClusterAgentClient struct {
	clusterName   string
	clusterConfig *config.RemoteClusterConfig
	httpClient    *http.Client
	logger        *logr.Logger

	commitSchedulingDecisionURI string
}

func NewRemoteClusterAgentClient(clusterName string, clusterConfig *config.RemoteClusterConfig, logger *logr.Logger) *RemoteClusterAgentClient {
	var err error

	cbc := &RemoteClusterAgentClient{
		clusterName:   clusterName,
		clusterConfig: clusterConfig,
		httpClient:    &http.Client{},
		logger:        logger,
	}

	cbc.commitSchedulingDecisionURI, err = url.JoinPath(clusterConfig.BaseURI, ClusterAgentEndpointsPrefix, CommitSchedulingDecisionEndpoint)
	if err != nil {
		panic(err)
	}

	return cbc
}

func (cbc *RemoteClusterAgentClient) ClusterName() string {
	return cbc.clusterName
}

func (cbc *RemoteClusterAgentClient) CommitSchedulingDecision(ctx context.Context, schedulingDecision *client.ClusterSchedulingDecision) error {
	httpReq, err := cbc.createPostRequest(ctx, cbc.commitSchedulingDecisionURI, schedulingDecision)
	if err != nil {
		return err
	}

	httpResp, err := cbc.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	if httpResp.Body != nil {
		defer httpResp.Body.Close()
	}

	if httpResp.StatusCode == http.StatusCreated {
		return nil
	} else {
		if agentError, err := parseErrorResponseBody(httpResp); err == nil {
			return agentError.Error
		} else {
			return err
		}
	}
}

func (cbc *RemoteClusterAgentClient) createPostRequest(ctx context.Context, requestURI string, bodyObj any) (*http.Request, error) {
	body, err := json.Marshal(bodyObj)
	if err != nil {
		return nil, err
	}
	bodyBuffer := bytes.NewBuffer(body)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", requestURI, bodyBuffer)
	if err != nil {
		return nil, err
	}

	httpReq.Header["Content-Type"] = []string{"application/json"}
	httpReq.Header["Accept"] = []string{"application/json"}

	return httpReq, nil
}

func parseErrorResponseBody(httpResp *http.Response) (*PolarisClusterAgentError, error) {
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	agentError := PolarisClusterAgentError{}
	if err := json.Unmarshal(body, &agentError); err != nil {
		return nil, err
	}

	return &agentError, nil
}
