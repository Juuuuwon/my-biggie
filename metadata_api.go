package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// getEC2Metadata retrieves metadata from the EC2 Instance Metadata Service using both v2 and v1.
func getEC2Metadata() (map[string]interface{}, error) {
	metadata := make(map[string]interface{})
	client := &http.Client{Timeout: 2 * time.Second}

	// --- EC2 Metadata v2 ---
	// Get token
	tokenURL := "http://169.254.169.254/latest/api/token"
	reqToken, err := http.NewRequest("PUT", tokenURL, nil)
	if err == nil {
		reqToken.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")
		respToken, err := client.Do(reqToken)
		if err == nil && respToken.StatusCode == http.StatusOK {
			tokenBytes, err := ioutil.ReadAll(respToken.Body)
			respToken.Body.Close()
			if err == nil {
				token := string(tokenBytes)
				// Fetch instance-id using v2
				instanceIDURL := "http://169.254.169.254/latest/meta-data/instance-id"
				reqID, err := http.NewRequest("GET", instanceIDURL, nil)
				if err == nil {
					reqID.Header.Set("X-aws-ec2-metadata-token", token)
					respID, err := client.Do(reqID)
					if err == nil && respID.StatusCode == http.StatusOK {
						idBytes, err := ioutil.ReadAll(respID.Body)
						respID.Body.Close()
						if err == nil {
							metadata["instance_id_v2"] = string(idBytes)
						}
					}
				}
			}
		}
	}

	// --- EC2 Metadata v1 (fallback) ---
	instanceIDURL := "http://169.254.169.254/latest/meta-data/instance-id"
	resp, err := client.Get(instanceIDURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		idBytes, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil {
			metadata["instance_id_v1"] = string(idBytes)
		}
	}
	return metadata, nil
}

// getECSMetadata retrieves metadata from the ECS Metadata Service (v2, for Fargate/EC2).
func getECSMetadata() (map[string]interface{}, error) {
	ecsURL := "http://169.254.170.2/v2/metadata"
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(ecsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

// getEKSMetadata collects metadata from environment variables injected in EKS.
func getEKSMetadata() map[string]interface{} {
	eks := make(map[string]interface{})
	// Typical EKS environment variables (set via downward API or injected)
	keys := []string{"POD_NAME", "POD_NAMESPACE", "POD_IP", "NODE_NAME", "REPLICA_SET"}
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			eks[key] = val
		}
	}
	return eks
}

// extractRevisionFromECS attempts to extract the task definition revision from ECS metadata.
// It looks for a "TaskARN" field and expects the format:
// arn:aws:ecs:region:account-id:task-definition/family:revision
func extractRevisionFromECS(ecsMeta map[string]interface{}) string {
	if taskARN, ok := ecsMeta["TaskARN"].(string); ok {
		// Split the ARN on ":"; the last segment after "task-definition/family" should be the revision.
		parts := strings.Split(taskARN, ":")
		if len(parts) >= 6 {
			// parts[5] should be "task-definition/family" and revision appended; split on "/" then on ":"
			subparts := strings.Split(parts[5], "/")
			if len(subparts) == 2 {
				familyRevision := subparts[1]
				// familyRevision should be in the format family:revision; split by ":"
				frParts := strings.Split(familyRevision, ":")
				if len(frParts) == 2 {
					return frParts[1]
				}
			}
		}
	}
	return ""
}

// extractRevisionFromEKS attempts to extract revision info from EKS metadata.
// It uses the REPLICA_SET environment variable if present; otherwise, it parses the POD_NAME.
func extractRevisionFromEKS(eksMeta map[string]interface{}) string {
	if replica, ok := eksMeta["REPLICA_SET"].(string); ok && replica != "" {
		return replica
	}
	if podName, ok := eksMeta["POD_NAME"].(string); ok && podName != "" {
		// Assume pod name format includes a hash (e.g., "myapp-7f8d4b9b7f")
		parts := strings.Split(podName, "-")
		if len(parts) > 1 {
			return parts[len(parts)-1]
		}
	}
	return ""
}

// hashRevisionToColor converts a revision string into a CSS hex color string.
func hashRevisionToColor(revision string) string {
	var sum int
	for _, ch := range revision {
		sum += int(ch)
	}
	colorValue := sum % 0xFFFFFF
	return fmt.Sprintf("#%06X", colorValue)
}

// MetadataAllHandler handles GET /metadata/all.
// It retrieves metadata from EC2 (v1 and v2), ECS, and EKS environment variables.
func MetadataAllHandler(c *gin.Context) {
	result := make(map[string]interface{})

	// EC2 metadata
	ec2, err := getEC2Metadata()
	if err != nil {
		result["ec2"] = fmt.Sprintf("error: %v", err)
	} else {
		result["ec2"] = ec2
	}

	// ECS metadata
	ecs, err := getECSMetadata()
	if err != nil {
		result["ecs"] = fmt.Sprintf("error: %v", err)
	} else {
		result["ecs"] = ecs
	}

	// EKS metadata from environment variables
	eks := getEKSMetadata()
	if len(eks) == 0 {
		result["eks"] = "not available"
	} else {
		result["eks"] = eks
	}

	ResponseJSON(c, http.StatusOK, result)
}

// RevisionColorHandler handles GET /metadata/revision_color.
// It retrieves revision numbers from ECS and EKS metadata, converts them to a CSS color,
// and returns an HTML page with that background color. If neither revision is available,
// a black background and error message are shown.
func RevisionColorHandler(c *gin.Context) {
	// Retrieve ECS metadata.
	ecsMeta, ecsErr := getECSMetadata()
	// Retrieve EKS metadata.
	eksMeta := getEKSMetadata()

	revisionECS := ""
	if ecsErr == nil {
		revisionECS = extractRevisionFromECS(ecsMeta)
	}
	revisionEKS := extractRevisionFromEKS(eksMeta)

	var combinedRevision string
	if revisionECS != "" && revisionEKS != "" {
		combinedRevision = revisionECS + "-" + revisionEKS
	} else if revisionECS != "" {
		combinedRevision = revisionECS
	} else if revisionEKS != "" {
		combinedRevision = revisionEKS
	}

	var color string
	var message string
	if combinedRevision != "" {
		color = hashRevisionToColor(combinedRevision)
		message = fmt.Sprintf("Revision: %s", combinedRevision)
	} else {
		color = "#000000" // black
		message = "ECS or EKS metadata unavailable"
	}

	// Build HTML response.
	html := fmt.Sprintf(`
		<html>
		<head>
			<title>Revision Color</title>
		</head>
		<body style="background-color:%s;">
			<h1>Revision Color</h1>
			<p>%s</p>
			<p>requested_at: %s</p>
		</body>
		</html>
	`, color, message, time.Now().UTC().Format(time.RFC3339Nano))

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
