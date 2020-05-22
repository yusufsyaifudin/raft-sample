package raft_handler

import (
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
	"net/http"
)

// requestJoin request payload for joining raft cluster
type requestJoin struct {
	NodeID      string `json:"node_id"`
	RaftAddress string `json:"raft_address"`
}

// JoinRaftHandler handling join raft
func (h handler) JoinRaftHandler(eCtx echo.Context) error {
	var form = requestJoin{}
	if err := eCtx.Bind(&form); err != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error binding: %s", err.Error()),
		})
	}

	var (
		nodeID   = form.NodeID
		raftAddr = form.RaftAddress
	)

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

	// This must be run on the leader or it will fail.
	f := h.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(raftAddr), 0, 0)
	if f.Error() != nil {
		return eCtx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error add voter: %s", f.Error().Error()),
		})
	}

	return eCtx.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("node %s at %s joined successfully", nodeID, raftAddr),
		"data":    h.raft.Stats(),
	})
}
