package helper

import (
	"vivek-ray/constants"
	"vivek-ray/models"

	"github.com/gin-gonic/gin"
)

type BulkInsertGraphRequest struct {
	*models.ModelJobs
	Edges []string `json:"edges"`
}

func VerifyDagAndUpdateDegree(dag []*BulkInsertGraphRequest) bool {
	nodeMap := make(map[string]*BulkInsertGraphRequest)
	inDegree := make(map[string]int)

	for _, node := range dag {
		nodeMap[node.UUID] = node
		if _, exists := inDegree[node.UUID]; !exists {
			inDegree[node.UUID] = 0
		}
		for _, target := range node.Edges {
			inDegree[target]++
		}
	}

	for _, node := range dag {
		node.Degree = inDegree[node.UUID]
	}

	queue := make([]string, 0)
	workingDegree := make(map[string]int)
	for uuid, degree := range inDegree {
		workingDegree[uuid] = degree
		if degree == 0 {
			queue = append(queue, uuid)
		}
	}

	processed := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		processed++

		node, exists := nodeMap[current]
		if !exists {
			continue
		}

		for _, target := range node.Edges {
			workingDegree[target]--
			if workingDegree[target] == 0 {
				queue = append(queue, target)
			}
		}
	}

	return processed == len(inDegree)
}

func GetBulkInsertCompleteGraphRequest(c *gin.Context) ([]*BulkInsertGraphRequest, error) {
	var request_nodes []*BulkInsertGraphRequest
	if err := c.ShouldBindJSON(&request_nodes); err != nil {
		return nil, err
	}
	if len(request_nodes) > constants.MaxNodesPerRequest {
		return nil, constants.ErrInvalidDAG
	}

	if !VerifyDagAndUpdateDegree(request_nodes) {
		return nil, constants.ErrInvalidDAG
	}
	return request_nodes, nil
}

// DAG Status Response Types
type JobStatusInfo struct {
	UUID      string `json:"uuid"`
	JobTitle  string `json:"job_title"`
	Status    string `json:"status"`
	Degree    int    `json:"degree"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Progress struct {
	Completed  int     `json:"completed"`
	Failed     int     `json:"failed"`
	Processing int     `json:"processing"`
	Pending    int     `json:"pending"`
	Percentage float64 `json:"percentage"`
}

type DAGStatusResponse struct {
	DAGStatus      string                `json:"dag_status"` // completed, failed, processing, pending, partial
	TotalJobs      int                   `json:"total_jobs"`
	StatusBreakdown map[string]int       `json:"status_breakdown"` // Count by status
	Jobs           map[string]JobStatusInfo `json:"jobs"`              // Individual job statuses
	Progress       Progress              `json:"progress"`
}

type DAGProgressResponse struct {
	Progress        Progress        `json:"progress"`
	StatusBreakdown map[string]int `json:"status_breakdown"`
	DAGStatus       string         `json:"dag_status"`
}

