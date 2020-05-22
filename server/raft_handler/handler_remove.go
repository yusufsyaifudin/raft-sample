package raft_handler

import (
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
	"net/http"
)

// requestRemove request payload for removing node from raft cluster
type requestRemove struct {
	NodeID string `json:"node_id"`
}

// RemoveRaftHandler handling removing raft
func (h handler) RemoveRaftHandler(eCtx echo.Context) error {
	var form = requestRemove{}
	if err := eCtx.Bind(&form); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error binding: %s", err.Error()),
		})
	}

	var nodeID = form.NodeID

	if h.raft.State() != raft.Leader {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "not the leader",
		})
	}

	configFuture := h.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("failed to get raft configuration: %s", err.Error()),
		})
	}

	future := h.raft.RemoveServer(raft.ServerID(nodeID), 0, 0)
	if err := future.Error(); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error removing existing node %s: %s", nodeID, err.Error()),
		})
	}

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("node %s removed successfully", nodeID),
		"data":    h.raft.Stats(),
	})
}
